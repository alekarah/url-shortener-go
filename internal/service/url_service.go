package service

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"

	"url-short/internal/models"
	"url-short/internal/repository"
	"url-short/pkg/shortener"
)

// URLService интерфейс для бизнес-логики работы с URL
type URLService interface {
	CreateShortURL(ctx context.Context, req *models.CreateURLRequest) (*models.URLResponse, error)
	GetOriginalURL(ctx context.Context, shortCode string) (string, error)
	GetURLByShortCode(ctx context.Context, shortCode string) (*models.URL, error)
	GetURLByID(ctx context.Context, id int64) (*models.URLResponse, error)
	GetAllURLs(ctx context.Context, limit, offset int) ([]*models.URLResponse, error)
	DeleteURL(ctx context.Context, id int64) error
	IncrementClicks(ctx context.Context, id int64) error
}

// urlService имплементация URLService
type urlService struct {
	urlRepo   repository.URLRepository
	generator shortener.Generator
	redis     *redis.Client
	baseURL   string
	cacheTTL  time.Duration
}

// NewURLService создает новый URL service
func NewURLService(
	urlRepo repository.URLRepository,
	generator shortener.Generator,
	redis *redis.Client,
	baseURL string,
	cacheTTL int,
) URLService {
	return &urlService{
		urlRepo:   urlRepo,
		generator: generator,
		redis:     redis,
		baseURL:   baseURL,
		cacheTTL:  time.Duration(cacheTTL) * time.Second,
	}
}

// CreateShortURL создает короткую ссылку
func (s *urlService) CreateShortURL(ctx context.Context, req *models.CreateURLRequest) (*models.URLResponse, error) {
	var shortCode string
	var err error

	// Если пользователь предоставил свой код
	if req.CustomCode != "" {
		// Проверяем валидность кода
		if !shortener.IsValidShortCode(req.CustomCode) {
			return nil, fmt.Errorf("невалидный короткий код")
		}

		// Проверяем что код еще не занят
		exists, err := s.urlRepo.ShortCodeExists(ctx, req.CustomCode)
		if err != nil {
			return nil, fmt.Errorf("ошибка проверки кода: %w", err)
		}
		if exists {
			return nil, fmt.Errorf("короткий код уже занят")
		}

		shortCode = req.CustomCode
	} else {
		// Генерируем короткий код
		maxAttempts := 5
		for i := 0; i < maxAttempts; i++ {
			shortCode, err = s.generator.Generate(7)
			if err != nil {
				return nil, fmt.Errorf("ошибка генерации кода: %w", err)
			}

			// Проверяем уникальность
			exists, err := s.urlRepo.ShortCodeExists(ctx, shortCode)
			if err != nil {
				return nil, fmt.Errorf("ошибка проверки кода: %w", err)
			}

			if !exists {
				break
			}

			// Если это последняя попытка и код все еще занят
			if i == maxAttempts-1 {
				return nil, fmt.Errorf("не удалось сгенерировать уникальный код")
			}
		}
	}

	// Создаем URL в БД
	url := &models.URL{
		ShortCode:   shortCode,
		OriginalURL: req.OriginalURL,
	}

	if req.ExpiresAt != nil {
		url.ExpiresAt.Time = *req.ExpiresAt
		url.ExpiresAt.Valid = true
	}

	if err := s.urlRepo.Create(ctx, url); err != nil {
		return nil, fmt.Errorf("ошибка создания URL: %w", err)
	}

	// Кешируем в Redis (игнорируем ошибку кеширования, основные данные уже в БД)
	if s.redis != nil {
		cacheKey := fmt.Sprintf("url:%s", shortCode)
		s.redis.Set(ctx, cacheKey, url.OriginalURL, s.cacheTTL) // nolint:errcheck
	}

	// Формируем ответ
	response := &models.URLResponse{
		ID:          url.ID,
		ShortCode:   url.ShortCode,
		ShortURL:    fmt.Sprintf("%s/api/r?code=%s", s.baseURL, url.ShortCode),
		OriginalURL: url.OriginalURL,
		CreatedAt:   url.CreatedAt,
		ClicksCount: url.ClicksCount,
	}

	if url.ExpiresAt.Valid {
		response.ExpiresAt = &url.ExpiresAt.Time
	}

	return response, nil
}

