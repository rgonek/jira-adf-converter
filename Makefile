.PHONY: build test test-race test-update lint fmt fmt-check staticcheck vuln-check tidy-check clean install check

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

# Verify gofmt has no pending changes
fmt-check:
	@test -z "$$(gofmt -l .)" || (gofmt -l . && exit 1)

# Run static analysis with staticcheck (requires staticcheck installed)
staticcheck:
	staticcheck ./...

# Run vulnerability scan (requires govulncheck installed)
vuln-check:
	govulncheck ./...

# Ensure go.mod/go.sum stay tidy
tidy-check:
	go mod tidy
	git diff --exit-code -- go.mod go.sum

# Clean build artifacts
clean:
	rm -rf bin/

# Install dependencies
install:
	go mod download

# Run all checks (fmt-check, lint, test, tidy-check)
check: fmt-check lint test tidy-check

# Default target
all: build
