.PHONY: help build test vet fmt run

help:
	@echo "Sahayak development targets:"
	@echo "  make run    - run the API server"
	@echo "  make test   - run Go tests"
	@echo "  make build  - build all Go packages"

build:
	go build ./...

vet:
	go vet ./...

fmt:
	gofmt -s -w .

test:
	go test ./...
