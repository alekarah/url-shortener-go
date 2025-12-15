package handlers

import (
	"context"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"url-short/internal/service"
)

// RedirectHandler обработчик для редиректа по короткому коду
type RedirectHandler struct {
	urlService       service.URLService
	analyticsService service.AnalyticsService
}

// NewRedirectHandler создает новый redirect handler
func NewRedirectHandler(
	urlService service.URLService,
	analyticsService service.AnalyticsService,
) *RedirectHandler {
	return &RedirectHandler{
		urlService:       urlService,
		analyticsService: analyticsService,
	}
}

// Redirect выполняет редирект на оригинальный URL
// GET /{shortCode}
func (h *RedirectHandler) Redirect(w http.ResponseWriter, r *http.Request) {
	shortCode := chi.URLParam(r, "shortCode")

	if shortCode == "" {
		http.Error(w, "Короткий код не указан", http.StatusBadRequest)
		return
	}

	// Получаем полный объект URL (с ID)
	url, err := h.urlService.GetURLByShortCode(r.Context(), shortCode)
	if err != nil {
		if strings.Contains(err.Error(), "не найден") {
			http.Error(w, "URL не найден", http.StatusNotFound)
			return
		}
		if strings.Contains(err.Error(), "истекла") {
			http.Error(w, "Ссылка истекла", http.StatusGone)
			return
		}
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	// Записываем аналитику (в фоне)
	go func() {
		ctx := context.Background() // Используем новый контекст, так как запрос может завершиться
		ipAddress := getIPAddress(r)
		userAgent := r.UserAgent()
		referer := r.Referer()

		// Записываем клик в аналитику (счетчик инкрементируется внутри RecordClick)
		if err := h.analyticsService.RecordClick(ctx, url.ID, ipAddress, userAgent, referer); err != nil {
			// Логируем ошибку, но не останавливаем редирект
			println("Ошибка записи аналитики:", err.Error())
		}
	}()

	// Выполняем редирект (302 вместо 301 чтобы избежать кеширования браузером)
	http.Redirect(w, r, url.OriginalURL, http.StatusFound)
}

// getIPAddress извлекает IP адрес из запроса
func getIPAddress(r *http.Request) string {
	// Проверяем заголовки прокси
	ip := r.Header.Get("X-Real-IP")
	if ip == "" {
		ip = r.Header.Get("X-Forwarded-For")
		if ip != "" {
			// X-Forwarded-For может содержать несколько IP через запятую
			ips := strings.Split(ip, ",")
			ip = strings.TrimSpace(ips[0])
		}
	}

	// Если заголовки пусты, берем RemoteAddr
	if ip == "" {
		ip = r.RemoteAddr
		// RemoteAddr включает порт, убираем его
		if idx := strings.LastIndex(ip, ":"); idx != -1 {
			ip = ip[:idx]
		}
	}

	return ip
}
