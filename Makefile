.PHONY: run-local db-up db-down

include .env
export

run-local:
	@echo "Запуск приложения локально..."
	CONFIG_PATH=./config/config.local.yaml go run cmd/app/main.go


db-up:
	@echo "Запуск локального контейнера PostgreSQL..."
	docker run --rm --name local-postgres \
	  -e POSTGRES_USER=${POSTGRES_USER} \
	  -e POSTGRES_PASSWORD=${POSTGRES_PASSWORD} \
	  -e POSTGRES_DB=${POSTGRES_DB} \
	  -p ${POSTGRES_PORT}:5432 \
	  -d postgres:17

db-down:
	@echo "Остановка контейнера PostgreSQL..."
	docker stop local-postgres

compose-up: 
	docker compose -f docker-compose.yaml up -d --build