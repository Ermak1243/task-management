ENV_FILE := configs/.env
COMPOSE := docker compose --env-file $(ENV_FILE)

.PHONY: help up down migrate test test-unit test-integration build swagger clean

help:
	@echo "Доступные команды:"
	@echo "  make up              - Собрать и запустить все сервисы в фоне"
	@echo "  make down            - Остановить все контейнеры"
	@echo "  make migrate         - Применить миграции к БД"
	@echo "  make test            - Запустить все тесты"
	@echo "  make test-unit       - Запустить только unit-тесты"
	@echo "  make test-integration- Запустить интеграционные тесты"
	@echo "  make build           - Собрать бинарник приложения (локально)"
	@echo "  make swagger         - Перегенерировать Swagger-документацию"
	@echo "  make clean           - Остановить контейнеры и удалить volumes"

up: down
	$(COMPOSE) up -d --build mysql redis
	@timeout /t 10 /nobreak > nul
	$(COMPOSE) --profile tools run --rm migrations
	$(COMPOSE) up -d --build app

down:
	$(COMPOSE) down

migrate:
	$(COMPOSE) --profile tools run --rm migrations

test: test-unit test-integration

test-unit:
	go test ./internal/service/... -cover

test-integration:
	go test ./test/integration/... -cover

build:
	go build -o bin/app ./cmd/app
	
swagger:
	swag init -g cmd/app/main.go -o docs --parseDependency --parseInternal

clean:
	$(COMPOSE) down -v
