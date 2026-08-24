package route

import (
	"task-management/api/server/controller"
	"task-management/internal/service"

	"github.com/gofiber/fiber/v2"
)

// SetupAuth регистрирует маршруты аутентификации: регистрацию, вход и выход.
func SetupAuth(group fiber.Router, svc *service.Auth) {
	ctrl := controller.NewAuth(svc)
	group.Post("/register", ctrl.Register)
	group.Post("/login", ctrl.Login)
	group.Post("/logout", ctrl.Logout)
}
