package route

import (
	"task-management/api/server/controller"
	"task-management/internal/service"

	"github.com/gofiber/fiber/v2"
)

// SetupTeam регистрирует маршруты управления командами:
// создание, список, приглашение, изменение роли, удаление участника.
func SetupTeam(group fiber.Router, svc *service.Team) {
	ctrl := controller.NewTeam(svc)
	group.Post("/teams", ctrl.Create)
	group.Get("/teams", ctrl.List)
	group.Post("/teams/:team_id/invite", ctrl.Invite)
	group.Put("/teams/:team_id/role", ctrl.ChangeRole)
	group.Delete("/teams/:team_id/members/:user_id", ctrl.RemoveMember)
}
