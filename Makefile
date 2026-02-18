.PHONY: build test test-race test-update lint fmt clean install

# Use a repo-local Go build cache to avoid permission issues.
GOCACHE ?= $(CURDIR)/.gocache
export GOCACHE

# Build the CLI binary
build:
	go build -o bin/jac cmd/jac/main.go

# Run all tests
test:
	go test ./...

# Run tests with race detector (requires CGO and a C compiler)
test-race:
	go test -race ./...

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
