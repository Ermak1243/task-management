package service_test

import (
	"context"
	"encoding/json"
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

func strPtr(s string) *string { return &s }
func intPtr(i int64) *int64   { return &i }

func newTaskTestEnv(t *testing.T) (*service.Task, *mocks.MockTaskRepository, *mocks.MockTaskHistoryRepository, *mocks.MockTeamMembership, *mocks.MockTaskCache, *mocks.MockTransactionManager) {
	taskRepo := mocks.NewMockTaskRepository(t)
	historyRepo := mocks.NewMockTaskHistoryRepository(t)
	commentRepo := mocks.NewMockTaskCommentRepository(t)
	memberRepo := mocks.NewMockTeamMembership(t)
	cache := mocks.NewMockTaskCache(t)
	txManager := mocks.NewMockTransactionManager(t)
	svc := service.NewTask(taskRepo, historyRepo, commentRepo, memberRepo, cache, txManager, time.Second)
	return svc, taskRepo, historyRepo, memberRepo, cache, txManager
}

func TestTask_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, taskRepo, historyRepo, memberRepo, cache, txManager := newTaskTestEnv(t)

		memberRepo.EXPECT().Role(mock.Anything, int64(1), int64(42)).Return("member", nil)
		runTxInline(txManager)
		taskRepo.EXPECT().
			Create(mock.Anything, mock.AnythingOfType("*models.Task")).
			Run(func(_ context.Context, task *models.Task) { task.ID = 9 }).
			Return(nil)
		historyRepo.EXPECT().Create(mock.Anything, mock.AnythingOfType("*models.TaskHistory")).Return(nil)
		cache.EXPECT().DelByPattern(mock.Anything, "tasks:team:1:*").Return(nil)

		req := &models.Task{TeamID: 1, Title: "T", Description: "D"}
		task, err := svc.Create(context.Background(), req, 42)
		require.NoError(t, err)
		assert.Equal(t, int64(9), task.ID)
		assert.Equal(t, "todo", task.Status)
		assert.Equal(t, int64(42), task.CreatedBy)
	})

	t.Run("rollback when task create fails", func(t *testing.T) {
		svc, taskRepo, _, memberRepo, _, txManager := newTaskTestEnv(t)

		memberRepo.EXPECT().Role(mock.Anything, int64(1), int64(42)).Return("member", nil)
		runTxInline(txManager)
		repoErr := errors.New("task insert failed")
		taskRepo.EXPECT().
			Create(mock.Anything, mock.AnythingOfType("*models.Task")).
			Return(repoErr)

		task, err := svc.Create(context.Background(), &models.Task{TeamID: 1}, 42)
		require.ErrorIs(t, err, repoErr)
		assert.Nil(t, task)
	})

	t.Run("rollback when history create fails", func(t *testing.T) {
		svc, taskRepo, historyRepo, memberRepo, _, txManager := newTaskTestEnv(t)

		memberRepo.EXPECT().Role(mock.Anything, int64(1), int64(42)).Return("member", nil)
		runTxInline(txManager)
		taskRepo.EXPECT().
			Create(mock.Anything, mock.AnythingOfType("*models.Task")).
			Run(func(_ context.Context, task *models.Task) { task.ID = 9 }).
			Return(nil)
		repoErr := errors.New("history insert failed")
		historyRepo.EXPECT().Create(mock.Anything, mock.AnythingOfType("*models.TaskHistory")).Return(repoErr)

		task, err := svc.Create(context.Background(), &models.Task{TeamID: 1}, 42)
		require.ErrorIs(t, err, repoErr)
		assert.Nil(t, task)
	})

	t.Run("not a member of team", func(t *testing.T) {
		svc, _, _, memberRepo, _, _ := newTaskTestEnv(t)

		memberRepo.EXPECT().Role(mock.Anything, int64(1), int64(42)).Return("", nil)

		task, err := svc.Create(context.Background(), &models.Task{TeamID: 1}, 42)
		require.EqualError(t, err, "user not member of team")
		assert.Nil(t, task)
	})
}

