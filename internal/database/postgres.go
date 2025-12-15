package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

// PostgresConfig конфигурация для PostgreSQL
type PostgresConfig struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// NewPostgresDB создает новое подключение к PostgreSQL
func NewPostgresDB(config PostgresConfig) (*sql.DB, error) {
	db, err := sql.Open("postgres", config.DSN)
	if err != nil {
		return nil, fmt.Errorf("ошибка открытия подключения к PostgreSQL: %w", err)
	}

	// Настройка пула соединений
	if config.MaxOpenConns > 0 {
		db.SetMaxOpenConns(config.MaxOpenConns)
	} else {
		db.SetMaxOpenConns(25) // значение по умолчанию
	}

	if config.MaxIdleConns > 0 {
		db.SetMaxIdleConns(config.MaxIdleConns)
	} else {
		db.SetMaxIdleConns(5) // значение по умолчанию
	}

	if config.ConnMaxLifetime > 0 {
		db.SetConnMaxLifetime(config.ConnMaxLifetime)
	} else {
		db.SetConnMaxLifetime(5 * time.Minute) // значение по умолчанию
	}

	// Проверка подключения
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ошибка проверки подключения к PostgreSQL: %w", err)
	}

	return db, nil
}

// CloseDB закрывает подключение к базе данных
func CloseDB(db *sql.DB) error {
	if db != nil {
		return db.Close()
	}
	return nil
}
