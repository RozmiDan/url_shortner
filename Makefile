.PHONY: run-local db-up db-down

# Загружаем переменные окружения из .env файла
include .env
export

# Запуск приложения локально с переменной CONFIG_PATH
run-local:
	@echo "Запуск приложения локально..."
	CONFIG_PATH=./config/config.local.yaml go run cmd/app/main.go

# Запуск PostgreSQL в Docker с параметрами из .env
db-up:
	@echo "Запуск локального контейнера PostgreSQL..."
	docker run --rm --name local-postgres \
	  -e POSTGRES_USER=${POSTGRES_USER} \
	  -e POSTGRES_PASSWORD=${POSTGRES_PASSWORD} \
	  -e POSTGRES_DB=${POSTGRES_DB} \
	  -p ${POSTGRES_PORT}:5432 \
	  -d postgres:17

# Остановка локального контейнера PostgreSQL
db-down:
	@echo "Остановка контейнера PostgreSQL..."
	docker stop local-postgres

# Run docker compose (with backend and db)
compose-up: 
	docker compose -f docker-compose.yaml up -d --build