// Package models определяет все доменные сущности и DTO-структуры
// запросов/ответов, используемые во всех слоях приложения.
package models

import "time"

// User — учётная запись пользователя приложения.
type User struct {
	ID           int64     `json:"id" db:"id"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"`
	Name         string    `json:"name" db:"name"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// Team — именованная группа пользователей, совместно работающая над задачами.
type Team struct {
	ID        int64     `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	CreatedBy int64     `json:"created_by" db:"created_by"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// TeamMember связывает пользователя с командой и задаёт его роль
// (owner, admin или member).
type TeamMember struct {
	TeamID int64  `json:"team_id" db:"team_id"`
	UserID int64  `json:"user_id" db:"user_id"`
	Role   string `json:"role" db:"role"` // "owner", "admin", "member"
}

// Task — рабочий элемент внутри команды. Поле version используется для
// оптимистичного контроля конкурентного доступа: обновление не удаётся,
// если версия уже изменилась.
type Task struct {
	ID          int64      `json:"id" db:"id"`
	TeamID      int64      `json:"team_id" db:"team_id"`
	Title       string     `json:"title" db:"title"`
	Description string     `json:"description" db:"description"`
	Status      string     `json:"status" db:"status"` // todo, in_progress, done
	CreatedBy   int64      `json:"created_by" db:"created_by"`
	AssigneeID  *int64     `json:"assignee_id,omitempty" db:"assignee_id"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	ClosedAt    *time.Time `json:"closed_at,omitempty" db:"closed_at"`
	Version     int        `json:"version" db:"version"`
}

// TaskHistory — запись аудит-журнала изменений задачи. Поле Changes содержит
// JSON-представление дифа изменённых полей.
type TaskHistory struct {
	ID        int64     `json:"id" db:"id"`
	TaskID    int64     `json:"task_id" db:"task_id"`
	ChangedBy int64     `json:"changed_by" db:"changed_by"`
	Changes   string    `json:"changes" db:"changes"` // JSON
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// TaskComment — комментарий пользователя к задаче.
type TaskComment struct {
	ID        int64     `json:"id" db:"id"`
	TaskID    int64     `json:"task_id" db:"task_id"`
	UserID    int64     `json:"user_id" db:"user_id"`
	Content   string    `json:"content" db:"content"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// ---------- DTO-структуры запросов и ответов ----------

// RegisterRequest — тело запроса на регистрацию пользователя.
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	Name     string `json:"name" validate:"required"`
}

// LoginRequest — тело запроса на аутентификацию пользователя.
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// CreateTeamRequest — тело запроса на создание новой команды.
type CreateTeamRequest struct {
	Name string `json:"name" validate:"required"`
}

// InviteRequest — тело запроса на приглашение пользователя в команду.
type InviteRequest struct {
	UserID int64  `json:"user_id" validate:"required,gt=0"`
	Role   string `json:"role" validate:"oneof=admin member"`
}

// UpdateRoleRequest — тело запроса на изменение роли участника команды.
type UpdateRoleRequest struct {
	UserID int64  `json:"user_id" validate:"required,gt=0"`
	Role   string `json:"role" validate:"oneof=admin member"`
}

// CreateTaskRequest — тело запроса на создание новой задачи.
type CreateTaskRequest struct {
	TeamID      int64  `json:"team_id" validate:"required,gt=0"`
	Title       string `json:"title" validate:"required"`
	Description string `json:"description"`
	AssigneeID  *int64 `json:"assignee_id,omitempty"`
}

// UpdateTaskRequest — тело запроса на частичное обновление задачи.
// Все поля опциональны: nil означает «без изменений».
type UpdateTaskRequest struct {
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
	Status      *string `json:"status,omitempty" validate:"omitempty,oneof=todo in_progress done"`
	AssigneeID  *int64  `json:"assignee_id,omitempty"` // nil = no change, 0 = set to NULL, >0 = set to user
}

// CommentRequest — тело запроса на добавление комментария к задаче.
type CommentRequest struct {
	Content string `json:"content" validate:"required"`
}

// TeamStats — агрегированная статистика по команде, формируемая одним
// SQL-запросом с CTE в репозитории статистики.
type TeamStats struct {
	Todo          int     `json:"todo"`
	InProgress    int     `json:"in_progress"`
	Done          int     `json:"done"`
	AvgCloseTime  float64 `json:"avg_close_time_seconds"`
	TopAssignees  []int64 `json:"top_assignees"`
	TotalComments int     `json:"total_comments"`
}
