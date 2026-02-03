.PHONY: build test test-update lint fmt clean install

# Build the CLI binary
build:
	go build -o bin/jac cmd/jac/main.go

# Run all tests
test:
	go test ./...

# Update golden files
test-update:
	go test ./... -update

# Run linter (go vet)
lint:
	go vet ./...

# Format all Go code
fmt:
	go fmt ./...

# Clean build artifacts
clean:
	rm -rf bin/

# Install dependencies
install:
	go mod download

# Run all checks (fmt, lint, test)
check: fmt lint test

# Default target
all: build
