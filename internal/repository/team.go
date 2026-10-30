package repository

import (
	"context"
	"database/sql"

	"task-management/internal/models"

	"github.com/jmoiron/sqlx"
)

// Team реализует CRUD-операции для сущности Team в MySQL.
type Team struct {
	db *sqlx.DB
}

// NewTeam создаёт новый экземпляр репозитория команд.
func NewTeam(db *sqlx.DB) *Team {
	return &Team{db: db}
}

// ByID возвращает команду по её ID. Если не найдена — (nil, nil).
func (t *Team) ByID(ctx context.Context, id int64) (*models.Team, error) {
	var team models.Team
	err := t.db.GetContext(ctx, &team, "SELECT * FROM teams WHERE id=?", id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &team, err
}

// ByUser возвращает все команды, в которых состоит пользователь (через JOIN с team_members).
func (t *Team) ByUser(ctx context.Context, userID int64) ([]models.Team, error) {
	var teams []models.Team
	err := t.db.SelectContext(ctx, &teams,
		`SELECT t.* FROM teams t
         JOIN team_members tm ON t.id = tm.team_id
         WHERE tm.user_id = ?`, userID)
	return teams, err
}

// Create вставляет новую команду в БД. Если tx != nil, запрос выполняется
// в рамках переданной транзакции; иначе — через основное соединение.
// После вставки записывает ID в team.ID.
func (t *Team) Create(ctx context.Context, tx *sqlx.Tx, team *models.Team) error {
	res, err := executor(t.db, tx).ExecContext(ctx,
		"INSERT INTO teams (name, created_by, created_at) VALUES (?, ?, ?)",
		team.Name, team.CreatedBy, team.CreatedAt)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	team.ID = id
	return nil
}
