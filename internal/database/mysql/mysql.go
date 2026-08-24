// Package mysql предоставляет обёртку над соединением с MySQL
// и менеджер транзакций для repository-слоя.
package mysql

import (
	"fmt"
	"log"

	"task-management/internal/config"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

// DB — обёртка над *sqlx.DB для управления жизненным циклом соединения.
type DB struct {
	db *sqlx.DB
}

// New создаёт соединение с MySQL, собирая DSN из параметров конфигурации.
// Panic, если соединение установить не удалось.
func New(cfg config.MySQL) *DB {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&tls=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.SSLMode)
	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		panic(err)
	}
	log.Println("Connected to MySQL")
	return &DB{db: db}
}

// SQLX возвращает底层 *sqlx.DB для выполнения запросов.
func (d *DB) SQLX() *sqlx.DB { return d.db }

// Close закрывает соединение с MySQL.
func (d *DB) Close() {
	if err := d.db.Close(); err != nil {
		log.Println(err)
	}
	log.Println("MySQL connection closed")
}
