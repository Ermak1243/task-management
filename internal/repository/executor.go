package repository

import "github.com/jmoiron/sqlx"

// executor возвращает источник выполнения запросов: явно переданную
// транзакцию, если она есть, иначе — основное соединение с БД.
// Это позволяет вызывать методы репозиториев как внутри транзакции,
// так и вне её, без передачи зависимостей через context.
func executor(db *sqlx.DB, tx *sqlx.Tx) sqlx.ExtContext {
	if tx != nil {
		return tx
	}
	return db
}
