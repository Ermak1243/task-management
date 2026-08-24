package repository

import (
	"context"
	"database/sql"

	"task-management/internal/models"

	"github.com/jmoiron/sqlx"
)

// TeamMember реализует операции с таблицей team_members: добавление,
// удаление, получение и изменение роли участника команды.
type TeamMember struct {
	db *sqlx.DB
}

// NewTeamMember создаёт новый экземпляр репозитория участников команд.
func NewTeamMember(db *sqlx.DB) *TeamMember {
	return &TeamMember{db: db}
}

// Add добавляет участника в команду. Поддерживает прозрачную транзакцию через context.
func (tm *TeamMember) Add(ctx context.Context, member *models.TeamMember) error {
	var err error
	if tx, ok := ctx.Value(TxKey).(*sqlx.Tx); ok {
		_, err = tx.ExecContext(ctx,
			"INSERT INTO team_members (team_id, user_id, role) VALUES (?, ?, ?)",
			member.TeamID, member.UserID, member.Role)
	} else {
		_, err = tm.db.ExecContext(ctx,
			"INSERT INTO team_members (team_id, user_id, role) VALUES (?, ?, ?)",
			member.TeamID, member.UserID, member.Role)
	}
	return err
}

// Remove удаляет участника из команды по pair (teamID, userID).
func (tm *TeamMember) Remove(ctx context.Context, teamID, userID int64) error {
	if tx, ok := ctx.Value(TxKey).(*sqlx.Tx); ok {
		_, err := tx.ExecContext(ctx,
			"DELETE FROM team_members WHERE team_id=? AND user_id=?", teamID, userID)
		return err
	}
	_, err := tm.db.ExecContext(ctx,
		"DELETE FROM team_members WHERE team_id=? AND user_id=?", teamID, userID)
	return err
}

// Role возвращает роль пользователя в команде. Если участник не найден,
// возвращает ("", nil).
func (tm *TeamMember) Role(ctx context.Context, teamID, userID int64) (string, error) {
	var role string
	if tx, ok := ctx.Value(TxKey).(*sqlx.Tx); ok {
		err := tx.GetContext(ctx, &role,
			"SELECT role FROM team_members WHERE team_id=? AND user_id=?", teamID, userID)
		if err == sql.ErrNoRows {
			return "", nil
		}
		return role, err
	}
	err := tm.db.GetContext(ctx, &role,
		"SELECT role FROM team_members WHERE team_id=? AND user_id=?", teamID, userID)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return role, err
}

// IsAdminOrOwner проверяет, является ли пользователь admin или owner команды.
func (tm *TeamMember) IsAdminOrOwner(ctx context.Context, teamID, userID int64) (bool, error) {
	role, err := tm.Role(ctx, teamID, userID)
	if err != nil {
		return false, err
	}
	return role == "owner" || role == "admin", nil
}

// UpdateRole изменяет роль участника в команде.
func (tm *TeamMember) UpdateRole(ctx context.Context, teamID, userID int64, role string) error {
	if tx, ok := ctx.Value(TxKey).(*sqlx.Tx); ok {
		_, err := tx.ExecContext(ctx,
			"UPDATE team_members SET role=? WHERE team_id=? AND user_id=?",
			role, teamID, userID)
		return err
	}
	_, err := tm.db.ExecContext(ctx,
		"UPDATE team_members SET role=? WHERE team_id=? AND user_id=?",
		role, teamID, userID)
	return err
}
