.PHONY: all build test test-race lint clean env-example run run-dev run-memory kill db-reset db-migrate db-recreate db-verify db-psql

# Database connection variables with sensible defaults
DB_HOST ?= localhost
DB_PORT ?= 5432
DB_USER ?= train_user2
DB_NAME ?= youtrack2
PSQL    ?= psql
PSQLFLAGS = -v ON_ERROR_STOP=1 -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -d $(DB_NAME)

all: test build

build:
	mkdir -p bin
	go build -v -o bin/server.exe ./cmd/server

test:
	go test -v ./...

test-race:
	go test -v -race ./...

lint:
	go vet ./...

env-example:
	go test -v -run TestSchemaEngine_IntrospectionAndTemplate ./internal/platform/config/...

run:
	go run ./cmd/server

# Run with PostgreSQL for local development.
run-dev: run

# Run with in-memory storage; data is reset on restart.
run-memory:
	go run ./cmd/server --db-driver memory

# Reset database schema (development environment only)
db-reset:
	$(PSQL) $(PSQLFLAGS) -f scripts/reset_db.sql

# Apply all pending embedded database migrations.
db-migrate:
	go run ./cmd/migrate

# Destructively reset the schema, then recreate it using all migrations.
db-recreate: db-reset db-migrate

# Run database schema and seed integrity verification
db-verify:
	$(PSQL) $(PSQLFLAGS) -f scripts/verify_schema.sql

# Interactive psql session
db-psql:
	$(PSQL) $(PSQLFLAGS)

# POSIX-compliant termination (kill -TERM, never kill -SIGTERM in /bin/sh).
kill:
	pkill -TERM -f "bin/server" || true
	pkill -TERM -f "exe/server" || true

clean:
	go clean -cache -testcache
