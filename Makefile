include .env
export

.PHONY: up down migrate-up migrate-down build run test seed ingest-ted ingest-datagovcy distribute

up:
	docker compose up -d
	@echo "waiting for postgres..."
	@until docker compose exec -T postgres pg_isready -U promitheies >/dev/null 2>&1; do sleep 1; done

down:
	docker compose down

migrate-up:
	go run ./cmd/migrate -direction up

migrate-down:
	go run ./cmd/migrate -direction down -steps 1

build:
	go build -o bin/server ./cmd/server
	go build -o bin/ingest-ted ./cmd/ingest-ted
	go build -o bin/ingest-datagovcy ./cmd/ingest-datagovcy
	go build -o bin/ingest-opentender ./cmd/ingest-opentender
	go build -o bin/distribute ./cmd/distribute
	go build -o bin/jobs ./cmd/jobs

run:
	go run ./cmd/server

test:
	go test ./...

seed:
	go run ./cmd/jobs seed-cpv

ingest-ted:
	go run ./cmd/ingest-ted

ingest-datagovcy:
	go run ./cmd/ingest-datagovcy

distribute:
	go run ./cmd/distribute
