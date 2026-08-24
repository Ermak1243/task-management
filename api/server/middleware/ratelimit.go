package middleware

import (
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
)

// RateLimit возвращает мидлварь, ограничивающий количество запросов
// от одного IP-адреса. Алгоритм: INCR + EXPIRE в Redis. При превышении
// лимита возвращается 429 Too Many Requests.
func RateLimit(rdb *redis.Client, prefix string, maxRequests int, windowSeconds int64) fiber.Handler {
	return func(c *fiber.Ctx) error {
		key := fmt.Sprintf("ratelimit:%s:%s", prefix, c.IP())
		ctx := c.Context()
		count, err := rdb.Incr(ctx, key).Result()
		if err != nil {
			return c.Next()
		}
		if count == 1 {
			rdb.Expire(ctx, key, time.Duration(windowSeconds)*time.Second)
		}
		if count > int64(maxRequests) {
			return c.Status(429).JSON(fiber.Map{"error": "too many requests"})
		}
		return c.Next()
	}
}
