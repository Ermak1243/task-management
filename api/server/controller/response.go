package controller

import (
	"task-management/internal/apperrors"

	"github.com/gofiber/fiber/v2"
)

// RespondError формирует HTTP-ответ на ошибку.
func RespondError(c *fiber.Ctx, err error) error {
	if appErr, ok := err.(*apperrors.AppError); ok {
		return c.Status(appErr.Status).JSON(fiber.Map{
			"error":   appErr.Code,
			"message": appErr.Message,
		})
	}
	return c.Status(500).JSON(fiber.Map{"error": "internal_error"})
}
