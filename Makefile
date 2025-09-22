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

# Docker-based testing
test-docker: build-docker-images
	@echo "🐳 Building NetTestLab for Docker testing..."
	@GOOS=linux GOARCH=amd64 go build -o bin/nettestlab ./cmd/server
	@echo "🚀 Starting Docker integration tests..."
	@docker-compose -f docker-compose.test.yml up --abort-on-container-exit
	@echo "📊 Collecting test results..."
	@docker-compose -f docker-compose.test.yml down

# Build Docker images with caching
build-docker-images:
	@echo "🏗️  Building Docker images (with cache)..."
	@docker-compose -f docker-compose.test.yml build

# Force rebuild Docker images without cache
rebuild-docker-images:
	@echo "🔄 Rebuilding Docker images (no cache)..."
	@docker-compose -f docker-compose.test.yml build --no-cache

# Pre-build and cache Docker images
cache-docker-images: rebuild-docker-images
	@echo "💾 Docker images cached successfully!"

# Clean up Docker test environment
test-docker-clean:
	@echo "🧹 Cleaning up Docker test environment..."
	@docker-compose -f docker-compose.test.yml down -v --remove-orphans
	@docker system prune -f

# Run integration tests in existing Docker environment
test-integration:
	@echo "🔧 Running integration tests..."
	@docker-compose -f docker-compose.test.yml exec test-client /tests/run-integration-tests.sh

# View test results
test-results:
	@echo "📋 Displaying test results..."
	@if [ -f ./tests/integration/results/test_report.json ]; then \
		cat ./tests/integration/results/test_report.json | jq .; \
	else \
		echo "No test results found. Run 'make test-docker' first."; \
	fi

# Monitor Docker containers during testing
test-logs:
	@echo "📝 Showing Docker container logs..."
	@docker-compose -f docker-compose.test.yml logs -f

# Show help
help:
	@echo "NetTestLab Build Commands:"
	@echo "  build               - Build the entire project"
	@echo "  proto               - Generate protobuf files"
	@echo "  clean               - Clean build artifacts"
	@echo "  test                - Run unit tests"
	@echo "  test-docker         - Run full Docker integration tests"
	@echo "  build-docker-images - Build Docker images (with cache)"
	@echo "  rebuild-docker-images- Rebuild Docker images (no cache)"
	@echo "  cache-docker-images - Pre-build and cache all Docker images"
	@echo "  test-integration    - Run integration tests in existing Docker env"
	@echo "  test-docker-clean   - Clean up Docker test environment"
	@echo "  test-results        - View test results"
	@echo "  test-logs           - Monitor Docker container logs"
	@echo "  install             - Install dependencies"
	@echo "  dev                 - Setup development environment"
	@echo "  lint                - Lint protobuf files"
	@echo "  fmt                 - Format Go code"
	@echo "  run                 - Run server locally"
	@echo "  docs                - Build documentation"
	@echo "  serve-docs          - Serve documentation locally"
	@echo "  help                - Show this help message"