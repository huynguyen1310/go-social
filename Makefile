include .env

MIGRATIONS_PATH=cmd/migrate/migrations

.PHONY: migrate-create migrate-up migrate-down

migrate-create:
	@read -p "Enter migration name: " name; \
	migrate create -seq -ext sql -dir $(MIGRATIONS_PATH) $$name

migrate-up:
	migrate -path=$(MIGRATIONS_PATH) -database=$(DB_ADDR) up

migrate-down:
	migrate -path=$(MIGRATIONS_PATH) -database=$(DB_ADDR) down $(filter-out $@,$(MAKECMDGOALS))
