package handlers

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"url-short/internal/service"
)

// AnalyticsHandler обработчик для аналитики
type AnalyticsHandler struct {
	analyticsService service.AnalyticsService
}

// NewAnalyticsHandler создает новый analytics handler
func NewAnalyticsHandler(analyticsService service.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{
		analyticsService: analyticsService,
	}
}

// GetURLStats получает статистику по URL
// GET /api/v1/urls/{id}/stats
func (h *AnalyticsHandler) GetURLStats(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Невалидный ID")
		return
	}

	stats, err := h.analyticsService.GetURLStats(r.Context(), id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, stats)
}
