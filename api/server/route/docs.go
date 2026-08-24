package route

import (
	_ "task-management/docs" // Swagger docs import

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
)

// SetupDocs решистрирует маршруты документации
func SetupDocs(group fiber.Router) {
	group.Get("/swagger/*", swagger.HandlerDefault) // Swagger UI endpoint
}
