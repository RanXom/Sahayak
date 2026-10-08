.PHONY: help build test test-race lint vet fmt run migrate migrate-down seed

help:
	@echo "Sahayak development targets:"
	@echo "  make run             - run the API server"
	@echo "  make migrate         - apply database migrations"
	@echo "  make seed            - seed reference data (states, documents, schemes)"
	@echo "  make test            - run Go tests"
	@echo "  make test-race       - run Go tests with the race detector"
	@echo "  make lint            - run golangci-lint (if installed)"

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
