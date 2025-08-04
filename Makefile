MIGRATIONS_DIR := migrations

.PHONY: migrate-create
migrate-create:
	@read -p "Enter migration name: " name; \
    goose -dir $(MIGRATIONS_DIR) create $${name} sql

.PHONY: go-deps
go-deps:
	@echo "==> Installing Go dependencies..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

.PHONY: go-lint
go-lint:
	@echo "==> Running linter..."
	@golangci-lint run ./... -v