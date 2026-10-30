package service_test

import (
	"context"
	"testing"
	"time"

	"task-management/internal/models"
	"task-management/internal/service"
	"task-management/internal/service/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAuth_Register_EmailAlreadyRegistered(t *testing.T) {
	userRepo := mocks.NewMockUserRepository(t)
	svc := service.NewAuth(userRepo, "secret", time.Hour, time.Second)
	userRepo.EXPECT().
		ByEmail(mock.Anything, "user@test.com").
		Return(&models.User{ID: 1, Email: "user@test.com"}, nil)
	user, err := svc.Register(context.Background(), "user@test.com", "password", "User")
	appErr := requireAppError(t, err, 409)
	assert.Equal(t, "email already registered", appErr.Message)
	assert.Nil(t, user)
}

func TestAuth_Login_InvalidCredentials(t *testing.T) {
	userRepo := mocks.NewMockUserRepository(t)
	svc := service.NewAuth(userRepo, "secret", time.Hour, time.Second)
	userRepo.EXPECT().
		ByEmail(mock.Anything, "user@test.com").
		Return(nil, nil)
	token, err := svc.Login(context.Background(), "user@test.com", "password")
	appErr := requireAppError(t, err, 401)
	assert.Equal(t, "invalid credentials", appErr.Message)
	assert.Empty(t, token)
}
