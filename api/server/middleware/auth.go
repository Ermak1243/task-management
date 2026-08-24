// Package middleware содержит HTTP-мидлвари: аутентификацию JWT,
// настройку CORS и ограничение частоты запросов (rate limiting).
package middleware

import (
	"strings"

	"task-management/pkg/jwt"

	"github.com/gofiber/fiber/v2"
)

// JWTAuth возвращает мидлварь, который извлекает Bearer-токен из заголовка
// Authorization, валидирует его и сохраняет user_id вlocals контекста Fiber.
// При отсутствии или невалидном токене возвращает 401.
func JWTAuth(secret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		auth := c.Get("Authorization")
		if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
			return c.Status(401).JSON(fiber.Map{"error": "missing or invalid token"})
		}
		token := strings.TrimPrefix(auth, "Bearer ")
		userID, err := jwt.Validate(token, []byte(secret))
		if err != nil {
			return c.Status(401).JSON(fiber.Map{"error": "invalid token"})
		}
		c.Locals("user_id", userID)
		return c.Next()
	}
}
