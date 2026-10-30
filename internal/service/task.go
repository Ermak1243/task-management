package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"task-management/internal/models"

	"github.com/jmoiron/sqlx"
)

type TaskRepository interface {
	ByID(ctx context.Context, id int64) (*models.Task, error)
	ListByTeam(ctx context.Context, teamID int64, status *string, assigneeID *int64, offset, limit int, cursor *int64) ([]models.Task, error)
	Create(ctx context.Context, tx *sqlx.Tx, t *models.Task) error
	Update(ctx context.Context, tx *sqlx.Tx, t *models.Task) error
}

type TaskHistoryRepository interface {
	Create(ctx context.Context, tx *sqlx.Tx, h *models.TaskHistory) error
	ByTask(ctx context.Context, taskID int64) ([]models.TaskHistory, error)
}

type TaskCommentRepository interface {
	Create(ctx context.Context, tx *sqlx.Tx, c *models.TaskComment) error
	ByTask(ctx context.Context, taskID int64) ([]models.TaskComment, error)
}

type TaskCache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, val []byte, ttl time.Duration) error
	DelByPattern(ctx context.Context, pattern string) error
}

// TeamMembership — минимальный контракт для проверки участия пользователя в команде,
// который нужен сервису Task (методы Role и IsAdminOrOwner).
// Полный контракт управления участниками объявлен в team.go (TeamMemberRepository).
type TeamMembership interface {
	Role(ctx context.Context, tx *sqlx.Tx, teamID, userID int64) (string, error)
	IsAdminOrOwner(ctx context.Context, tx *sqlx.Tx, teamID, userID int64) (bool, error)
}

type Task struct {
	taskRepo    TaskRepository
	historyRepo TaskHistoryRepository
	commentRepo TaskCommentRepository
	memberRepo  TeamMembership
	cache       TaskCache
	txManager   TransactionManager
	timeout     time.Duration
}

// NewTask создаёт экземпляр сервиса задач.
func NewTask(
	taskRepo TaskRepository,
	historyRepo TaskHistoryRepository,
	commentRepo TaskCommentRepository,
	memberRepo TeamMembership,
	cache TaskCache,
	txManager TransactionManager,
	timeout time.Duration,
) *Task {
	return &Task{
		taskRepo:    taskRepo,
		historyRepo: historyRepo,
		commentRepo: commentRepo,
		memberRepo:  memberRepo,
		cache:       cache,
		txManager:   txManager,
		timeout:     timeout,
	}
}

// ListByTeam возвращает список задач команды с фильтрацией и пагинацией.
// Сначала проверяет, что пользователь является участником команды.
// Результат кешируется в Redis на 5 минут; при попадании в кеш
// возвращается десериализованная копия без обращения к БД.
func (t *Task) ListByTeam(ctx context.Context, teamID int64, userID int64, status *string, assigneeID *int64, offset, limit int, cursor *int64) ([]models.Task, error) {
	ctx, cancel := context.WithTimeout(ctx, t.timeout)
	defer cancel()

	role, err := t.memberRepo.Role(ctx, nil, teamID, userID)
	if err != nil || role == "" {
		return nil, errors.New("user not member of team")
	}

	cacheKey := fmt.Sprintf("tasks:team:%d:status:%v:assignee:%v:offset:%d:limit:%d:cursor:%v", teamID, status, assigneeID, offset, limit, cursor)
	cached, err := t.cache.Get(ctx, cacheKey)
	if err == nil {
		var tasks []models.Task
		if json.Unmarshal([]byte(cached), &tasks) == nil {
			return tasks, nil
		}
	}

	tasks, err := t.taskRepo.ListByTeam(ctx, teamID, status, assigneeID, offset, limit, cursor)
	if err != nil {
		return nil, err
	}
	data, _ := json.Marshal(tasks)
	_ = t.cache.Set(ctx, cacheKey, data, 5*time.Minute)
	return tasks, nil
}

// Create создаёт новую задачу в команде. Проверяет, что создатель и назначенный
// исполнитель (если указан) являются участниками команды. Задача создаётся
// со статусом «todo» внутри транзакции вместе с записью в истории изменений.
// После создания инвалидирует кеш задач команды.
func (t *Task) Create(ctx context.Context, task *models.Task, userID int64) (*models.Task, error) {
	ctx, cancel := context.WithTimeout(ctx, t.timeout)
	defer cancel()

	role, err := t.memberRepo.Role(ctx, nil, task.TeamID, userID)
	if err != nil || role == "" {
		return nil, errors.New("user not member of team")
	}
	if task.AssigneeID != nil {
		r, err := t.memberRepo.Role(ctx, nil, task.TeamID, *task.AssigneeID)
		if err != nil || r == "" {
			return nil, errors.New("assignee not member of team")
		}
	}

	task.CreatedBy = userID
	task.Status = "todo"
	task.CreatedAt = time.Now()
	task.UpdatedAt = task.CreatedAt

	if err := t.txManager.RunInTransaction(ctx, func(txCtx context.Context, tx *sqlx.Tx) error {
		if err := t.taskRepo.Create(txCtx, tx, task); err != nil {
			return err
		}
		changes, _ := json.Marshal(map[string]interface{}{"action": "created"})
		history := &models.TaskHistory{
			TaskID:    task.ID,
			ChangedBy: userID,
			Changes:   string(changes),
			CreatedAt: time.Now(),
		}
		return t.historyRepo.Create(txCtx, tx, history)
	}); err != nil {
		return nil, err
	}

	t.invalidateCache(ctx, task.TeamID)
	return task, nil
}

