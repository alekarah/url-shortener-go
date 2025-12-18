package models

import (
	"database/sql"
	"time"
)

// Analytics представляет запись о клике по короткой ссылке
type Analytics struct {
	ID        int64          `json:"id"`
	URLID     int64          `json:"url_id"`
	ClickedAt time.Time      `json:"clicked_at"`
	IPAddress sql.NullString `json:"ip_address,omitempty"`
	UserAgent sql.NullString `json:"user_agent,omitempty"`
	Referer   sql.NullString `json:"referer,omitempty"`
	Country   sql.NullString `json:"country,omitempty"`
	City      sql.NullString `json:"city,omitempty"`
}

// ClickEvent данные о клике для записи
type ClickEvent struct {
	URLID     int64
	IPAddress string
	UserAgent string
	Referer   string
	Country   string
	City      string
}

// URLStats статистика по URL
type URLStats struct {
	TotalClicks     int64             `json:"total_clicks"`
	UniqueIPs       int64             `json:"unique_ips"`
	ClicksByDate    []ClicksByDate    `json:"clicks_by_date"`
	ClicksByCountry []ClicksByCountry `json:"clicks_by_country"`
	RecentClicks    []Analytics       `json:"recent_clicks"`
}

// ClicksByDate клики по датам
type ClicksByDate struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

// ClicksByCountry клики по странам
type ClicksByCountry struct {
	Country string `json:"country"`
	Count   int64  `json:"count"`
}
