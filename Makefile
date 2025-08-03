MIGRATIONS_DIR := migrations

.PHONY: migrate-create
migrate-create:
	@read -p "Enter migration name: " name; \
    goose -dir $(MIGRATIONS_DIR) create $${name} sql