package repository

import (
	"context"
	"database/sql"
	"fmt"

	"url-short/internal/models"
)

// AnalyticsRepository интерфейс для работы с аналитикой
type AnalyticsRepository interface {
	RecordClick(ctx context.Context, event *models.ClickEvent) error
	GetStatsByURL(ctx context.Context, urlID int64, limit int) (*models.URLStats, error)
}

// analyticsRepository имплементация AnalyticsRepository
type analyticsRepository struct {
	db *sql.DB
}

// NewAnalyticsRepository создает новый Analytics repository
func NewAnalyticsRepository(db *sql.DB) AnalyticsRepository {
	return &analyticsRepository{db: db}
}

// RecordClick записывает клик в аналитику
func (r *analyticsRepository) RecordClick(ctx context.Context, event *models.ClickEvent) error {
	query := `
		INSERT INTO analytics (url_id, ip_address, user_agent, referer, country, city)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		event.URLID,
		nullString(event.IPAddress),
		nullString(event.UserAgent),
		nullString(event.Referer),
		nullString(event.Country),
		nullString(event.City),
	)

	if err != nil {
		return fmt.Errorf("ошибка записи клика: %w", err)
	}

	return nil
}

// GetStatsByURL получает статистику по URL
func (r *analyticsRepository) GetStatsByURL(ctx context.Context, urlID int64, limit int) (*models.URLStats, error) {
	stats := &models.URLStats{}

	// Получаем общее количество кликов
	if err := r.getTotalClicks(ctx, urlID, stats); err != nil {
		return nil, err
	}

	// Получаем количество уникальных IP
	if err := r.getUniqueIPs(ctx, urlID, stats); err != nil {
		return nil, err
	}

	// Получаем клики по датам
	if err := r.getClicksByDate(ctx, urlID, stats); err != nil {
		return nil, err
	}

	// Получаем клики по странам
	if err := r.getClicksByCountry(ctx, urlID, stats); err != nil {
		return nil, err
	}

	// Получаем последние клики
	if err := r.getRecentClicks(ctx, urlID, limit, stats); err != nil {
		return nil, err
	}

	return stats, nil
}

// getTotalClicks получает общее количество кликов
func (r *analyticsRepository) getTotalClicks(ctx context.Context, urlID int64, stats *models.URLStats) error {
	query := `SELECT COUNT(*) FROM analytics WHERE url_id = $1`
	err := r.db.QueryRowContext(ctx, query, urlID).Scan(&stats.TotalClicks)
	if err != nil {
		return fmt.Errorf("ошибка получения общего количества кликов: %w", err)
	}
	return nil
}

// getUniqueIPs получает количество уникальных IP адресов
func (r *analyticsRepository) getUniqueIPs(ctx context.Context, urlID int64, stats *models.URLStats) error {
	query := `SELECT COUNT(DISTINCT ip_address) FROM analytics WHERE url_id = $1 AND ip_address IS NOT NULL`
	err := r.db.QueryRowContext(ctx, query, urlID).Scan(&stats.UniqueIPs)
	if err != nil {
		return fmt.Errorf("ошибка получения уникальных IP: %w", err)
	}
	return nil
}

// getClicksByDate получает клики сгруппированные по датам
func (r *analyticsRepository) getClicksByDate(ctx context.Context, urlID int64, stats *models.URLStats) error {
	query := `
		SELECT DATE(clicked_at) as date, COUNT(*) as count
		FROM analytics
		WHERE url_id = $1
		GROUP BY DATE(clicked_at)
		ORDER BY date DESC
		LIMIT 30
	`

	rows, err := r.db.QueryContext(ctx, query, urlID)
	if err != nil {
		return fmt.Errorf("ошибка получения кликов по датам: %w", err)
	}
	defer rows.Close()

	stats.ClicksByDate = []models.ClicksByDate{}
	for rows.Next() {
		var item models.ClicksByDate
		if err := rows.Scan(&item.Date, &item.Count); err != nil {
			return fmt.Errorf("ошибка сканирования кликов по датам: %w", err)
		}
		stats.ClicksByDate = append(stats.ClicksByDate, item)
	}

	return rows.Err()
}

// getClicksByCountry получает клики сгруппированные по странам
func (r *analyticsRepository) getClicksByCountry(ctx context.Context, urlID int64, stats *models.URLStats) error {
	query := `
		SELECT country, COUNT(*) as count
		FROM analytics
		WHERE url_id = $1 AND country IS NOT NULL
		GROUP BY country
		ORDER BY count DESC
		LIMIT 10
	`

	rows, err := r.db.QueryContext(ctx, query, urlID)
	if err != nil {
		return fmt.Errorf("ошибка получения кликов по странам: %w", err)
	}
	defer rows.Close()

	stats.ClicksByCountry = []models.ClicksByCountry{}
	for rows.Next() {
		var item models.ClicksByCountry
		if err := rows.Scan(&item.Country, &item.Count); err != nil {
			return fmt.Errorf("ошибка сканирования кликов по странам: %w", err)
		}
		stats.ClicksByCountry = append(stats.ClicksByCountry, item)
	}

	return rows.Err()
}

// getRecentClicks получает последние клики
func (r *analyticsRepository) getRecentClicks(ctx context.Context, urlID int64, limit int, stats *models.URLStats) error {
	query := `
		SELECT id, url_id, clicked_at, ip_address, user_agent, referer, country, city
		FROM analytics
		WHERE url_id = $1
		ORDER BY clicked_at DESC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, urlID, limit)
	if err != nil {
		return fmt.Errorf("ошибка получения последних кликов: %w", err)
	}
	defer rows.Close()

	stats.RecentClicks = []models.Analytics{}
	for rows.Next() {
		var item models.Analytics
		if err := rows.Scan(
			&item.ID,
			&item.URLID,
			&item.ClickedAt,
			&item.IPAddress,
			&item.UserAgent,
			&item.Referer,
			&item.Country,
			&item.City,
		); err != nil {
			return fmt.Errorf("ошибка сканирования последних кликов: %w", err)
		}
		stats.RecentClicks = append(stats.RecentClicks, item)
	}

	return rows.Err()
}

// nullString конвертирует string в sql.NullString
func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}
