.phony: build build-all build-linux build-mac build-windows install run clean test test_coverage deps lint fmt dev coverage-dir

APP_NAME=haki

build:
	@echo "Building..."
	@go build -o bin/$(APP_NAME) .

build-all: build-linux build-mac build-windows

build-linux:
	@echo "Building for Linux..."
	@GOOS=linux GOARCH=amd64 go build -o bin/$(APP_NAME)-linux-amd64 .

build-mac:
	@echo "Building for macOS..."
	@GOOS=darwin GOARCH=amd64 go build -o bin/$(APP_NAME)-darwin-amd64 .

build-windows:
	@echo "Building for Windows..."
	@GOOS=windows GOARCH=amd64 go build -o bin/$(APP_NAME)-windows-amd64.exe .

install:
	@echo "Installing..."
	@go install .

run: build
	@echo "Running..."
	@./bin/$(APP_NAME)

clean:
	@echo "Cleaning..."
	@rm -rf bin

test:
	@echo "Testing..."
	@go test -v ./... 

test_coverage: coverage-dir
	@echo "Testing..."
	@go test -v -race -coverprofile=coverage/coverage.txt -covermode=atomic ./... 

deps:
	@echo "Installing dependencies..."
	@go mod tidy

lint:
	@echo "Linting..."
	@golangci-lint run

fmt:
	@echo "Formatting..."
	@go fmt ./...

dev: fmt lint test build

coverage-dir:
	@mkdir -p coverage