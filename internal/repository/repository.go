package repository

// txKey — приватный тип для ключа контекста транзакции.
type txKey struct{}

// TxKey — экземпляр ключа для извлечения *sqlx.Tx из контекста.
var TxKey = txKey{}
