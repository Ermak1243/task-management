// Package apperrors содержит типизированные ошибки приложения.
package apperrors

// AppError — типизированная ошибка приложения с HTTP-статусом.
type AppError struct {
	Status  int
	Code    string
	Message string
}

// Error реализует интерфейс error.
func (e *AppError) Error() string { return e.Message }

// NewBadRequest создаёт ошибку с HTTP-статусом 400.
func NewBadRequest(msg string) *AppError { return &AppError{400, "bad_request", msg} }

// NewUnauthorized создаёт ошибку с HTTP-статусом 401.
func NewUnauthorized(msg string) *AppError { return &AppError{401, "unauthorized", msg} }

// NewForbidden создаёт ошибку с HTTP-статусом 403.
func NewForbidden(msg string) *AppError { return &AppError{403, "forbidden", msg} }

// NewNotFound создаёт ошибку с HTTP-статусом 404.
func NewNotFound(msg string) *AppError { return &AppError{404, "not_found", msg} }

// NewConflict создаёт ошибку с HTTP-статусом 409.
func NewConflict(msg string) *AppError { return &AppError{409, "conflict", msg} }
