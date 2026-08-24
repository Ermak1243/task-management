# Task Management Service

REST API для управления задачами внутри команд с ролевой моделью, историей изменений, кешированием и SQL-отчётами.

---

## 📦 Стек

- **Go** 1.24
- **MySQL** 8.0
- **Redis** 7
- **Docker** & **Docker Compose**
- **JWT** (аутентификация)
- **Swagger** (документация)
- **Rate Limiting**, **Validator**
- **SQLx**, **go-redis**, **zap** (логгер)

---

## 🚀 Быстрый старт

### 1. Клонируйте репозиторий
```bash
git clone https://github.com/your-username/task-management.git
cd task-management
```

### 2. Настройте переменные окружения
Скопируйте пример конфигурации и отредактируйте при необходимости:
```bash
cp configs/.env.example configs/.env
```

### 3. Запустите приложение через Makefile
```bash
make up
```
Эта команда выполнит:
- остановку старых контейнеров;
- запуск MySQL и Redis;
- применение миграций к БД;
- сборку и запуск приложения.

После запуска API будет доступно по адресу: `http://localhost:8080` (порт можно изменить в `.env`).

---

## 📚 Документация API

Swagger UI доступен после генерации документации:

1. Установите `swag` (если ещё нет):
   ```bash
   go install github.com/swaggo/swag/cmd/swag@latest
   ```
2. Сгенерируйте файлы:
   ```bash
   swag init -g cmd/app/main.go -o docs --parseDependency --parseInternal
   ```
3. Откройте в браузере: `http://localhost:8080/swagger/index.html`
   
   Важно!!! Для авторизации в Swagger UI вставьте полученный при логине токен в поле Authorize с префиксом Bearer (например, Bearer eyJhbGciOi...). Middleware ожидает заголовок Authorization: Bearer <token>, и Swagger не добавляет префикс автоматически.

---

## 🧪 Тестирование

Все тесты:
```bash
make test
```

Только unit-тесты (без инфраструктуры):
```bash
make test-unit
```

Только интеграционные тесты (требуют работающих MySQL и Redis):
```bash
make test-integration
```

---

## 🗄️ Управление миграциями

Миграции применяются автоматически при `make up`, но вы также можете выполнить их вручную:
```bash
make migrate
```

Все SQL-файлы лежат в каталоге `migrations/`.  
При добавлении новых миграций просто положите их туда и повторно запустите `make migrate`.

---

## 🛠️ Доступные команды Makefile

| Команда                | Описание |
|------------------------|----------|
| `make up`              | Остановить контейнеры → запустить MySQL/Redis → применить миграции → запустить приложение |
| `make down`            | Остановить контейнеры |
| `make migrate`         | Применить миграции к БД |
| `make test`            | Запустить все тесты |
| `make test-unit`       | Запустить только unit-тесты |
| `make test-integration`| Запустить интеграционные тесты |
| `make build`           | Собрать бинарник приложения (локально) |
| `make clean`           | Остановить контейнеры и удалить тома с данными |

---

## 🔧 Конфигурация

Все настройки задаются через файл `configs/.env`. Основные параметры:

| Переменная               | Описание |
|--------------------------|----------|
| `SERVER_PORT`            | Порт, на котором слушает приложение |
| `JWT_SECRET`             | Секретный ключ для подписи JWT (обязательно измените в продакшене!) |
| `JWT_EXPIRY`             | Время жизни токена |
| `MYSQL_HOST`, `MYSQL_PORT`, `MYSQL_USER`, ... | Подключение к MySQL |
| `REDIS_HOST`, `REDIS_PORT`, ... | Подключение к Redis |
| `LOG_LEVEL`              | Уровень логирования (debug, info, warn, error) |
| `LOG_ENCODING`           | Формат логов (console или json) |

---

## 🌐 Примеры запросов

### Регистрация
```bash
curl -X POST http://localhost:8080/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"123456","name":"John"}'
```

### Логин
```bash
curl -X POST http://localhost:8080/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"123456"}'
```
Ответ содержит JWT-токен, который нужно передавать в заголовке `Authorization: Bearer <token>`.

### Создание команды (только авторизованные)
```bash
curl -X POST http://localhost:8080/api/v1/teams \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"name":"Моя команда"}'
```

### Получение статистики команды (только admin/owner)
```bash
curl http://localhost:8080/api/v1/teams/1/stats \
  -H "Authorization: Bearer <token>"
```

Полный перечень эндпоинтов доступен в Swagger.

---

## 🏗️ Архитектура

Проект построен по принципам **чистой архитектуры**:

- **Слой репозиториев** – работа с БД (MySQL) через `sqlx`.
- **Слой сервисов** – бизнес-логика, транзакции, кеширование.
- **Слой контроллеров** – HTTP-обработчики (Fiber).
- **Слой маршрутов** – регистрация эндпоинтов с middleware (JWT, rate limit, CORS).

### Ключевые особенности
- **Оптимистическая блокировка** (поле `version`) для защиты от конкурентных обновлений.
- **Транзакционность** при создании/обновлении задач и записи истории.
- **Кеширование** списка задач в Redis с инвалидацией через SCAN (безопасно для продакшена).
- **Разделённый rate limit** для публичных и защищённых маршрутов.
- **Курсорная + offset-пагинация** (параметры `?limit=20&offset=0` или `?cursor=100&limit=20`).
- **Единый SQL-отчёт** для статистики по команде (только для администраторов).

---

## 📦 Запуск без Docker (локальная разработка)

Убедитесь, что у вас установлены Go, MySQL и Redis. Затем:

```bash
# Скопировать .env и настроить подключение к локальным БД
cp configs/.env.example configs/.env

# Применить миграции (через утилиту migrate или вручную)
# Собрать и запустить
make build
./bin/app
```

---