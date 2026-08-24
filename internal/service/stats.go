package service

import (
	"context"
	"errors"
	"time"

	"task-management/internal/models"
)

// StatsRepository — контракт получения агрегированной статистики по команде.
// SQL-логика инкапсулирована в internal/repository/stats.go.
type StatsRepository interface {
	TeamStats(ctx context.Context, teamID int64) (*models.TeamStats, error)
}

// TeamAccess — минимальный контракт проверки прав администратора/владельца команды,
// который нужен сервису Stats. Объявлен здесь, так как Stats — единственный
// потребитель подмножества IsAdminOrOwner из TeamMemberRepository.
type TeamAccess interface {
	IsAdminOrOwner(ctx context.Context, teamID, userID int64) (bool, error)
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
// Доступ разрешён только admin и owner — обычные участники получают ошибку
// «access denied». Статистика включает количество задач по статусам,
// среднее время закрытия, топ-3 исполнителей и число комментариев.
func (s *Stats) GetTeamStats(ctx context.Context, teamID int64, userID int64) (*models.TeamStats, error) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	ok, err := s.memberRepo.IsAdminOrOwner(ctx, teamID, userID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New("access denied")
	}
	return s.statsRepo.TeamStats(ctx, teamID)
}