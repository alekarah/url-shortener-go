package repository

import (
	"context"
	"database/sql"
	"fmt"

	"url-short/internal/models"
)

// URLRepository интерфейс для работы с URL в БД
type URLRepository interface {
	Create(ctx context.Context, url *models.URL) error
	GetByShortCode(ctx context.Context, shortCode string) (*models.URL, error)
	GetByID(ctx context.Context, id int64) (*models.URL, error)
	Update(ctx context.Context, url *models.URL) error
	Delete(ctx context.Context, id int64) error
	IncrementClicks(ctx context.Context, id int64) error
	ShortCodeExists(ctx context.Context, shortCode string) (bool, error)
}

// urlRepository имплементация URLRepository
type urlRepository struct {
	db *sql.DB
}

// NewURLRepository создает новый URL repository
func NewURLRepository(db *sql.DB) URLRepository {
	return &urlRepository{db: db}
}

// Create создает новую короткую ссылку
func (r *urlRepository) Create(ctx context.Context, url *models.URL) error {
	query := `
		INSERT INTO urls (short_code, original_url, expires_at, user_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, clicks_count
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		url.ShortCode,
		url.OriginalURL,
		url.ExpiresAt,
		url.UserID,
	).Scan(&url.ID, &url.CreatedAt, &url.ClicksCount)

	if err != nil {
		return fmt.Errorf("ошибка создания URL: %w", err)
	}

	return nil
}

// GetByShortCode получает URL по короткому коду
func (r *urlRepository) GetByShortCode(ctx context.Context, shortCode string) (*models.URL, error) {
	query := `
		SELECT id, short_code, original_url, created_at, expires_at, user_id, clicks_count, last_clicked_at
		FROM urls
		WHERE short_code = $1
	`

	url := &models.URL{}
	err := r.db.QueryRowContext(ctx, query, shortCode).Scan(
		&url.ID,
		&url.ShortCode,
		&url.OriginalURL,
		&url.CreatedAt,
		&url.ExpiresAt,
		&url.UserID,
		&url.ClicksCount,
		&url.LastClickedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("URL с кодом %s не найден", shortCode)
	}

	if err != nil {
		return nil, fmt.Errorf("ошибка получения URL: %w", err)
	}

	return url, nil
}

// GetByID получает URL по ID
func (r *urlRepository) GetByID(ctx context.Context, id int64) (*models.URL, error) {
	query := `
		SELECT id, short_code, original_url, created_at, expires_at, user_id, clicks_count, last_clicked_at
		FROM urls
		WHERE id = $1
	`

	url := &models.URL{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&url.ID,
		&url.ShortCode,
		&url.OriginalURL,
		&url.CreatedAt,
		&url.ExpiresAt,
		&url.UserID,
		&url.ClicksCount,
		&url.LastClickedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("URL с ID %d не найден", id)
	}

	if err != nil {
		return nil, fmt.Errorf("ошибка получения URL: %w", err)
	}

	return url, nil
}

// Update обновляет URL
func (r *urlRepository) Update(ctx context.Context, url *models.URL) error {
	query := `
		UPDATE urls
		SET original_url = $1, expires_at = $2
		WHERE id = $3
	`

	result, err := r.db.ExecContext(ctx, query, url.OriginalURL, url.ExpiresAt, url.ID)
	if err != nil {
		return fmt.Errorf("ошибка обновления URL: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("ошибка проверки обновления: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("URL с ID %d не найден", url.ID)
	}

	return nil
}

// Delete удаляет URL
func (r *urlRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM urls WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("ошибка удаления URL: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("ошибка проверки удаления: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("URL с ID %d не найден", id)
	}

	return nil
}

// IncrementClicks увеличивает счетчик кликов
func (r *urlRepository) IncrementClicks(ctx context.Context, id int64) error {
	query := `
		UPDATE urls
		SET clicks_count = clicks_count + 1,
		    last_clicked_at = NOW()
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("ошибка увеличения счетчика кликов: %w", err)
	}

	return nil
}

// ShortCodeExists проверяет существование короткого кода
func (r *urlRepository) ShortCodeExists(ctx context.Context, shortCode string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM urls WHERE short_code = $1)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, shortCode).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("ошибка проверки существования кода: %w", err)
	}

	return exists, nil
}
