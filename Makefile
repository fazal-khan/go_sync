# Variables
BINARY_NAME=gosync
MAIN_PACKAGE=./cmd

.PHONY: all build run test coverage clean tool help

# Default target runs when you just type 'make'
all: test build

## build: Build the Go binary
build:
	@echo "Building binary..."
	go build -o bin/$(BINARY_NAME) $(MAIN_PACKAGE)

## run: Build and execute the binary
run: build
	@echo "Running application..."
	./bin/$(BINARY_NAME)

## test: Run all project tests
test:
	@echo "Running tests..."
	go test -v ./...

## coverage: Run tests with coverage report
coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	@echo "\n--- Coverage Summary ---"
	@go tool cover -func=coverage.out
	@echo "\nHTML report: go tool cover -html=coverage.out"

## clean: Remove build artifacts
clean:
	@echo "Cleaning up..."
	go clean
	rm -rf bin/

## tool: Run go vet and fmt
tool:
	@echo "Formatting and vetting code..."
	go fmt ./...
	go vet ./...

## help: Display this help message
help:
	@echo "Usage:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'
