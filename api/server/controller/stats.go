package controller

import (
	"strconv"

	_ "task-management/internal/models"
	"task-management/internal/service"

	"github.com/gofiber/fiber/v2"
)

type Stats struct {
	svc *service.Stats
}

func NewStats(svc *service.Stats) *Stats {
	return &Stats{svc: svc}
}

// TeamStats godoc
// @Summary Get team statistics
// @Tags stats
// @Security ApiKeyAuth
// @Produce json
// @Param team_id path int true "Team ID"
// @Success 200 {object} models.TeamStats
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /teams/{team_id}/stats [get]
func (s *Stats) TeamStats(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(int64)
	teamID, err := strconv.ParseInt(c.Params("team_id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid team_id"})
	}
	stats, err := s.svc.GetTeamStats(c.Context(), teamID, userID)
	if err != nil {
		return RespondError(c, err)
	}
	return c.JSON(stats)
}
