.PHONY: help build test test-race lint vet fmt run migrate seed seed-admin docker-up docker-down docker-logs frontend-install frontend-dev frontend-build

help:
	@echo "Sahayak development targets:"
	@echo "  make run             - run the API server"
	@echo "  make migrate         - apply database migrations"
	@echo "  make seed            - seed reference data (states, documents, schemes)"
	@echo "  make seed-admin      - seed + create admin (PHONE=+91...)"
	@echo "  make test            - run Go tests"
	@echo "  make test-race       - run Go tests with the race detector"
	@echo "  make lint            - run golangci-lint (if installed)"
	@echo "  make docker-up       - start db + api via docker compose"
	@echo "  make docker-down     - stop docker compose stack"
	@echo "  make frontend-dev    - start the SvelteKit dev server"

build:
	go build ./...

vet:
	go vet ./...

fmt:
	gofmt -s -w .

test:
	go test ./...

test-race:
	go test -race -coverprofile=coverage.out ./...

lint:
	golangci-lint run

run:
	go run ./cmd/server

migrate:
	go run ./cmd/migrate up

migrate-down:
	go run ./cmd/migrate down

seed:
	go run ./cmd/seed

seed-admin:
	go run ./cmd/seed --admin-phone=$(PHONE)

docker-up:
	docker compose up -d --build

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f api

frontend-install:
	cd frontend && pnpm install

frontend-dev:
	cd frontend && pnpm dev

frontend-build:
	cd frontend && pnpm build
