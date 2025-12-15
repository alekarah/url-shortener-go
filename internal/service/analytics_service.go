package service

import (
	"context"

	"url-short/internal/models"
	"url-short/internal/repository"
)

// AnalyticsService интерфейс для бизнес-логики аналитики
type AnalyticsService interface {
	RecordClick(ctx context.Context, urlID int64, ipAddress, userAgent, referer string) error
	GetURLStats(ctx context.Context, urlID int64) (*models.URLStats, error)
}

// analyticsService имплементация AnalyticsService
type analyticsService struct {
	analyticsRepo repository.AnalyticsRepository
	urlRepo       repository.URLRepository
}

// NewAnalyticsService создает новый Analytics service
func NewAnalyticsService(
	analyticsRepo repository.AnalyticsRepository,
	urlRepo repository.URLRepository,
) AnalyticsService {
	return &analyticsService{
		analyticsRepo: analyticsRepo,
		urlRepo:       urlRepo,
	}
}

// RecordClick записывает клик и увеличивает счетчик
func (s *analyticsService) RecordClick(ctx context.Context, urlID int64, ipAddress, userAgent, referer string) error {
	// Записываем клик в аналитику
	event := &models.ClickEvent{
		URLID:     urlID,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Referer:   referer,
		// Country и City можно определить через GeoIP сервис
		// Для простоты пока оставим пустыми
	}

	if err := s.analyticsRepo.RecordClick(ctx, event); err != nil {
		return err
	}

	// Увеличиваем счетчик кликов в таблице URLs
	if err := s.urlRepo.IncrementClicks(ctx, urlID); err != nil {
		return err
	}

	return nil
}

// GetURLStats получает статистику по URL
func (s *analyticsService) GetURLStats(ctx context.Context, urlID int64) (*models.URLStats, error) {
	// Получаем статистику (последние 100 кликов)
	stats, err := s.analyticsRepo.GetStatsByURL(ctx, urlID, 100)
	if err != nil {
		return nil, err
	}

	return stats, nil
}
