include .env
export

MIGRATIONS_PATH=cmd/migrate/migrations

.PHONY: migrate-create migrate-up migrate-down migrate-force migrate-version \
        migrate-drop migrate-fresh migrate-refresh migrate-status seed test

## Like `artisan make:migration`
migrate-create:
	@read -p "Enter migration name: " name; \
	migrate create -seq -ext sql -dir $(MIGRATIONS_PATH) $$name

## Like `artisan migrate`
migrate-up:
	@migrate -path=$(MIGRATIONS_PATH) -database=$(DB_ADDR) up

## Like `artisan migrate:rollback` (default 1 step, or pass N)
migrate-down:
	@migrate -path=$(MIGRATIONS_PATH) -database=$(DB_ADDR) down $(filter-out $@,$(MAKECMDGOALS))

## Like `artisan migrate:reset` (rollback everything)
migrate-drop:
	@migrate -path=$(MIGRATIONS_PATH) -database=$(DB_ADDR) down -all

## Like `artisan migrate:fresh` (drop everything, re-migrate from scratch)
## Usage: make migrate-fresh [SEED=true]
migrate-fresh:
	@migrate -path=$(MIGRATIONS_PATH) -database=$(DB_ADDR) drop -f
	@migrate -path=$(MIGRATIONS_PATH) -database=$(DB_ADDR) up
	@if [ "$(SEED)" = "true" ]; then $(MAKE) seed; fi

## Like `artisan migrate:refresh` (rollback all, then re-run)
migrate-refresh: migrate-drop migrate-up

## Fix a "dirty" migration state (rare in Laravel, common gotcha in golang-migrate)
migrate-force:
	@migrate -path=$(MIGRATIONS_PATH) -database=$(DB_ADDR) force $(filter-out $@,$(MAKECMDGOALS))

## Like `artisan migrate:status`
migrate-version:
	@migrate -path=$(MIGRATIONS_PATH) -database=$(DB_ADDR) version

## Like `artisan db:seed`
seed:
	@go run ./cmd/seed

# allows passing extra args like: make migrate-down 2
%:
	@:

## Run tests
## Usage: make test [ARGS="-v -run TestName"]
test:
	@go test $(ARGS) ./...

gen-docs:
	@swag init -g ./api/main.go -d cmd,internal && swag fmt
