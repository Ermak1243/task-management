package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"task-management/internal/mocks"
	"task-management/internal/models"
	"task-management/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newStatsTestEnv(t *testing.T) (*service.Stats, *mocks.MockStatsRepository, *mocks.MockTeamAccess) {
	statsRepo := mocks.NewMockStatsRepository(t)
	memberRepo := mocks.NewMockTeamAccess(t)
	svc := service.NewStats(statsRepo, memberRepo, time.Second)
	return svc, statsRepo, memberRepo
}

func TestStats_GetTeamStats(t *testing.T) {
	t.Run("access denied", func(t *testing.T) {
		svc, statsRepo, memberRepo := newStatsTestEnv(t)
		memberRepo.EXPECT().IsAdminOrOwner(mock.Anything, int64(1), int64(42)).Return(false, nil)

		stats, err := svc.GetTeamStats(context.Background(), 1, 42)

		require.EqualError(t, err, "access denied")
		assert.Nil(t, stats)
		statsRepo.AssertNotCalled(t, "TeamStats", mock.Anything, int64(1))
		memberRepo.AssertExpectations(t)
	})

	t.Run("isAdminOrOwner error propagates", func(t *testing.T) {
		svc, statsRepo, memberRepo := newStatsTestEnv(t)
		repoErr := errors.New("role lookup failed")
		memberRepo.EXPECT().IsAdminOrOwner(mock.Anything, int64(1), int64(42)).Return(false, repoErr)

		stats, err := svc.GetTeamStats(context.Background(), 1, 42)

		require.ErrorIs(t, err, repoErr)
		assert.Nil(t, stats)
		statsRepo.AssertNotCalled(t, "TeamStats", mock.Anything, int64(1))
		memberRepo.AssertExpectations(t)
	})

	t.Run("success returns repository result", func(t *testing.T) {
		svc, statsRepo, memberRepo := newStatsTestEnv(t)
		memberRepo.EXPECT().IsAdminOrOwner(mock.Anything, int64(1), int64(42)).Return(true, nil)
		expected := &models.TeamStats{
			Todo:          1,
			InProgress:    2,
			Done:          3,
			AvgCloseTime:  120.5,
			TopAssignees:  []int64{10, 20, 30},
			TotalComments: 7,
		}
		statsRepo.EXPECT().TeamStats(mock.Anything, int64(1)).Return(expected, nil)

		stats, err := svc.GetTeamStats(context.Background(), 1, 42)

		require.NoError(t, err)
		assert.Equal(t, expected, stats)
		memberRepo.AssertExpectations(t)
	})

	t.Run("repository error propagates", func(t *testing.T) {
		svc, statsRepo, memberRepo := newStatsTestEnv(t)
		memberRepo.EXPECT().IsAdminOrOwner(mock.Anything, int64(1), int64(42)).Return(true, nil)
		repoErr := errors.New("stats query failed")
		statsRepo.EXPECT().TeamStats(mock.Anything, int64(1)).Return(nil, repoErr)

		stats, err := svc.GetTeamStats(context.Background(), 1, 42)

		require.ErrorIs(t, err, repoErr)
		assert.Nil(t, stats)
		memberRepo.AssertExpectations(t)
	})
}
