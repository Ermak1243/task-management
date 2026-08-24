package main

import (
	"task-management/internal/app"
)

// @title Task Management API
// @version 1.0
// @description API для управления задачами, командами и статистикой.
// @host localhost:8080
// @BasePath /api/v1
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main() {
	app.Run()
}
