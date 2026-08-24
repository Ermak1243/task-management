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

func runTxInline(txManager *mocks.MockTransactionManager) {
	txManager.EXPECT().
		RunInTransaction(mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})
}

func newTeamTestEnv(t *testing.T) (*service.Team, *mocks.MockTeamRepository, *mocks.MockTeamMemberRepository, *mocks.MockUserLookup, *mocks.MockTransactionManager) {
	teamRepo := mocks.NewMockTeamRepository(t)
	memberRepo := mocks.NewMockTeamMemberRepository(t)
	userRepo := mocks.NewMockUserLookup(t)
	txManager := mocks.NewMockTransactionManager(t)
	svc := service.NewTeam(teamRepo, memberRepo, userRepo, txManager, time.Second)
	return svc, teamRepo, memberRepo, userRepo, txManager
}

func TestTeam_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, teamRepo, memberRepo, _, txManager := newTeamTestEnv(t)
		runTxInline(txManager)

		teamRepo.EXPECT().
			Create(mock.Anything, mock.AnythingOfType("*models.Team")).
			Run(func(_ context.Context, team *models.Team) { team.ID = 1 }).
			Return(nil)
		memberRepo.EXPECT().
			Add(mock.Anything, mock.AnythingOfType("*models.TeamMember")).
			Run(func(_ context.Context, tm *models.TeamMember) {
				assert.Equal(t, int64(1), tm.TeamID)
				assert.Equal(t, int64(42), tm.UserID)
				assert.Equal(t, "owner", tm.Role)
			}).
			Return(nil)

		team, err := svc.Create(context.Background(), "Team A", 42)
		require.NoError(t, err)
		assert.Equal(t, int64(1), team.ID)
		assert.Equal(t, "Team A", team.Name)
		assert.Equal(t, int64(42), team.CreatedBy)
	})

	t.Run("rollback on repo error", func(t *testing.T) {
		svc, teamRepo, _, _, txManager := newTeamTestEnv(t)
		runTxInline(txManager)

		repoErr := errors.New("db error")
		teamRepo.EXPECT().
			Create(mock.Anything, mock.AnythingOfType("*models.Team")).
			Return(repoErr)

		team, err := svc.Create(context.Background(), "Team A", 42)
		require.ErrorIs(t, err, repoErr)
		assert.Nil(t, team)
	})

	t.Run("rollback on member add error", func(t *testing.T) {
		svc, teamRepo, memberRepo, _, txManager := newTeamTestEnv(t)
		runTxInline(txManager)

		teamRepo.EXPECT().
			Create(mock.Anything, mock.AnythingOfType("*models.Team")).
			Run(func(_ context.Context, team *models.Team) { team.ID = 1 }).
			Return(nil)
		repoErr := errors.New("member insert failed")
		memberRepo.EXPECT().
			Add(mock.Anything, mock.AnythingOfType("*models.TeamMember")).
			Return(repoErr)

		team, err := svc.Create(context.Background(), "Team A", 42)
		require.ErrorIs(t, err, repoErr)
		assert.Nil(t, team)
	})
}

func TestTeam_Invite(t *testing.T) {
	u := &models.User{ID: 7, Email: "user@test.com", Name: "User"}

	tests := []struct {
		name    string
		role    string
		setup   func(memberRepo *mocks.MockTeamMemberRepository, userRepo *mocks.MockUserLookup)
		wantErr string
	}{
		{
			name: "success",
			setup: func(memberRepo *mocks.MockTeamMemberRepository, userRepo *mocks.MockUserLookup) {
				memberRepo.EXPECT().IsAdminOrOwner(mock.Anything, int64(1), int64(2)).Return(true, nil)
				userRepo.EXPECT().ByID(mock.Anything, int64(7)).Return(u, nil)
				memberRepo.EXPECT().Role(mock.Anything, int64(1), int64(7)).Return("", nil)
				memberRepo.EXPECT().
					Add(mock.Anything, mock.AnythingOfType("*models.TeamMember")).
					Run(func(_ context.Context, tm *models.TeamMember) {
						assert.Equal(t, int64(7), tm.UserID)
						assert.Equal(t, "member", tm.Role)
					}).
					Return(nil)
			},
		},
		{
			name:    "permission denied",
			wantErr: "permission denied",
			setup: func(memberRepo *mocks.MockTeamMemberRepository, userRepo *mocks.MockUserLookup) {
				memberRepo.EXPECT().IsAdminOrOwner(mock.Anything, int64(1), int64(2)).Return(false, nil)
			},
		},
		{
			name:    "cannot assign owner role",
			role:    "owner",
			wantErr: "cannot assign owner role via invite",
			setup: func(memberRepo *mocks.MockTeamMemberRepository, userRepo *mocks.MockUserLookup) {
				memberRepo.EXPECT().IsAdminOrOwner(mock.Anything, int64(1), int64(2)).Return(true, nil)
			},
		},
		{
			name:    "user not found",
			wantErr: "user not found",
			setup: func(memberRepo *mocks.MockTeamMemberRepository, userRepo *mocks.MockUserLookup) {
				memberRepo.EXPECT().IsAdminOrOwner(mock.Anything, int64(1), int64(2)).Return(true, nil)
				userRepo.EXPECT().ByID(mock.Anything, int64(7)).Return(nil, nil)
			},
		},
		{
			name:    "user already in team",
			wantErr: "user already in team",
			setup: func(memberRepo *mocks.MockTeamMemberRepository, userRepo *mocks.MockUserLookup) {
				memberRepo.EXPECT().IsAdminOrOwner(mock.Anything, int64(1), int64(2)).Return(true, nil)
				userRepo.EXPECT().ByID(mock.Anything, int64(7)).Return(u, nil)
				memberRepo.EXPECT().Role(mock.Anything, int64(1), int64(7)).Return("member", nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, _, memberRepo, userRepo, _ := newTeamTestEnv(t)
			tt.setup(memberRepo, userRepo)

			role := tt.role
			if role == "" {
				role = "member"
			}
			err := svc.Invite(context.Background(), 1, 7, 2, role)
			if tt.wantErr != "" {
				require.EqualError(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestTeam_ChangeRole_PermissionDenied(t *testing.T) {
	svc, _, memberRepo, _, _ := newTeamTestEnv(t)
	memberRepo.EXPECT().IsAdminOrOwner(mock.Anything, int64(1), int64(2)).Return(false, nil)

	err := svc.ChangeRole(context.Background(), 1, 7, 2, "admin")
	require.EqualError(t, err, "permission denied")
}

func TestTeam_RemoveMember_PermissionDenied(t *testing.T) {
	svc, _, memberRepo, _, _ := newTeamTestEnv(t)
	memberRepo.EXPECT().IsAdminOrOwner(mock.Anything, int64(1), int64(2)).Return(false, nil)

	err := svc.RemoveMember(context.Background(), 1, 7, 2)
	require.EqualError(t, err, "permission denied")
}
