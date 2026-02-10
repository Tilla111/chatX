include .env
MIGRATIONS_PATH = ./cmd/migrate/migrations

DB_URL = postgres://admin:admin@localhost:5432/chat_db?sslmode=disable


.PHONY: Build the app
build:
	go build -o bin/$(chatX) cmd/app

.PHONY: Run the app
run:
	go run cmd/app


.PHONY: migrate-create
migration:
	@migrate create -seq -ext sql -dir $(MIGRATIONS_PATH) $(filter-out $@,$(MAKECMDGOALS))


.PHONY: migrate-up
migrate-up:
	@migrate -path=$(MIGRATIONS_PATH) -database=$(DB_URL) up


.PHONY: migrate-down
migrate-down:
	@migrate -path=$(MIGRATIONS_PATH) -database=$(DB_URL) down $(filter-out $@,$(MAKECMDGOALS))


.PHONY: up build down logs clean
up: 
	@echo "Docker konteynerlarni ishga tushirish..."
	docker compose up -d

build:
	@echo "Docker konteynerlarni qayta qurish..."
	docker compose build


down:
	@echo "Docker konteynerlarni to'xtatish va o'chirish..."
	docker compose down

logs:
	@echo "Barcha xizmatlarning loglari..."
	docker compose logs 

clean: down
	@echo "Volume va tarmoqlarni tozalash..."
	docker volume prune 
	docker network prune 