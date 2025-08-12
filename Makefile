MIGRATIONS_DIR := migrations
DB_URL := "postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=${POSTGRES_SSLMODE}"

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
	@goose -dir $(MIGRATIONS_DIR) postgres "$(DB_URL)" down
	
.PHONY: go-deps
go-deps:
	@echo "==> Installing Go dependencies..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

.PHONY: go-lint
go-lint:
	@echo "==> Running linter..."
	@golangci-lint run ./... -v


.PHONY: go-test
go-test:
	@echo "==> Running tests..."
	@go test -v ./... 