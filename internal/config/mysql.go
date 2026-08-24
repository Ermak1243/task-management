package config

// MySQL хранит параметры подключения к базе данных MySQL.
type MySQL struct {
	Host     string `env:"MYSQL_HOST" env-required:"true"`
	Port     string `env:"MYSQL_PORT" env-default:"3306"`
	User     string `env:"MYSQL_USER" env-required:"true"`
	Password string `env:"MYSQL_PASSWORD" env-required:"true"`
	DBName   string `env:"MYSQL_DB" env-required:"true"`
	SSLMode  string `env:"MYSQL_SSLMODE" env-default:"false"`
}
