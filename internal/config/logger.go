package config

type Logger struct {
	IsDev             bool   `env:"DEVELOPMENT" env-default:"true"` // Режим разработки (цветной вывод, читаемый формат)
	DisableCaller     bool   `env:"DISABLE_CALLER"`                 // Отключить вывод информации о вызывающем коде
	DisableStacktrace bool   `env:"DISABLE_STACKTRACE"`             // Отключить вывод stack trace для ошибок
	Encoding          string `env:"ENCODING" env-default:"console"` // Формат вывода: console или json
	Level             string `env:"LEVEL" env-default:"debug"`      // Уровень логирования: debug, info, warn, error и т.д.
}
