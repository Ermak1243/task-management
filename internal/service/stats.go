package service

import (
	"context"
	"time"

	"task-management/internal/apperrors"
	"task-management/internal/models"

	"github.com/jmoiron/sqlx"
)

// StatsRepository — контракт получения агрегированной статистики по команде.
type StatsRepository interface {
	TeamStats(ctx context.Context, teamID int64) (*models.TeamStats, error)
}

// TeamAccess — минимальный контракт проверки прав администратора/владельца команды.
type TeamAccess interface {
	IsAdminOrOwner(ctx context.Context, tx *sqlx.Tx, teamID, userID int64) (bool, error)
}

type Stats struct {
	statsRepo  StatsRepository
	memberRepo TeamAccess
	timeout    time.Duration
}

// NewStats создаёт экземпляр сервиса статистики.
func NewStats(statsRepo StatsRepository, memberRepo TeamAccess, timeout time.Duration) *Stats {
	return &Stats{
		statsRepo:  statsRepo,
		memberRepo: memberRepo,
		timeout:    timeout,
	}
}

// GetTeamStats возвращает агрегированную статистику по команде.
func (s *Stats) GetTeamStats(ctx context.Context, teamID int64, userID int64) (*models.TeamStats, error) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	ok, err := s.memberRepo.IsAdminOrOwner(ctx, nil, teamID, userID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, apperrors.NewForbidden("access denied")
	}
	return s.statsRepo.TeamStats(ctx, teamID)
}
