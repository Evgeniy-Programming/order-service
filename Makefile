SERVICE_NAME=order-service
GO_FILES=$(shell find . -name '*.go')

# --- Docker ---

#Запуск всех контейнеров в фоновом режиме
up:
	@echo "Starting Docker containers..."
	@docker-compose up -d

#Остановка и удаление контейнеров
down:
	@echo "Stopping Docker containers..."
	@docker-compose down

#Логирование докер-контейнеров
logs:
	@docker-compose logs -f

# --- Go Application ---

#Основной код
run:
	@echo "Running the main service..."
	@go run ./cmd/service/main.go

#Установка зависимостей
tidy:
	@echo "Tidying Go modules..."
	@go mod tidy

#Установка библиотек(валидатор etc.)
install-deps: tidy
	@echo "Installing dependencies..."
	@go get github.com/go-playground/validator/v10
	@go get github.com/go-chi/chi/v5
	@go get github.com/brianvoe/gofakeit/v6
	@go get go.uber.org/mock/mockgen@latest

# --- Database ---

#SQL-миграция
migrate-up:
	@echo "Applying database migrations..."
	@cat schema.sql | docker exec -i order_db psql -U order_user -d orders_db

# --- Publisher ---

#Публикатор случайных заказов
#Пример: make publish count=50
publish:
	@echo "Publishing messages..."
	@go run ./tools/publisher/main.go --count=$(count)

test:
	@go test ./...
#Для полного нулевого запуска
start: down up migrate-up run

.PHONY: up down logs run tidy install-deps migrate-up publish start