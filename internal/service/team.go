package service

import (
	"context"
	"time"

	"task-management/internal/apperrors"
	"task-management/internal/models"

	"github.com/jmoiron/sqlx"
)

type TeamRepository interface {
	ByID(ctx context.Context, id int64) (*models.Team, error)
	ByUser(ctx context.Context, userID int64) ([]models.Team, error)
	Create(ctx context.Context, tx *sqlx.Tx, t *models.Team) error
}

// TeamMemberRepository — полный контракт управления участниками команды.
// Объявлен здесь, так как Team — основной потребитель всех пяти методов.
// Task и Stats используют узкие подмножества (см. TeamMembership в task.go и TeamAccess в stats.go).
type TeamMemberRepository interface {
	Add(ctx context.Context, tx *sqlx.Tx, tm *models.TeamMember) error
	Remove(ctx context.Context, tx *sqlx.Tx, teamID, userID int64) error
	Role(ctx context.Context, tx *sqlx.Tx, teamID, userID int64) (string, error)
	IsAdminOrOwner(ctx context.Context, tx *sqlx.Tx, teamID, userID int64) (bool, error)
	UpdateRole(ctx context.Context, tx *sqlx.Tx, teamID, userID int64, role string) error
}

// UserLookup — узкий контракт поиска пользователя по ID, нужный Team
// при проверке существования пользователя в Invite. Объявлен здесь, а не в auth.go,
// так как Auth использует другой набор методов (см. UserRepository в auth.go).
type UserLookup interface {
	ByID(ctx context.Context, id int64) (*models.User, error)
}

type Team struct {
	teamRepo   TeamRepository
	memberRepo TeamMemberRepository
	userRepo   UserLookup
	txManager  TransactionManager
	timeout    time.Duration
}

// NewTeam создаёт экземпляр сервиса команд.
func NewTeam(teamRepo TeamRepository, memberRepo TeamMemberRepository, userRepo UserLookup, txManager TransactionManager, timeout time.Duration) *Team {
	return &Team{
		teamRepo:   teamRepo,
		memberRepo: memberRepo,
		userRepo:   userRepo,
		txManager:  txManager,
		timeout:    timeout,
	}
}

// Create создаёт команду и автоматически добавляет создателя как owner.
// Обе операции (создание команды + добавление участника) выполняются
// в одной транзакции.
func (t *Team) Create(ctx context.Context, name string, userID int64) (*models.Team, error) {
	ctx, cancel := context.WithTimeout(ctx, t.timeout)
	defer cancel()

	var team *models.Team
	if err := t.txManager.RunInTransaction(ctx, func(txCtx context.Context, tx *sqlx.Tx) error {
		team = &models.Team{
			Name:      name,
			CreatedBy: userID,
			CreatedAt: time.Now(),
		}
		if err := t.teamRepo.Create(txCtx, tx, team); err != nil {
			return err
		}
		member := &models.TeamMember{
			TeamID: team.ID,
			UserID: userID,
			Role:   "owner",
		}
		return t.memberRepo.Add(txCtx, tx, member)
	}); err != nil {
		return nil, err
	}
	return team, nil
}

// ListByUser возвращает все команды, в которых состоит пользователь.
func (t *Team) ListByUser(ctx context.Context, userID int64) ([]models.Team, error) {
	ctx, cancel := context.WithTimeout(ctx, t.timeout)
	defer cancel()
	return t.teamRepo.ByUser(ctx, userID)
}

// Invite приглашает пользователя в команду. Проверяет:
//   - приглашающий является admin или owner;
//   - назначаемая роль не «owner» (owner назначается только при создании);
//   - приглашаемый пользователь существует в системе;
//   - приглашаемый ещё не является участником команды.
func (t *Team) Invite(ctx context.Context, teamID, userID, inviterID int64, role string) error {
	ctx, cancel := context.WithTimeout(ctx, t.timeout)
	defer cancel()

	ok, err := t.memberRepo.IsAdminOrOwner(ctx, nil, teamID, inviterID)
	if err != nil {
		return err
	}
	if !ok {
		return apperrors.NewForbidden("permission denied")
	}
	if role == "owner" {
		return apperrors.NewBadRequest("cannot assign owner role via invite")
	}
	u, err := t.userRepo.ByID(ctx, userID)
	if err != nil || u == nil {
		return apperrors.NewNotFound("user not found")
	}
	existingRole, err := t.memberRepo.Role(ctx, nil, teamID, userID)
	if err != nil {
		return err
	}
	if existingRole != "" {
		return apperrors.NewConflict("user already in team")
	}
	member := &models.TeamMember{
		TeamID: teamID,
		UserID: userID,
		Role:   role,
	}
	return t.memberRepo.Add(ctx, nil, member)
}

// ChangeRole изменяет роль участника команды. Проверяет:
//   - изменяющий является admin или owner;
//   - новая роль не «owner»;
//   - целевой пользователь является участником команды;
//   - нельзя изменить роль owner.
func (t *Team) ChangeRole(ctx context.Context, teamID, userID, changerID int64, newRole string) error {
	ctx, cancel := context.WithTimeout(ctx, t.timeout)
	defer cancel()

	ok, err := t.memberRepo.IsAdminOrOwner(ctx, nil, teamID, changerID)
	if err != nil || !ok {
		return apperrors.NewForbidden("permission denied")
	}
	if newRole == "owner" {
		return apperrors.NewBadRequest("cannot set owner role")
	}
	role, err := t.memberRepo.Role(ctx, nil, teamID, userID)
	if err != nil || role == "" {
		return apperrors.NewNotFound("user is not a member")
	}
	if role == "owner" {
		return apperrors.NewBadRequest("cannot change owner's role")
	}
	return t.memberRepo.UpdateRole(ctx, nil, teamID, userID, newRole)
}

// RemoveMember удаляет участника из команды. Проверяет:
//   - удаляющий является admin или owner;
//   - целевой пользователь является участником команды;
//   - нельзя удалить owner.
func (t *Team) RemoveMember(ctx context.Context, teamID, userID, removerID int64) error {
	ctx, cancel := context.WithTimeout(ctx, t.timeout)
	defer cancel()

	ok, err := t.memberRepo.IsAdminOrOwner(ctx, nil, teamID, removerID)
	if err != nil || !ok {
		return apperrors.NewForbidden("permission denied")
	}
	role, err := t.memberRepo.Role(ctx, nil, teamID, userID)
	if err != nil {
		return err
	}
	if role == "" {
		return apperrors.NewNotFound("user not in team")
	}
	if role == "owner" {
		return apperrors.NewBadRequest("cannot remove owner")
	}
	return t.memberRepo.Remove(ctx, nil, teamID, userID)
}
