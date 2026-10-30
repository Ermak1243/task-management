package repository

import (
	"context"
	"database/sql"

	"task-management/internal/apperrors"
	"task-management/internal/models"

	"github.com/jmoiron/sqlx"
)

// Task реализует CRUD-операции для сущности Task в MySQL.
type Task struct {
	db *sqlx.DB
}

// NewTask создаёт новый экземпляр репозитория задач.
func NewTask(db *sqlx.DB) *Task {
	return &Task{db: db}
}

// ByID возвращает задачу по её ID. Если не найдена — (nil, nil).
func (t *Task) ByID(ctx context.Context, id int64) (*models.Task, error) {
	var task models.Task
	err := t.db.GetContext(ctx, &task, "SELECT * FROM tasks WHERE id=?", id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &task, err
}

// ListByTeam возвращает список задач команды с фильтрацией по статусу
// и исполнителю. Поддерживает два режима пагинации:
//   - курсорная: если cursor != nil, выборка идёт по id > cursor LIMIT n;
//   - offset-ная: иначе LIMIT offset, OFFSET limit.
func (t *Task) ListByTeam(ctx context.Context, teamID int64, status *string, assigneeID *int64,
	offset, limit int, cursor *int64) ([]models.Task, error) {
	query := "SELECT * FROM tasks WHERE team_id = ?"
	args := []interface{}{teamID}
	if status != nil {
		query += " AND status = ?"
		args = append(args, *status)
	}
	if assigneeID != nil {
		query += " AND assignee_id = ?"
		args = append(args, *assigneeID)
	}
	if cursor != nil {
		query += " AND id > ?"
		args = append(args, *cursor)
		query += " ORDER BY id LIMIT ?"
		args = append(args, limit)
	} else {
		query += " ORDER BY id LIMIT ? OFFSET ?"
		args = append(args, limit, offset)
	}

	var tasks []models.Task
	err := t.db.SelectContext(ctx, &tasks, query, args...)
	return tasks, err
}

// Create вставляет новую задачу в БД с начальным version=1. Если tx != nil,
// запрос выполняется в рамках переданной транзакции; иначе — через
// основное соединение.
func (t *Task) Create(ctx context.Context, tx *sqlx.Tx, task *models.Task) error {
	res, err := executor(t.db, tx).ExecContext(ctx,
		`INSERT INTO tasks (team_id, title, description, status, created_by, assignee_id, created_at, updated_at, version)
         VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		task.TeamID, task.Title, task.Description, task.Status, task.CreatedBy, task.AssigneeID, task.CreatedAt, task.UpdatedAt, 1)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	task.ID = id
	return nil
}

// Update обновляет задачу с оптимистичной блокировкой: WHERE id = ? AND version = ?.
// Если rowsAffected == 0, значит версия уже изменилась (конфликт) или задача
// не найдена — возвращается ошибка. При успешном обновлении version увеличивается на 1.
func (t *Task) Update(ctx context.Context, tx *sqlx.Tx, task *models.Task) error {
	query := `UPDATE tasks SET
        title = ?,
        description = ?,
        status = ?,
        assignee_id = ?,
        updated_at = ?,
        closed_at = ?,
        version = version + 1
    WHERE id = ? AND version = ?`
	res, err := executor(t.db, tx).ExecContext(ctx, query,
		task.Title, task.Description, task.Status, task.AssigneeID, task.UpdatedAt, task.ClosedAt, task.ID, task.Version)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return apperrors.NewConflict("task version mismatch or task not found")
	}
	task.Version++
	return nil
}
