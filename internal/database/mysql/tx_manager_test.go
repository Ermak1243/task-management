package mysql

import (
	"context"
	"errors"
	"testing"

	"task-management/internal/repository"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTxManager_Commit(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mgr := NewTxManager(sqlx.NewDb(db, "mysql"))

	mock.ExpectBegin()
	mock.ExpectCommit()

	err = mgr.RunInTransaction(context.Background(), func(ctx context.Context) error {
		return nil
	})
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTxManager_Rollback(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mgr := NewTxManager(sqlx.NewDb(db, "mysql"))

	repoErr := errors.New("db error")
	mock.ExpectBegin()
	mock.ExpectRollback()

	err = mgr.RunInTransaction(context.Background(), func(ctx context.Context) error {
		return repoErr
	})
	require.ErrorIs(t, err, repoErr)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTxManager_CommitFailure(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mgr := NewTxManager(sqlx.NewDb(db, "mysql"))

	commitErr := errors.New("commit failed")
	mock.ExpectBegin()
	mock.ExpectCommit().WillReturnError(commitErr)

	err = mgr.RunInTransaction(context.Background(), func(ctx context.Context) error {
		return nil
	})
	require.ErrorIs(t, err, commitErr)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTxManager_TxKeyInContext(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mgr := NewTxManager(sqlx.NewDb(db, "mysql"))

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO teams").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err = mgr.RunInTransaction(context.Background(), func(ctx context.Context) error {
		tx, ok := ctx.Value(repository.TxKey).(*sqlx.Tx)
		assert.True(t, ok)
		assert.NotNil(t, tx)
		_, err := tx.ExecContext(ctx, "INSERT INTO teams (name) VALUES (?)", "Team")
		return err
	})
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}