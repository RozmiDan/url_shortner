.PHONY: run-app stop-app db-up db-down prom-up prom-down run-local stop-local run-jmeter

include .env
export

# Запуск всего приложения
run-local: db-up prom-up graf-up
	@echo "Ожидание запуска PostgreSQL"
	sleep 1
	$(MAKE) run-app

# Остановка всего приложения
stop-local: db-down prom-down graf-down stop-app

# Запуск приложения локально с переменной CONFIG_PATH
run-app:
	@echo "Запуск приложения локально"
	go build -o bin/url_shortener ./cmd/app/main.go
	CONFIG_PATH=./config/config.local.yaml ./bin/url_shortener

stop-app:
	@echo "Остановка приложения"
	-pkill -SIGTERM -f './bin/url_shortener' || true

# Запуск PostgreSQL в Docker с параметрами из .env
db-up:
	@echo "Запуск контейнера PostgreSQL..."
	docker run --rm --name local-postgres \
	  -e POSTGRES_USER=${POSTGRES_USER} \
	  -e POSTGRES_PASSWORD=${POSTGRES_PASSWORD} \
	  -e POSTGRES_DB=${POSTGRES_DB} \
	  -p ${POSTGRES_PORT}:5432 \
	  -d postgres:17

# Остановка контейнера PostgreSQL
db-down:
	@echo "Остановка контейнера PostgreSQL..."
	docker stop local-postgres

graf-up:
	@echo "Запуск контейнера Grafana"
	docker run \
	--name grafana \
	--rm \
	-p 3000:3000 \
	-d grafana/grafana

graf-down:
	@echo "Остановка контейнера Grafana"
	docker stop grafana

# Запуск Prometheus
prom-up:
	@echo "Запуск контейнера Prometheus"
	docker run \
	--name prometheus \
	--rm \
	-d \
	-p 9090:9090 \
	-v ./prometheus.yml:/etc/prometheus/prometheus.yml \
	prom/prometheus

prom-down:
	@echo "Остановка контейнера Prometheus"
	docker stop prometheus

# Run docker compose
compose-up: 
	docker compose -f docker-compose.yaml up -d --build

compose-down: 
	docker compose down

run-jmeter:
	@echo "Запуск JMeter load-тестов..."
	jmeter -n \
	  -t jmeter_files/Test_500users.jmx \
	  -l jmeter_files/log1.jtl \
	  -j jmeter_files/jmeter.log \
	  -e \
	  -o jmeter_files/report
	@echo "Отчёт сгенерирован в jmeter_files/report"