package repository

import (
	"context"
	"task-management/internal/models"

	"github.com/jmoiron/sqlx"
)

// TaskHistory реализует операции с журналом изменений задач.
type TaskHistory struct {
	db *sqlx.DB
}

// NewTaskHistory создаёт новый экземпляр репозитория истории задач.
func NewTaskHistory(db *sqlx.DB) *TaskHistory {
	return &TaskHistory{db: db}
}

// Create записывает запись об изменении задачи. Если tx != nil, запрос
// выполняется в рамках переданной транзакции; иначе — через основное соединение.
func (th *TaskHistory) Create(ctx context.Context, tx *sqlx.Tx, h *models.TaskHistory) error {
	_, err := executor(th.db, tx).ExecContext(ctx,
		"INSERT INTO task_history (task_id, changed_by, changes, created_at) VALUES (?, ?, ?, ?)",
		h.TaskID, h.ChangedBy, h.Changes, h.CreatedAt)
	return err
}

// ByTask возвращает всю историю изменений задачи, отсортированную по дате.
func (th *TaskHistory) ByTask(ctx context.Context, taskID int64) ([]models.TaskHistory, error) {
	var history []models.TaskHistory
	err := th.db.SelectContext(ctx, &history,
		"SELECT * FROM task_history WHERE task_id=? ORDER BY created_at", taskID)
	return history, err
}
