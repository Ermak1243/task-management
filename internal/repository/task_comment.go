package repository

import (
	"context"
	"task-management/internal/models"

	"github.com/jmoiron/sqlx"
)

// TaskComment реализует операции с комментариями к задачам.
type TaskComment struct {
	db *sqlx.DB
}

// NewTaskComment создаёт новый экземпляр репозитория комментариев.
func NewTaskComment(db *sqlx.DB) *TaskComment {
	return &TaskComment{db: db}
}

// Create добавляет новый комментарий к задаче. Если tx != nil, запрос
// выполняется в рамках переданной транзакции; иначе — через основное
// соединение. После вставки записывает ID в c.ID.
func (tc *TaskComment) Create(ctx context.Context, tx *sqlx.Tx, c *models.TaskComment) error {
	res, err := executor(tc.db, tx).ExecContext(ctx,
		"INSERT INTO task_comments (task_id, user_id, content, created_at) VALUES (?, ?, ?, ?)",
		c.TaskID, c.UserID, c.Content, c.CreatedAt)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	c.ID = id
	return nil
}

// ByTask возвращает все комментарии задачи, отсортированные по дате создания.
func (tc *TaskComment) ByTask(ctx context.Context, taskID int64) ([]models.TaskComment, error) {
	var comments []models.TaskComment
	err := tc.db.SelectContext(ctx, &comments,
		"SELECT * FROM task_comments WHERE task_id=? ORDER BY created_at", taskID)
	return comments, err
}
