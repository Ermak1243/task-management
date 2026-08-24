package repository

import (
	"context"
	"encoding/json"

	"task-management/internal/models"

	"github.com/jmoiron/sqlx"
)

// Stats реализует агрегирующие SQL-запросы для статистики по команде.
type Stats struct {
	db *sqlx.DB
}

// NewStats создаёт новый экземпляр репозитория статистики.
func NewStats(db *sqlx.DB) *Stats {
	return &Stats{db: db}
}

// TeamStats выполняет CTE-запрос, агрегирующий статистику по команде:
// количество задач по статусам (todo/in_progress/done), среднее время закрытия,
// топ-3 исполнителя за последние 30 дней и общее число комментариев.
// Результат десериализуется в models.TeamStats.
func (r *Stats) TeamStats(ctx context.Context, teamID int64) (*models.TeamStats, error) {
	query := `
    WITH task_metrics AS (
        SELECT
            SUM(CASE WHEN status = 'todo' THEN 1 ELSE 0 END) AS todo,
            SUM(CASE WHEN status = 'in_progress' THEN 1 ELSE 0 END) AS in_progress,
            SUM(CASE WHEN status = 'done' THEN 1 ELSE 0 END) AS done,
            COALESCE(AVG(TIMESTAMPDIFF(SECOND, created_at, closed_at)), 0) AS avg_close_time_seconds
        FROM tasks
        WHERE team_id = ?
    ),
    top_assignees AS (
        SELECT
            assignee_id,
            COUNT(*) AS closed_count
        FROM tasks
        WHERE team_id = ?
            AND status = 'done'
            AND closed_at >= DATE_SUB(NOW(), INTERVAL 30 DAY)
            AND assignee_id IS NOT NULL
        GROUP BY assignee_id
        ORDER BY closed_count DESC
        LIMIT 3
    ),
    comments_count AS (
        SELECT COUNT(*) AS total_comments
        FROM task_comments tc
        JOIN tasks t ON tc.task_id = t.id
        WHERE t.team_id = ?
    )
    SELECT
        tm.todo,
        tm.in_progress,
        tm.done,
        tm.avg_close_time_seconds,
        (SELECT COALESCE(JSON_ARRAYAGG(assignee_id), '[]') FROM top_assignees) AS top_assignees_json,
        cc.total_comments
    FROM task_metrics tm
    CROSS JOIN comments_count cc
    LIMIT 1;
    `

	var todo, inProgress, done int
	var avgClose float64
	var topAssigneesJSON string
	var totalComments int

	err := r.db.QueryRowContext(ctx, query, teamID, teamID, teamID).Scan(
		&todo, &inProgress, &done, &avgClose, &topAssigneesJSON, &totalComments,
	)
	if err != nil {
		return nil, err
	}

	var topAssignees []int64
	if topAssigneesJSON != "" && topAssigneesJSON != "[]" {
		if err := json.Unmarshal([]byte(topAssigneesJSON), &topAssignees); err != nil {
			return nil, err
		}
	}

	return &models.TeamStats{
		Todo:          todo,
		InProgress:    inProgress,
		Done:          done,
		AvgCloseTime:  avgClose,
		TopAssignees:  topAssignees,
		TotalComments: totalComments,
	}, nil
}