// Package jwt предоставляет утилиты для генерации и валидации
// JWT-токенов (HS256) с claim'ом user_id.
package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Config хранит параметры генерации JWT: секрет и время жизни токена.
type Config struct {
	Secret []byte
	Expiry time.Duration
}

// Claims — пользовательские claim'ы JWT, содержащие ID пользователя.
type Claims struct {
	UserID int64 `json:"user_id"`
	jwt.RegisteredClaims
}

// Generate создаёт подписанный HS256-JWT с claim'ами user_id,
// issuedAt и expiresAt. Возвращает строковое представление токена.
func Generate(cfg Config, userID int64) (string, error) {
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(cfg.Expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(cfg.Secret)
}

// Validate парсит и валидирует JWT-строку. Возвращает user_id из claim'ов
// или ошибку, если токен невалиден, просрочён или использует неизвестный
// алгоритм подписи.
func Validate(tokenString string, secret []byte) (int64, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return secret, nil
	})
	if err != nil {
		return 0, err
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims.UserID, nil
	}
	return 0, errors.New("invalid token")
}