// GetOriginalURL получает оригинальный URL по короткому коду
func (s *urlService) GetOriginalURL(ctx context.Context, shortCode string) (string, error) {
	// Проверяем кеш
	if s.redis != nil {
		cacheKey := fmt.Sprintf("url:%s", shortCode)
		cachedURL, err := s.redis.Get(ctx, cacheKey).Result()
		if err == nil {
			return cachedURL, nil
		}
	}

	// Получаем из БД
	url, err := s.urlRepo.GetByShortCode(ctx, shortCode)
	if err != nil {
		return "", err
	}

	// Проверяем не истекла ли ссылка
	if url.ExpiresAt.Valid && url.ExpiresAt.Time.Before(time.Now()) {
		return "", fmt.Errorf("ссылка истекла")
	}

	// Кешируем в Redis (игнорируем ошибку кеширования)
	if s.redis != nil {
		cacheKey := fmt.Sprintf("url:%s", shortCode)
		s.redis.Set(ctx, cacheKey, url.OriginalURL, s.cacheTTL) // nolint:errcheck
	}

	return url.OriginalURL, nil
}

// GetURLByID получает информацию о URL по ID
func (s *urlService) GetURLByID(ctx context.Context, id int64) (*models.URLResponse, error) {
	url, err := s.urlRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	response := &models.URLResponse{
		ID:          url.ID,
		ShortCode:   url.ShortCode,
		ShortURL:    fmt.Sprintf("%s/api/r?code=%s", s.baseURL, url.ShortCode),
		OriginalURL: url.OriginalURL,
		CreatedAt:   url.CreatedAt,
		ClicksCount: url.ClicksCount,
	}

	if url.ExpiresAt.Valid {
		response.ExpiresAt = &url.ExpiresAt.Time
	}

	return response, nil
}

// GetAllURLs получает список всех URL с пагинацией
func (s *urlService) GetAllURLs(ctx context.Context, limit, offset int) ([]*models.URLResponse, error) {
	// Устанавливаем разумные лимиты
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	urls, err := s.urlRepo.GetAll(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	responses := make([]*models.URLResponse, 0, len(urls))
	for _, url := range urls {
		response := &models.URLResponse{
			ID:          url.ID,
			ShortCode:   url.ShortCode,
			ShortURL:    fmt.Sprintf("%s/api/r?code=%s", s.baseURL, url.ShortCode),
			OriginalURL: url.OriginalURL,
			CreatedAt:   url.CreatedAt,
			ClicksCount: url.ClicksCount,
		}

		if url.ExpiresAt.Valid {
			response.ExpiresAt = &url.ExpiresAt.Time
		}

		responses = append(responses, response)
	}

	return responses, nil
}

// GetURLByShortCode получает полный объект URL по короткому коду
func (s *urlService) GetURLByShortCode(ctx context.Context, shortCode string) (*models.URL, error) {
	url, err := s.urlRepo.GetByShortCode(ctx, shortCode)
	if err != nil {
		return nil, err
	}

	// Проверяем не истекла ли ссылка
	if url.ExpiresAt.Valid && url.ExpiresAt.Time.Before(time.Now()) {
		return nil, fmt.Errorf("ссылка истекла")
	}

	return url, nil
}

// IncrementClicks увеличивает счетчик кликов
func (s *urlService) IncrementClicks(ctx context.Context, id int64) error {
	return s.urlRepo.IncrementClicks(ctx, id)
}

// DeleteURL удаляет URL
func (s *urlService) DeleteURL(ctx context.Context, id int64) error {
	// Получаем URL чтобы узнать short_code
	url, err := s.urlRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Удаляем из БД
	if err := s.urlRepo.Delete(ctx, id); err != nil {
		return err
	}

	// Удаляем из кеша (игнорируем ошибку)
	if s.redis != nil {
		cacheKey := fmt.Sprintf("url:%s", url.ShortCode)
		s.redis.Del(ctx, cacheKey) // nolint:errcheck
	}

	return nil
}
