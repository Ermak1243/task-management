// Package redis предоставляет обёртку над go-redis клиентом
// и реализацию кеша на основе Redis.
package redis

import (
	"context"

	"task-management/internal/config"

	"github.com/go-redis/redis/v8"
)

// Client — обёртка над *redis.Client с методом Close для корректного
// закрытия соединения.
type Client struct {
	*redis.Client
}

// New создаёт клиент Redis и проверяет доступность через PING.
// Panic, если Redis недоступен.
func New(cfg config.Redis) *Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Host + ":" + cfg.Port,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		panic(err)
	}
	return &Client{rdb}
}

// Close закрывает соединение с Redis.
func (c *Client) Close() error {
	return c.Client.Close()
}
