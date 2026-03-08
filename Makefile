.PHONY: build run test clean migrate-up migrate-down docker-up docker-down help

BINARY_NAME=codedb
VERSION?=$(shell git describe --tags --always --dirty)
LDFLAGS=-ldflags "-X main.Version=$(VERSION)"

help:
	@echo "CodeDB - Database-Native Collaborative Code Authoring"
	@echo ""
	@echo "Usage:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'

## build: Build the binary
build:
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/server

## run: Run the server locally
run:
	go run ./cmd/server

## test: Run all tests
test:
	go test -v -race -coverprofile=coverage.out ./...

## test-coverage: Run tests and show coverage
test-coverage: test
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated at coverage.html"

## lint: Run linter
lint:
	@which golangci-lint > /dev/null || go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	golangci-lint run ./...

## fmt: Format code
fmt:
	go fmt ./...

## vet: Run go vet
vet:
	go vet ./...

## clean: Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

## migrate-up: Run database migrations up
migrate-up:
	@which migrate > /dev/null || go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	migrate -path migrations -database "$(DATABASE_URL)" up

## migrate-down: Rollback database migrations
migrate-down:
	migrate -path migrations -database "$(DATABASE_URL)" down

## migrate-create: Create a new migration file (usage: make migrate-create name=migration_name)
migrate-create:
	migrate create -ext sql -dir migrations -seq $(name)

## docker-up: Start PostgreSQL with Docker
docker-up:
	docker run -d --name codedb-postgres \
		-e POSTGRES_USER=postgres \
		-e POSTGRES_PASSWORD=postgres \
		-e POSTGRES_DB=codedb \
		-p 5432:5432 \
		postgres:16-alpine

## docker-down: Stop PostgreSQL container
docker-down:
	docker stop codedb-postgres && docker rm codedb-postgres

## deps: Install dependencies
deps:
	go mod download
	go mod tidy

## proto: Generate protobuf code (if using gRPC)
proto:
	@which protoc > /dev/null || (echo "protoc not installed" && exit 1)
	protoc --go_out=. --go-grpc_out=. proto/*.proto

## all: Run fmt, vet, lint, test, and build
all: fmt vet lint test build
