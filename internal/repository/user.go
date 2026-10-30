// Package repository реализует доступ к данным для всех доменных сущностей.
// Каждый метод, изменяющий данные, принимает явный параметр tx *sqlx.Tx:
// если он равен nil, запрос выполняется через основное соединение с БД.
package repository

import (
	"context"
	"database/sql"

	"task-management/internal/models"

	"github.com/jmoiron/sqlx"
)

// User реализует CRUD-операции для сущности User в MySQL.
type User struct {
	db *sqlx.DB
}

// NewUser создаёт новый экземпляр репозитория пользователей.
func NewUser(db *sqlx.DB) *User {
	return &User{db: db}
}

// ByID возвращает пользователя по его ID. Если пользователь не найден,
// возвращает (nil, nil) — отсутствие записи не считается ошибкой.
func (u *User) ByID(ctx context.Context, id int64) (*models.User, error) {
	var user models.User
	err := u.db.GetContext(ctx, &user, "SELECT * FROM users WHERE id=?", id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &user, err
}

// ByEmail возвращает пользователя по email. Если запись не найдена,
// возвращает (nil, nil).
func (u *User) ByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := u.db.GetContext(ctx, &user, "SELECT * FROM users WHERE email=?", email)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &user, err
}

// Create вставляет нового пользователя в БД. Если tx != nil, запрос
// выполняется в рамках переданной транзакции; иначе — напрямую через
// соединение. После вставки записывает ID в user.ID.
func (u *User) Create(ctx context.Context, tx *sqlx.Tx, user *models.User) error {
	res, err := executor(u.db, tx).ExecContext(ctx,
		"INSERT INTO users (email, password_hash, name, created_at) VALUES (?, ?, ?, ?)",
		user.Email, user.PasswordHash, user.Name, user.CreatedAt)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	user.ID = id
	return nil
}
