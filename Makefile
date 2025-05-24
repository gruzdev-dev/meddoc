.PHONY: up test lint deps mocks test-coverage help

up:
	docker compose up -d --build

down:
	docker compose down -v --remove-orphans

test:
	go test -v ./...

lint:
	golangci-lint run

deps:
	go mod download
	go mod tidy
