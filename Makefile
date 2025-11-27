.PHONY: help build test clean fmt lint

# Default target
help:
	@echo "confluent-go - Kubernetes-friendly Go package for Confluent"
	@echo ""
	@echo "Available targets:"
	@echo "  build      - Build the example binary"
	@echo "  test       - Run tests"
	@echo "  test-cover - Run tests with coverage"
	@echo "  fmt        - Format code"
	@echo "  lint       - Run linter"
	@echo "  clean      - Clean build artifacts"
	@echo "  help       - Show this help message"

build:
	@echo "Building example..."
	go build -o dist/example ./cmd/examples

test:
	@echo "Running tests..."
	go test -v ./...

test-cover:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

fmt:
	@echo "Formatting code..."
	gofmt -w -s ./pkg ./cmd

lint:
	@echo "Running linter..."
	go vet ./...

clean:
	@echo "Cleaning up..."
	rm -rf dist/
	rm -f coverage.out coverage.html
