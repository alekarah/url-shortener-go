package models

import (
	"database/sql"
	"time"
)

// URL представляет короткую ссылку
type URL struct {
	ID            int64        `json:"id"`
	ShortCode     string       `json:"short_code"`
	OriginalURL   string       `json:"original_url"`
	CreatedAt     time.Time    `json:"created_at"`
	ExpiresAt     sql.NullTime `json:"expires_at,omitempty"`
	UserID        sql.NullInt64 `json:"user_id,omitempty"`
	ClicksCount   int64        `json:"clicks_count"`
	LastClickedAt sql.NullTime `json:"last_clicked_at,omitempty"`
}

// CreateURLRequest запрос на создание короткой ссылки
type CreateURLRequest struct {
	OriginalURL string `json:"original_url" validate:"required,url"`
	CustomCode  string `json:"custom_code,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

// URLResponse ответ с информацией о ссылке
type URLResponse struct {
	ID          int64     `json:"id"`
	ShortCode   string    `json:"short_code"`
	ShortURL    string    `json:"short_url"`
	OriginalURL string    `json:"original_url"`
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	ClicksCount int64     `json:"clicks_count"`
}
