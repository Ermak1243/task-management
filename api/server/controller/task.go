package controller

import (
	"strconv"

	"task-management/internal/models"
	"task-management/internal/service"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type Task struct {
	svc      *service.Task
	validate *validator.Validate
}

func NewTask(svc *service.Task) *Task {
	return &Task{
		svc:      svc,
		validate: validator.New(),
	}
}

// ListTasks godoc
// @Summary List tasks
// @Tags tasks
// @Security ApiKeyAuth
// @Produce json
// @Param team_id query int true "Team ID"
// @Param status query string false "Filter by status"
// @Param assignee_id query int false "Filter by assignee"
// @Param offset query int false "Pagination offset" default(0)
// @Param limit query int false "Limit" default(20)
// @Param cursor query int false "Pagination cursor (overrides offset)"
// @Success 200 {array} models.Task
// @Failure 400 {object} map[string]string
// @Router /tasks [get]
func (t *Task) List(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(int64)
	teamID, err := strconv.ParseInt(c.Query("team_id"), 10, 64)
	if err != nil || teamID == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "team_id required"})
	}
	var status *string
	if s := c.Query("status"); s != "" {
		status = &s
	}
	var assigneeID *int64
	if a := c.Query("assignee_id"); a != "" {
		val, err := strconv.ParseInt(a, 10, 64)
		if err == nil {
			assigneeID = &val
		}
	}
	offset := 0
	if o := c.Query("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = v
		}
	}
	limit := 20
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			limit = v
		}
	}
	var cursor *int64
	if cur := c.Query("cursor"); cur != "" {
		val, err := strconv.ParseInt(cur, 10, 64)
		if err == nil {
			cursor = &val
		}
	}

	tasks, err := t.svc.ListByTeam(c.Context(), teamID, userID, status, assigneeID, offset, limit, cursor)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(tasks)
}

// CreateTask godoc
// @Summary Create task
// @Tags tasks
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param request body models.CreateTaskRequest true "Task data"
// @Success 201 {object} models.Task
// @Failure 400 {object} map[string]string
// @Router /tasks [post]
func (t *Task) Create(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(int64)
	var req models.CreateTaskRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}
	if err := t.validate.Struct(req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	task := &models.Task{
		TeamID:      req.TeamID,
		Title:       req.Title,
		Description: req.Description,
		AssigneeID:  req.AssigneeID,
	}
	created, err := t.svc.Create(c.Context(), task, userID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(201).JSON(created)
}

// UpdateTask godoc
// @Summary Update task
// @Tags tasks
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path int true "Task ID"
// @Param request body models.UpdateTaskRequest true "Update fields"
// @Success 200 {object} models.Task
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /tasks/{id} [put]
func (t *Task) Update(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(int64)
	taskID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid task id"})
	}
	var req models.UpdateTaskRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}
	if req.Status != nil {
		if err := t.validate.Var(req.Status, "omitempty,oneof=todo in_progress done"); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid status"})
		}
	}
	updated, err := t.svc.Update(c.Context(), taskID, &req, userID)
	if err != nil {
		if err.Error() == "task version mismatch or task not found" {
			return c.Status(409).JSON(fiber.Map{"error": "task was modified concurrently, please retry"})
		}
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(updated)
}

// GetHistory godoc
// @Summary Get task history
// @Tags tasks
// @Security ApiKeyAuth
// @Produce json
// @Param id path int true "Task ID"
// @Success 200 {array} models.TaskHistory
// @Failure 400 {object} map[string]string
// @Router /tasks/{id}/history [get]
func (t *Task) History(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(int64)
	taskID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid task id"})
	}
	history, err := t.svc.History(c.Context(), taskID, userID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(history)
}

// GetComments godoc
// @Summary Get task comments
// @Tags tasks
// @Security ApiKeyAuth
// @Produce json
// @Param id path int true "Task ID"
// @Success 200 {array} models.TaskComment
// @Failure 400 {object} map[string]string
// @Router /tasks/{id}/comments [get]
func (t *Task) Comments(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(int64)
	taskID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid task id"})
	}
	comments, err := t.svc.Comments(c.Context(), taskID, userID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(comments)
}

// AddComment godoc
// @Summary Add comment to task
// @Tags tasks
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path int true "Task ID"
// @Param request body models.CommentRequest true "Comment content"
// @Success 201 {object} models.TaskComment
// @Failure 400 {object} map[string]string
// @Router /tasks/{id}/comments [post]
func (t *Task) AddComment(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(int64)
	taskID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid task id"})
	}
	var req models.CommentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}
	if err := t.validate.Struct(req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	comment, err := t.svc.AddComment(c.Context(), taskID, req.Content, userID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(201).JSON(comment)
}
