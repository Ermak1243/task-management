// Package app реализует запуск приложения: загрузку конфигурации,
// связывание зависимостей, поднятие HTTP-сервера и корректное завершение.
package app

import (
	"context"
	"log"
	"os"
	"os/signal"
	"time"

	"task-management/api/server/route"
	"task-management/internal/config"
	"task-management/internal/database/mysql"
	"task-management/internal/database/redis"
	"task-management/internal/repository"
	"task-management/internal/service"
	"task-management/pkg/logger"

	"github.com/gofiber/fiber/v2"
)

// Run — точка входа приложения. Загружает конфигурацию, инициализирует
// инфраструктуру (MySQL, Redis), создаёт репозитории, сервисы и маршруты,
// запускает HTTP-сервер и блокируется до получения сигнала OS (Interrupt),
// после чего выполняет graceful shutdown с таймаутом 10 секунд.
func Run() {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "./configs/.env"
	}

	cfg, err := config.New(configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	appLogger := logger.New(logger.Config{
		Level:    cfg.Logger.Level,
		Encoding: cfg.Logger.Encoding,
		IsDev:    cfg.Logger.IsDev,
	})

	dbConn := mysql.New(cfg.MySQL)
	defer dbConn.Close()
	db := dbConn.SQLX()

	client := redis.New(cfg.Redis)
	defer client.Close()

	txManager := mysql.NewTxManager(db)
	taskCache := redis.NewCache(client.Client)

	userRepo := repository.NewUser(db)
	teamRepo := repository.NewTeam(db)
	memberRepo := repository.NewTeamMember(db)
	taskRepo := repository.NewTask(db)
	historyRepo := repository.NewTaskHistory(db)
	commentRepo := repository.NewTaskComment(db)

	auth := service.NewAuth(userRepo, cfg.JWTSecret, cfg.JWTExpiry, cfg.ContextTimeout)
	team := service.NewTeam(teamRepo, memberRepo, userRepo, txManager, cfg.ContextTimeout)
	task := service.NewTask(taskRepo, historyRepo, commentRepo, memberRepo, taskCache, txManager, cfg.ContextTimeout)
	statsRepo := repository.NewStats(db)
	stats := service.NewStats(statsRepo, memberRepo, cfg.ContextTimeout)

	app := fiber.New()

	route.Setup(app, cfg, auth, team, task, stats, client)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		appLogger.Info("Shutting down...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := app.ShutdownWithContext(ctx); err != nil {
			appLogger.Errorf("server shutdown error: %v", err)
		}
	}()

	appLogger.Infof("Server starting on port %s", cfg.ServerPort)
	if err := app.Listen(":" + cfg.ServerPort); err != nil {
		log.Panic(err)
	}
}
