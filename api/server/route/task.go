package route

import (
	"task-management/api/server/controller"
	"task-management/internal/service"

	"github.com/gofiber/fiber/v2"
)

// SetupTask регистрирует маршруты CRUD-операций над задачами:
// список, создание, обновление, история, комментарии.
func SetupTask(group fiber.Router, svc *service.Task) {
	ctrl := controller.NewTask(svc)
	group.Get("/tasks", ctrl.List)
	group.Post("/tasks", ctrl.Create)
	group.Put("/tasks/:id", ctrl.Update)
	group.Get("/tasks/:id/history", ctrl.History)
	group.Get("/tasks/:id/comments", ctrl.Comments)
	group.Post("/tasks/:id/comments", ctrl.AddComment)
}
