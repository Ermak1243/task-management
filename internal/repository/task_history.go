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

// Create записывает запись об изменении задачи. Поддерживает прозрачную
// транзакцию через context.
func (th *TaskHistory) Create(ctx context.Context, h *models.TaskHistory) error {
	var err error
	if tx, ok := ctx.Value(TxKey).(*sqlx.Tx); ok {
		_, err = tx.ExecContext(ctx,
			"INSERT INTO task_history (task_id, changed_by, changes, created_at) VALUES (?, ?, ?, ?)",
			h.TaskID, h.ChangedBy, h.Changes, h.CreatedAt)
	} else {
		_, err = th.db.ExecContext(ctx,
			"INSERT INTO task_history (task_id, changed_by, changes, created_at) VALUES (?, ?, ?, ?)",
			h.TaskID, h.ChangedBy, h.Changes, h.CreatedAt)
	}
	return err
}

// ByTask возвращает всю историю изменений задачи, отсортированную по дате.
func (th *TaskHistory) ByTask(ctx context.Context, taskID int64) ([]models.TaskHistory, error) {
	var history []models.TaskHistory
	err := th.db.SelectContext(ctx, &history,
		"SELECT * FROM task_history WHERE task_id=? ORDER BY created_at", taskID)
	return history, err
}
