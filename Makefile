.ONESHELL:
TAB=echo "\t"
CURRENT_DIR = $(shell pwd)

# Переменные для подключения к БД
DB_DSN=postgresql://developer:developer@localhost:5432/gobase?sslmode=disable

help:
	@echo "Доступные команды:"
	@$(TAB) make up-server      - запустить сервер
	@$(TAB) make up-docker      - запуск PostgreSQL в Docker
	@$(TAB) make down-docker    - остановка Docker контейнеров
	@$(TAB) make migrate-up     - применить все миграции
	@$(TAB) make migrate-down   - откатить последнюю миграцию
	@$(TAB) make migrate-create name=<имя> - создать новую миграцию
	@$(TAB) make lint           - запустить статический анализ кода
	@$(TAB) make test           - запустить тесты
	@$(TAB) make deps           - установить зависимости
	@$(TAB) make doc            - сгенерировать Swagger документацию
	@$(TAB) make clean          - очистить Docker volumes
	@$(TAB) make help           - показать эту справку

# Запуск сервера с конфигом
up-server:
	go run ./cmd/server/main.go config.yaml

# Запуск PostgreSQL в Docker
up-docker:
	docker-compose up -d

# Остановка Docker контейнеров
down-docker:
	docker-compose down

# Полная очистка (включая volumes)
clean:
	docker-compose down -v
	rm -rf ./docker/postgres/data

# Применение миграций
migrate-up:
	migrate -path migrations -database "$(DB_DSN)" up

# Откат последней миграции
migrate-down:
	migrate -path migrations -database "$(DB_DSN)" down 1

# Создание новой миграции
migrate-create:
	@if [ -z "$(name)" ]; then \
		echo "Ошибка: укажите имя миграции через name=<имя>"; \
		exit 1; \
	fi
	migrate create -ext sql -dir migrations -seq $(name)

# Установка зависимостей
deps:
	go mod download
	go mod tidy

# Запуск тестов
test:
	go test -v ./...

# Статический анализ кода
lint:
	@echo "Запуск статического анализа..."
	go run ./cmd/staticlint/main.go ./...

# Генерация Swagger документации
doc:
	@echo "Генерация Swagger документации..."
	~/go/bin/swag init -g ./cmd/server/main.go -o ./docs
	@echo "Документация сгенерирована в ./docs"
	@echo "После запуска сервера доступна по адресу: http://localhost:8080/swagger/index.html"

.PHONY: help up-server up-docker down-docker clean migrate-up migrate-down migrate-create deps test lint doc