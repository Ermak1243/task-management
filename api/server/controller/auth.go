package controller

import (
	"task-management/internal/models"
	"task-management/internal/service"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type Auth struct {
	svc      *service.Auth
	validate *validator.Validate
}

func NewAuth(svc *service.Auth) *Auth {
	return &Auth{
		svc:      svc,
		validate: validator.New(),
	}
}

// Register godoc
// @Summary Register user
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.RegisterRequest true "Register data"
// @Success 201 {object} models.User
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /register [post]
func (a *Auth) Register(c *fiber.Ctx) error {
	var req models.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}
	if err := a.validate.Struct(req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	user, err := a.svc.Register(c.Context(), req.Email, req.Password, req.Name)
	if err != nil {
		return RespondError(c, err)
	}
	return c.Status(201).JSON(user)
}

// Login godoc
// @Summary Login
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.LoginRequest true "Login credentials"
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /login [post]
func (a *Auth) Login(c *fiber.Ctx) error {
	var req models.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}
	if err := a.validate.Struct(req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	token, err := a.svc.Login(c.Context(), req.Email, req.Password)
	if err != nil {
		return RespondError(c, err)
	}
	return c.JSON(fiber.Map{"token": token})
}

// Logout godoc
// @Summary Logout (client should discard token)
// @Tags auth
// @Security ApiKeyAuth
// @Success 204
// @Router /logout [post]
func (a *Auth) Logout(c *fiber.Ctx) error {
	return c.SendStatus(204)
}
