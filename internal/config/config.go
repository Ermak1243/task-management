// Package config содержит строго типизированные структуры конфигурации,
// загружаемые из переменных окружения или .env-файла через cleanenv.
package config

import (
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

// Config хранит все параметры конфигурации приложения.
type Config struct {
	ServerPort     string        `env:"SERVER_PORT" env-default:"8080"`
	ContextTimeout time.Duration `env:"CONTEXT_TIMEOUT" env-default:"30s"`
	JWTSecret      string        `env:"JWT_SECRET" env-required:"true"`
	JWTExpiry      time.Duration `env:"JWT_EXPIRY" env-default:"24h"`
	MySQL          MySQL
	Redis          Redis
	Logger         Logger
}

// New создаёт Config, читая envPath если файл существует, иначе —
// из переменных окружения процесса. Возвращает ошибку при парсинге.
// Поддерживает файлы с расширением .env.example через godotenv.
func New(envPath string) (*Config, error) {
	var cfg Config
	if _, err := os.Stat(envPath); err == nil {
		// cleanenv выбирает парсер по расширению файла и не читает файлы
		// вида .env.example. Загружаем такой файл через godotenv и
		// повторяем чтение уже из переменных окружения.
		if err := godotenv.Load(envPath); err != nil {
			return nil, err
		}
		if err := cleanenv.ReadEnv(&cfg); err != nil {
			return nil, err
		}
	} else {
		if err := cleanenv.ReadEnv(&cfg); err != nil {
			return nil, err
		}
	}
	return &cfg, nil
}
