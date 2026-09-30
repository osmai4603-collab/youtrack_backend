.PHONY: all build run db-up stop restart test clean tidy fmt

APP_NAME=youtrack_backend
MAIN_PATH=./cmd/server

all: build

build:
	@echo "Building binary..."
	@mkdir -p bin
	@go build -o bin/$(APP_NAME) $(MAIN_PATH)

db-up:
	@docker compose up -d --wait postgres

stop:
	@for port in 8099 8080; do \
		pids=$$( (lsof -t -iTCP:$$port -sTCP:LISTEN 2>/dev/null || ss -ltnp "sport = :$$port" | awk 'NR>1 {print $$NF}' | sed -n 's/.*pid=\([0-9]\+\).*/\1/p' || true) | grep -v "^$$$$" || true ); \
		if [ -n "$$pids" ]; then \
			kill -TERM $$pids 2>/dev/null || true; \
		fi; \
	done

restart: stop
	@$(MAKE) run

run: build db-up stop
	@./bin/$(APP_NAME)

test:
	@go test -v -race ./...

clean:
	@rm -rf bin/

tidy:
	@go mod tidy

fmt:
	@go fmt ./...
