package service

import (
	"context"

	"github.com/jmoiron/sqlx"
)

// TransactionManager — сквозная инфраструктурная зависимость: используется и Task,
// и Team для атомарных транзакционных операций. Вынесен в отдельный файл, так как
// не привязан к конкретному сервису-потребителю. Единственное место объявления — этот файл.
type TransactionManager interface {
	RunInTransaction(ctx context.Context, fn func(txCtx context.Context, tx *sqlx.Tx) error) error
}
