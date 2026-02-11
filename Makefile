include .env

MIGRATIONS_PATH := ./cmd/migrate/migrations
DB_URL := postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable

.PHONY: build-api run-api migration migrate-up migrate-down docker-up docker-build docker-down docker-logs clean gen-docs

build-api:
	go build -o bin/chatx-api ./cmd/api

run-api:
	go run ./cmd/api

migration:
	@migrate create -seq -ext sql -dir $(MIGRATIONS_PATH) $(filter-out $@,$(MAKECMDGOALS))

migrate-up:
	@migrate -path=$(MIGRATIONS_PATH) -database=$(DB_URL) up

migrate-down:
	@migrate -path=$(MIGRATIONS_PATH) -database=$(DB_URL) down $(filter-out $@,$(MAKECMDGOALS))

docker-up:
	docker compose up -d

docker-build:
	docker compose build

docker-down:
	docker compose down

docker-logs:
	docker compose logs

clean: docker-down
	docker volume prune
	docker network prune

gen-docs:
	@swag init -g main.go -d cmd/api,internal -o docs && swag fmt
