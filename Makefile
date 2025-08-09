MIGRATIONS_DIR := migrations
DB_URL := "postgres://postgres:postgres@localhost:5432/orderdb?sslmode=disable"
.PHONY: migrate-create
migrate-create:
	@read -p "Enter migration name: " name; \
    goose -dir $(MIGRATIONS_DIR) create $${name} sql


.PHONY: migrate-up
migrate-up:
	@echo "==> Running migrations up..."
	@goose -dir $(MIGRATIONS_DIR) postgres "$(DB_URL)" up

.PHONY: migrate-down
migrate-down:
	@echo "==> Running migrations down..."
	@goose -dir $(MIGRATIONS_DIR) down
	
.PHONY: go-deps
go-deps:
	@echo "==> Installing Go dependencies..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

.PHONY: go-lint
go-lint:
	@echo "==> Running linter..."
	@golangci-lint run ./... -v

