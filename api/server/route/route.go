// Package route регистрирует HTTP-маршруты и группирует их
// по уровню доступа (публичные / защищённые JWT-ом).
package route

import (
	"task-management/api/server/middleware"
	"task-management/internal/config"
	"task-management/internal/database/redis"
	"task-management/internal/service"

	"github.com/gofiber/fiber/v2"
)

func Setup(
	app *fiber.App,
	cfg *config.Config,
	auth *service.Auth,
	team *service.Team,
	task *service.Task,
	stats *service.Stats,
	client *redis.Client,
) {
	middleware.Setup(app)

	docs := app.Group("/docs")
	SetupDocs(docs)

	api := app.Group("/api/v1")

	// Публичные маршруты
	public := api.Group("", middleware.RateLimit(client.Client, "public", 50, 60))
	SetupAuth(public, auth)

	// Защищённые маршруты
	protected := api.Group("", middleware.JWTAuth(cfg.JWTSecret), middleware.RateLimit(client.Client, "protected", 100, 60))
	SetupTeam(protected, team)
	SetupTask(protected, task)
	SetupStats(protected, stats)
}
