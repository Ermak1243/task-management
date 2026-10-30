package mysql

import (
	"context"

	"github.com/jmoiron/sqlx"
)

// TxManager реализует интерфейс service.TransactionManager. Каждый вызов
// RunInTransaction открывает транзакцию, передаёт её явным параметром
// в функцию fn и коммитит при успехе или откатывает при ошибке/панике.
type TxManager struct {
	db *sqlx.DB
}

// NewTxManager создаёт менеджер транзакций на основе переданного *sqlx.DB.
func NewTxManager(db *sqlx.DB) *TxManager {
	return &TxManager{db: db}
}

// RunInTransaction выполняет fn внутри транзакции. При панике внутри fn
// транзакция откатывается и паника пробрасывается дальше.
func (tm *TxManager) RunInTransaction(ctx context.Context, fn func(txCtx context.Context, tx *sqlx.Tx) error) error {
	tx, err := tm.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()
	if err := fn(ctx, tx); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}
