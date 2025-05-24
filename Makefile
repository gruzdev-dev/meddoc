.PHONY: up down lint deps test test-unit test-integration test-up test-down 

up:
	docker compose up -d --build

down:
	docker compose down -v --remove-orphans

lint:
	go fmt ./...
	golangci-lint run

deps:
	go mod download
	go mod tidy

test: test-unit test-integration

test-unit:
	go test -v ./...

test-up:
	docker compose up -d mongodb

test-down:
	docker compose down mongodb

test-integration: test-up
	go test -v ./tests/... -tags=integration
	$(MAKE) test-down
