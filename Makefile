# NetTestLab Makefile

.PHONY: all build clean test proto install dev help

# Default target
all: build

# Build the project
build:
	@echo "🔨 Building NetTestLab..."
	@./scripts/build.sh

# Generate protobuf files
proto:
	@echo "📦 Generating protobuf files..."
	@buf generate

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	@rm -rf bin/
	@rm -rf api/
	@rm -rf clients/*/src/
	@rm -rf clients/*/lib/
	@rm -rf clients/*/target/
	@rm -rf docs/site/

# Run tests
test:
	@echo "🧪 Running tests..."
	@go test -v ./...

# Install dependencies
install:
	@echo "📦 Installing dependencies..."
	@go mod download
	@go mod tidy

# Development setup
dev: install proto
	@echo "🚀 Development environment ready!"

# Lint protobuf files
lint:
	@echo "🔍 Linting protobuf files..."
	@buf lint

# Format Go code
fmt:
	@echo "✨ Formatting Go code..."
	@go fmt ./...

# Run the server locally
run:
	@echo "🚀 Starting NetTestLab server..."
	@go run ./cmd/server

# Build documentation
docs:
	@echo "📚 Building documentation..."
	@mkdocs build

# Serve documentation locally
serve-docs:
	@echo "📖 Serving documentation..."
	@mkdocs serve

# Show help
help:
	@echo "NetTestLab Build Commands:"
	@echo "  build      - Build the entire project"
	@echo "  proto      - Generate protobuf files"
	@echo "  clean      - Clean build artifacts"
	@echo "  test       - Run tests"
	@echo "  install    - Install dependencies"
	@echo "  dev        - Setup development environment"
	@echo "  lint       - Lint protobuf files"
	@echo "  fmt        - Format Go code"
	@echo "  run        - Run server locally"
	@echo "  docs       - Build documentation"
	@echo "  serve-docs - Serve documentation locally"
	@echo "  help       - Show this help message"