// Update частично обновляет задачу. Проверяет права доступа:
//   - admin/owner — полные права;
//   - создатель задачи — полные права;
//   - исполнитель — может менять только статус.
//
// При переводе задачи в статус «done» автоматически устанавливается closed_at.
// Все изменения фиксируются в истории (task_history) внутри транзакции.
// Используется оптимистичная блокировка (version) для защиты от конкурентных обновлений.
func (t *Task) Update(ctx context.Context, taskID int64, req *models.UpdateTaskRequest, userID int64) (*models.Task, error) {
	ctx, cancel := context.WithTimeout(ctx, t.timeout)
	defer cancel()

	task, err := t.taskRepo.ByID(ctx, taskID)
	if err != nil || task == nil {
		return nil, errors.New("task not found")
	}

	isAdmin, err := t.memberRepo.IsAdminOrOwner(ctx, nil, task.TeamID, userID)
	if err != nil {
		return nil, err
	}
	isCreator := task.CreatedBy == userID
	isAssignee := task.AssigneeID != nil && *task.AssigneeID == userID

	if !isAdmin && !isCreator && !isAssignee {
		return nil, errors.New("permission denied")
	}
	if !isAdmin && !isCreator && isAssignee {
		if req.Title != nil || req.Description != nil || req.AssigneeID != nil {
			return nil, errors.New("as assignee you can only change status")
		}
	}

	old := *task
	changes := make(map[string]interface{})

	if req.Title != nil {
		task.Title = *req.Title
		changes["title"] = map[string]string{"old": old.Title, "new": task.Title}
	}
	if req.Description != nil {
		task.Description = *req.Description
		changes["description"] = map[string]string{"old": old.Description, "new": task.Description}
	}
	if req.Status != nil {
		task.Status = *req.Status
		if *req.Status == "done" && task.ClosedAt == nil {
			now := time.Now()
			task.ClosedAt = &now
		}
		if *req.Status != "done" && task.ClosedAt != nil {
			task.ClosedAt = nil
		}
		changes["status"] = map[string]string{"old": old.Status, "new": task.Status}
	}
	if req.AssigneeID != nil {
		if *req.AssigneeID == 0 {
			task.AssigneeID = nil
			changes["assignee_id"] = map[string]interface{}{"old": old.AssigneeID, "new": nil}
		} else {
			r, err := t.memberRepo.Role(ctx, nil, task.TeamID, *req.AssigneeID)
			if err != nil || r == "" {
				return nil, errors.New("new assignee not member of team")
			}
			task.AssigneeID = req.AssigneeID
			changes["assignee_id"] = map[string]interface{}{"old": old.AssigneeID, "new": req.AssigneeID}
		}
	}

	task.UpdatedAt = time.Now()

	if err := t.txManager.RunInTransaction(ctx, func(txCtx context.Context, tx *sqlx.Tx) error {
		if err := t.taskRepo.Update(txCtx, tx, task); err != nil {
			return err
		}
		if len(changes) > 0 {
			data, _ := json.Marshal(changes)
			history := &models.TaskHistory{
				TaskID:    taskID,
				ChangedBy: userID,
				Changes:   string(data),
				CreatedAt: time.Now(),
			}
			if err := t.historyRepo.Create(txCtx, tx, history); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	t.invalidateCache(ctx, task.TeamID)
	return task, nil
}

// History возвращает журнал изменений задачи. Доступ разрешён только
// участникам команды, которой принадлежит задача.
func (t *Task) History(ctx context.Context, taskID int64, userID int64) ([]models.TaskHistory, error) {
	ctx, cancel := context.WithTimeout(ctx, t.timeout)
	defer cancel()

	task, err := t.taskRepo.ByID(ctx, taskID)
	if err != nil || task == nil {
		return nil, errors.New("task not found")
	}
	role, err := t.memberRepo.Role(ctx, nil, task.TeamID, userID)
	if err != nil || role == "" {
		return nil, errors.New("access denied")
	}
	return t.historyRepo.ByTask(ctx, taskID)
}

// Comments возвращает список комментариев задачи. Доступ разрешён только
// участникам команды, которой принадлежит задача.
func (t *Task) Comments(ctx context.Context, taskID int64, userID int64) ([]models.TaskComment, error) {
	ctx, cancel := context.WithTimeout(ctx, t.timeout)
	defer cancel()

	task, err := t.taskRepo.ByID(ctx, taskID)
	if err != nil || task == nil {
		return nil, errors.New("task not found")
	}
	role, err := t.memberRepo.Role(ctx, nil, task.TeamID, userID)
	if err != nil || role == "" {
		return nil, errors.New("access denied")
	}
	return t.commentRepo.ByTask(ctx, taskID)
}

// AddComment добавляет комментарий к задаче. Проверяет, что пользователь
// является участником команды. Возвращает созданный комментарий.
func (t *Task) AddComment(ctx context.Context, taskID int64, content string, userID int64) (*models.TaskComment, error) {
	ctx, cancel := context.WithTimeout(ctx, t.timeout)
	defer cancel()

	task, err := t.taskRepo.ByID(ctx, taskID)
	if err != nil || task == nil {
		return nil, errors.New("task not found")
	}
	role, err := t.memberRepo.Role(ctx, nil, task.TeamID, userID)
	if err != nil || role == "" {
		return nil, errors.New("access denied")
	}
	comment := &models.TaskComment{
		TaskID:    taskID,
		UserID:    userID,
		Content:   content,
		CreatedAt: time.Now(),
	}
	if err := t.commentRepo.Create(ctx, nil, comment); err != nil {
		return nil, err
	}
	return comment, nil
}

// invalidateCache удаляет все закешированные списки задач команды
// по паттерну «tasks:team:{teamID}:*».
func (t *Task) invalidateCache(ctx context.Context, teamID int64) {
	pattern := fmt.Sprintf("tasks:team:%d:*", teamID)
	_ = t.cache.DelByPattern(ctx, pattern)
}
