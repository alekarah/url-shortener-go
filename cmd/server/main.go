package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"url-short/internal/config"
	"url-short/internal/database"
	"url-short/internal/handlers"
	"url-short/internal/middleware"
	"url-short/internal/repository"
	"url-short/internal/service"
	"url-short/pkg/shortener"
)

func main() {
	// Загружаем конфигурацию
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	// Подключаемся к PostgreSQL
	db, err := database.NewPostgresDB(database.PostgresConfig{
		DSN: cfg.Postgres.GetDSN(),
	})
	if err != nil {
		log.Fatalf("Ошибка подключения к PostgreSQL: %v", err)
	}
	defer func() {
		if err := database.CloseDB(db); err != nil {
			log.Printf("Ошибка закрытия PostgreSQL: %v", err)
		}
	}()
	log.Println("✓ Подключено к PostgreSQL")

	// Подключаемся к Redis
	redisClient, err := database.NewRedisClient(database.RedisConfig{
		Address:  cfg.Redis.GetAddress(),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	if err != nil {
		log.Fatalf("Ошибка подключения к Redis: %v", err)
	}
	defer func() {
		if err := database.CloseRedis(redisClient); err != nil {
			log.Printf("Ошибка закрытия Redis: %v", err)
		}
	}()
	log.Println("✓ Подключено к Redis")

	// Инициализируем repositories
	urlRepo := repository.NewURLRepository(db)
	analyticsRepo := repository.NewAnalyticsRepository(db)

	// Инициализируем services
	generator := shortener.NewGenerator()
	urlService := service.NewURLService(urlRepo, generator, redisClient, cfg.App.BaseURL, cfg.App.CacheTTL)
	analyticsService := service.NewAnalyticsService(analyticsRepo, urlRepo)

	// Инициализируем handlers
	urlHandler := handlers.NewURLHandler(urlService)
	analyticsHandler := handlers.NewAnalyticsHandler(analyticsService)
	redirectHandler := handlers.NewRedirectHandler(urlService, analyticsService)

	// Настраиваем роутер
	r := chi.NewRouter()

	// Middleware
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.Timeout(60 * time.Second))

	// Health check endpoint для Render
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("OK")); err != nil {
			log.Printf("Ошибка записи health check: %v", err)
		}
	})

	// Статические файлы
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/index.html")
	})
	r.Get("/links", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/links.html")
	})
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// API routes
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/urls", urlHandler.GetAllURLs)
		r.Post("/urls", urlHandler.CreateShortURL)
		r.Get("/urls/{id}", urlHandler.GetURL)
		r.Delete("/urls/{id}", urlHandler.DeleteURL)
		r.Get("/urls/{id}/stats", analyticsHandler.GetURLStats)
	})

	// Redirect route - внутри /api/ namespace для обхода ограничений Render
	r.Get("/api/r", redirectHandler.Redirect)

	// Настраиваем сервер
	server := &http.Server{
		Addr:         cfg.Server.GetServerAddress(),
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Запускаем сервер в горутине
	go func() {
		log.Printf("🚀 Сервер запущен на %s", cfg.Server.GetServerAddress())
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Ошибка запуска сервера: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Остановка сервера...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Ошибка остановки сервера: %v", err)
	}

	log.Println("✓ Сервер остановлен")
}
