package database

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
)

// RedisConfig конфигурация для Redis
type RedisConfig struct {
	Address  string
	Password string
	DB       int
}

// NewRedisClient создает новое подключение к Redis
func NewRedisClient(config RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     config.Address,
		Password: config.Password,
		DB:       config.DB,
	})

	// Проверка подключения
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("ошибка подключения к Redis: %w", err)
	}

	return client, nil
}

// CloseRedis закрывает подключение к Redis
func CloseRedis(client *redis.Client) error {
	if client != nil {
		return client.Close()
	}
	return nil
}
