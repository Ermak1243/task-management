package redis

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

// Cache — реализация кеша на основе Redis с поддержкой TTL
// и массового удаления по паттерну через SCAN.
type Cache struct {
	client *redis.Client
}

// NewCache создаёт экземпляр кеша на основе переданного Redis-клиента.
func NewCache(client *redis.Client) *Cache {
	return &Cache{client: client}
}

// Get возвращает значение по ключу из Redis. Если ключ не найден — ошибка.
func (c *Cache) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, key).Result()
}

// Set сохраняет значение в Redis с указанным TTL.
func (c *Cache) Set(ctx context.Context, key string, val []byte, ttl time.Duration) error {
	return c.client.Set(ctx, key, val, ttl).Err()
}

// DelByPattern удаляет все ключи, соответствующие паттерну, используя
// итеративный SCAN для безопасной работы с большими базами данных.
func (c *Cache) DelByPattern(ctx context.Context, pattern string) error {
	iter := c.client.Scan(ctx, 0, pattern, 0).Iterator()
	var keys []string
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return err
	}
	if len(keys) > 0 {
		return c.client.Del(ctx, keys...).Err()
	}
	return nil
}
