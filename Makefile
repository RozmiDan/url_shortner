.PHONY: run-local 

# Запуск приложения локально с переменной CONFIG_PATH
run-local:
	@echo "Запуск приложения локально..."
	CONFIG_PATH=./config/config.yaml go run cmd/app/main.go
