.PHONY: all build run test clean tidy fmt

APP_NAME=youtrack_backend
MAIN_PATH=./cmd/api

all: build

build:
	@echo "Building binary..."
	@mkdir -p bin
	@go build -o bin/$(APP_NAME) $(MAIN_PATH)

run:
	@go run $(MAIN_PATH)

test:
	@go test -v -race ./...

clean:
	@rm -rf bin/

tidy:
	@go mod tidy

fmt:
	@go fmt ./...
