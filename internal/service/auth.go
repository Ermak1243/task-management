package service

import (
	"context"
	"time"

	"task-management/internal/apperrors"
	"task-management/internal/models"
	"task-management/pkg/jwt"

	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

// UserRepository — контракт доступа к учётным записям пользователей для сервиса Auth.
// Содержит только методы, реально используемые аутентификацией (ByEmail, Create).
// Team использует другой набор методов (см. UserLookup в team.go), поэтому общий
// трёхметодный интерфейс был разбит на два узких.
type UserRepository interface {
	ByEmail(ctx context.Context, email string) (*models.User, error)
	Create(ctx context.Context, tx *sqlx.Tx, u *models.User) error
}

type Auth struct {
	userRepo UserRepository
	jwtCfg   jwt.Config
	timeout  time.Duration
}

// NewAuth создаёт экземпляр сервиса аутентификации.
func NewAuth(userRepo UserRepository, jwtSecret string, jwtExpiry, timeout time.Duration) *Auth {
	return &Auth{
		userRepo: userRepo,
		jwtCfg:   jwt.Config{Secret: []byte(jwtSecret), Expiry: jwtExpiry},
		timeout:  timeout,
	}
}

// Register создаёт нового пользователя. Проверяет уникальность email,
// хеширует пароль через bcrypt и сохраняет в БД. Возвращает созданного
// пользователя или ошибку, если email уже занят.
func (a *Auth) Register(ctx context.Context, email, password, name string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(ctx, a.timeout)
	defer cancel()

	existing, err := a.userRepo.ByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, apperrors.NewConflict("email already registered")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	u := &models.User{
		Email:        email,
		PasswordHash: string(hash),
		Name:         name,
		CreatedAt:    time.Now(),
	}
	if err := a.userRepo.Create(ctx, nil, u); err != nil {
		return nil, err
	}
	return u, nil
}

// Login проверяет учётные данные пользователя (email + пароль) и возвращает
// подписанный JWT-токен. При неверных данных возвращает ошибку
// «invalid credentials» без уточнения, что именно неверно.
func (a *Auth) Login(ctx context.Context, email, password string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, a.timeout)
	defer cancel()

	u, err := a.userRepo.ByEmail(ctx, email)
	if err != nil || u == nil {
		return "", apperrors.NewUnauthorized("invalid credentials")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return "", apperrors.NewUnauthorized("invalid credentials")
	}
	token, err := jwt.Generate(a.jwtCfg, u.ID)
	if err != nil {
		return "", err
	}
	return token, nil
}
