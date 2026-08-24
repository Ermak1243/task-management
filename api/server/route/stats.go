package route

import (
	"task-management/api/server/controller"
	"task-management/internal/service"

	"github.com/gofiber/fiber/v2"
)

// SetupStats регистрирует маршрут получения статистики по команде.
func SetupStats(group fiber.Router, svc *service.Stats) {
	ctrl := controller.NewStats(svc)
	group.Get("/teams/:team_id/stats", ctrl.TeamStats)
}
