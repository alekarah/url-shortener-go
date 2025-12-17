package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"url-short/internal/models"
	"url-short/internal/service"
)

// URLHandler обработчик для URL endpoints
type URLHandler struct {
	urlService service.URLService
}

// NewURLHandler создает новый URL handler
func NewURLHandler(urlService service.URLService) *URLHandler {
	return &URLHandler{
		urlService: urlService,
	}
}

// CreateShortURL создает короткую ссылку
// POST /api/v1/urls
func (h *URLHandler) CreateShortURL(w http.ResponseWriter, r *http.Request) {
	var req models.CreateURLRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Невалидный JSON")
		return
	}

	// Базовая валидация
	if req.OriginalURL == "" {
		respondWithError(w, http.StatusBadRequest, "URL обязателен")
		return
	}

	response, err := h.urlService.CreateShortURL(r.Context(), &req)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, response)
}

// GetURL получает информацию о URL по ID
// GET /api/v1/urls/{id}
func (h *URLHandler) GetURL(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Невалидный ID")
		return
	}

	response, err := h.urlService.GetURLByID(r.Context(), id)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "URL не найден")
		return
	}

	respondWithJSON(w, http.StatusOK, response)
}

// DeleteURL удаляет URL
// DELETE /api/v1/urls/{id}
func (h *URLHandler) DeleteURL(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Невалидный ID")
		return
	}

	if err := h.urlService.DeleteURL(r.Context(), id); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{
		"message": "URL успешно удален",
	})
}

// GetAllURLs получает список всех URL
// GET /api/v1/urls
func (h *URLHandler) GetAllURLs(w http.ResponseWriter, r *http.Request) {
	// Получаем параметры пагинации из query string
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 50 // по умолчанию
	offset := 0

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			offset = o
		}
	}

	urls, err := h.urlService.GetAllURLs(r.Context(), limit, offset)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Ошибка получения списка URL")
		return
	}

	respondWithJSON(w, http.StatusOK, urls)
}

// respondWithJSON отправляет JSON ответ
func respondWithJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// respondWithError отправляет ошибку в JSON формате
func respondWithError(w http.ResponseWriter, statusCode int, message string) {
	respondWithJSON(w, statusCode, map[string]string{
		"error": message,
	})
}