func TestTask_Update(t *testing.T) {
	closedAt := time.Now().Add(-time.Hour)
	tests := []struct {
		name            string
		existing        *models.Task
		req             *models.UpdateTaskRequest
		userID          int64
		isAdmin         bool
		wantErr         string
		wantStatus      string
		wantClosedAtNil bool
	}{
		{
			name:       "success as creator",
			existing:   &models.Task{ID: 1, TeamID: 1, Title: "old", Description: "d", Status: "todo", CreatedBy: 2, Version: 1},
			req:        &models.UpdateTaskRequest{Status: strPtr("done")},
			userID:     2,
			isAdmin:    true,
			wantStatus: "done",
		},
		{
			name:            "assignee can reopen done task",
			existing:        &models.Task{ID: 1, TeamID: 1, Title: "old", Description: "d", Status: "done", CreatedBy: 2, AssigneeID: intPtr(3), ClosedAt: &closedAt, Version: 1},
			req:             &models.UpdateTaskRequest{Status: strPtr("in_progress")},
			userID:          3,
			isAdmin:         false,
			wantStatus:      "in_progress",
			wantClosedAtNil: true,
		},
		{
			name:     "non member cannot update",
			existing: &models.Task{ID: 1, TeamID: 1, Title: "old", CreatedBy: 2},
			req:      &models.UpdateTaskRequest{},
			userID:   3,
			isAdmin:  false,
			wantErr:  "permission denied",
		},
		{
			name:     "assignee can only change status",
			existing: &models.Task{ID: 1, TeamID: 1, Title: "old", CreatedBy: 2, AssigneeID: intPtr(3)},
			req:      &models.UpdateTaskRequest{Title: strPtr("x")},
			userID:   3,
			isAdmin:  false,
			wantErr:  "as assignee you can only change status",
		},
		{
			name:     "task not found",
			existing: nil,
			req:      &models.UpdateTaskRequest{},
			userID:   3,
			wantErr:  "task not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, taskRepo, historyRepo, memberRepo, cache, txManager := newTaskTestEnv(t)

			if tt.existing == nil {
				taskRepo.EXPECT().ByID(mock.Anything, int64(1)).Return(nil, nil)
			} else {
				taskRepo.EXPECT().ByID(mock.Anything, tt.existing.ID).Return(tt.existing, nil)
				memberRepo.EXPECT().IsAdminOrOwner(mock.Anything, tt.existing.TeamID, tt.userID).Return(tt.isAdmin, nil)
			}

			if tt.wantErr == "" {
				runTxInline(txManager)
				taskRepo.EXPECT().Update(mock.Anything, mock.AnythingOfType("*models.Task")).Return(nil)
				historyRepo.EXPECT().Create(mock.Anything, mock.AnythingOfType("*models.TaskHistory")).Return(nil)
				cache.EXPECT().DelByPattern(mock.Anything, "tasks:team:1:*").Return(nil)
			}

			task, err := svc.Update(context.Background(), 1, tt.req, tt.userID)
			if tt.wantErr != "" {
				require.EqualError(t, err, tt.wantErr)
				assert.Nil(t, task)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, task.Status)
			if tt.wantClosedAtNil {
				assert.Nil(t, task.ClosedAt)
			} else if tt.wantStatus == "done" {
				assert.NotNil(t, task.ClosedAt)
			}
		})
	}
}

func TestTask_Update_Rollback(t *testing.T) {
	svc, taskRepo, _, memberRepo, _, txManager := newTaskTestEnv(t)

	existing := &models.Task{ID: 1, TeamID: 1, Title: "old", Status: "todo", CreatedBy: 2, Version: 1}
	taskRepo.EXPECT().ByID(mock.Anything, int64(1)).Return(existing, nil)
	memberRepo.EXPECT().IsAdminOrOwner(mock.Anything, int64(1), int64(2)).Return(true, nil)
	runTxInline(txManager)
	repoErr := errors.New("update failed")
	taskRepo.EXPECT().Update(mock.Anything, mock.AnythingOfType("*models.Task")).Return(repoErr)

	task, err := svc.Update(context.Background(), 1, &models.UpdateTaskRequest{Title: strPtr("new")}, 2)
	require.ErrorIs(t, err, repoErr)
	assert.Nil(t, task)
}

func TestTask_ListByTeam(t *testing.T) {
	t.Run("not a member", func(t *testing.T) {
		svc, _, _, memberRepo, _, _ := newTaskTestEnv(t)
		memberRepo.EXPECT().Role(mock.Anything, int64(1), int64(5)).Return("", nil)

		tasks, err := svc.ListByTeam(context.Background(), 1, 5, nil, nil, 0, 10, nil)
		require.EqualError(t, err, "user not member of team")
		assert.Nil(t, tasks)
	})

	t.Run("served from cache", func(t *testing.T) {
		svc, _, _, memberRepo, cache, _ := newTaskTestEnv(t)
		memberRepo.EXPECT().Role(mock.Anything, int64(1), int64(5)).Return("member", nil)
		expected := []models.Task{{ID: 1, Title: "cached"}}
		data, _ := json.Marshal(expected)
		cache.EXPECT().Get(mock.Anything, mock.AnythingOfType("string")).Return(string(data), nil)

		tasks, err := svc.ListByTeam(context.Background(), 1, 5, nil, nil, 0, 10, nil)
		require.NoError(t, err)
		assert.Equal(t, expected, tasks)
	})

	t.Run("cache miss queries db and stores", func(t *testing.T) {
		svc, taskRepo, _, memberRepo, cache, _ := newTaskTestEnv(t)
		memberRepo.EXPECT().Role(mock.Anything, int64(1), int64(5)).Return("member", nil)
		cache.EXPECT().Get(mock.Anything, mock.AnythingOfType("string")).Return("", errors.New("redis: nil"))
		expected := []models.Task{{ID: 1, Title: "db"}}
		taskRepo.EXPECT().
			ListByTeam(mock.Anything, int64(1), (*string)(nil), (*int64)(nil), 0, 10, (*int64)(nil)).
			Return(expected, nil)
		cache.EXPECT().
			Set(mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("[]uint8"), mock.AnythingOfType("time.Duration")).
			Return(nil)

		tasks, err := svc.ListByTeam(context.Background(), 1, 5, nil, nil, 0, 10, nil)
		require.NoError(t, err)
		assert.Equal(t, expected, tasks)
	})
}
