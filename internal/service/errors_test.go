package service_test

import (
	"testing"

	"task-management/internal/apperrors"

	"github.com/stretchr/testify/require"
)

// requireAppError проверяет, что ошибка является *apperrors.AppError
// с ожидаемым HTTP-статусом, и возвращает её для дальнейших проверок.
func requireAppError(t *testing.T, err error, status int) *apperrors.AppError {
	t.Helper()
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, status, appErr.Status)
	return appErr
}
