package controller

import (
	"strconv"

	"task-management/internal/models"
	"task-management/internal/service"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type Team struct {
	svc      *service.Team
	validate *validator.Validate
}

func NewTeam(svc *service.Team) *Team {
	return &Team{
		svc:      svc,
		validate: validator.New(),
	}
}

// CreateTeam godoc
// @Summary Create a team
// @Tags teams
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param request body models.CreateTeamRequest true "Team name"
// @Success 201 {object} models.Team
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /teams [post]
func (t *Team) Create(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(int64)
	var req models.CreateTeamRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}
	if err := t.validate.Struct(req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	team, err := t.svc.Create(c.Context(), req.Name, userID)
	if err != nil {
		return RespondError(c, err)
	}
	return c.Status(201).JSON(team)
}

// ListTeams godoc
// @Summary List teams for current user
// @Tags teams
// @Security ApiKeyAuth
// @Produce json
// @Success 200 {array} models.Team
// @Router /teams [get]
func (t *Team) List(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(int64)
	teams, err := t.svc.ListByUser(c.Context(), userID)
	if err != nil {
		return RespondError(c, err)
	}
	return c.JSON(teams)
}

// InviteMember godoc
// @Summary Add user to team
// @Tags teams
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param team_id path int true "Team ID"
// @Param request body models.InviteRequest true "User ID and role"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /teams/{team_id}/invite [post]
func (t *Team) Invite(c *fiber.Ctx) error {
	teamID, err := strconv.ParseInt(c.Params("team_id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid team_id"})
	}
	userID := c.Locals("user_id").(int64)
	var req models.InviteRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}
	if err := t.validate.Struct(req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	err = t.svc.Invite(c.Context(), teamID, req.UserID, userID, req.Role)
	if err != nil {
		return RespondError(c, err)
	}
	return c.JSON(fiber.Map{"status": "invited"})
}

// ChangeRole godoc
// @Summary Change user role in team
// @Tags teams
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param team_id path int true "Team ID"
// @Param request body models.UpdateRoleRequest true "User ID and new role"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /teams/{team_id}/role [put]
func (t *Team) ChangeRole(c *fiber.Ctx) error {
	teamID, err := strconv.ParseInt(c.Params("team_id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid team_id"})
	}
	userID := c.Locals("user_id").(int64)
	var req models.UpdateRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}
	if err := t.validate.Struct(req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	err = t.svc.ChangeRole(c.Context(), teamID, req.UserID, userID, req.Role)
	if err != nil {
		return RespondError(c, err)
	}
	return c.JSON(fiber.Map{"status": "role updated"})
}

// RemoveMember godoc
// @Summary Remove user from team
// @Tags teams
// @Security ApiKeyAuth
// @Param team_id path int true "Team ID"
// @Param user_id path int true "User ID to remove"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /teams/{team_id}/members/{user_id} [delete]
func (t *Team) RemoveMember(c *fiber.Ctx) error {
	teamID, err := strconv.ParseInt(c.Params("team_id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid team_id"})
	}
	userID := c.Locals("user_id").(int64)
	targetUserID, err := strconv.ParseInt(c.Params("user_id"), 10, 64)
	if err != nil || targetUserID == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "invalid user_id"})
	}
	err = t.svc.RemoveMember(c.Context(), teamID, targetUserID, userID)
	if err != nil {
		return RespondError(c, err)
	}
	return c.JSON(fiber.Map{"status": "member removed"})
}
