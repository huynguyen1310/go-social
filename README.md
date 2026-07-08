# social

A social media API built with Go.

## Stack

- **Router:** [chi](https://github.com/go-chi/chi) v5
- **Database:** PostgreSQL 16
- **Migrations:** [golang-migrate](https://github.com/golang-migrate/migrate)

## Setup

### Prerequisites

- Go 1.26+
- PostgreSQL (via Docker or local)
- [golang-migrate](https://github.com/golang-migrate/migrate) CLI

### Quick start

```sh
# Start PostgreSQL
docker compose up -d

# Copy and configure env
cp .env.example .env

# Run migrations
make migrate-up

# Start server
go run ./cmd/api
```

## Makefile commands

```sh
make migrate-up            # Apply pending migrations
make migrate-down          # Rollback 1 step
make migrate-down N        # Rollback N steps
make migrate-create        # Create new migration (prompts for name)
```

## Project structure

```
cmd/
└── api/              # HTTP server entrypoint
internal/
├── env/              # Env var helpers
└── store/            # Data access layer
migrations/           # SQL migration files
```
