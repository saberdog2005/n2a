# Makefile for Note2Anki

BINARY_NAME=note2anki
GO=go
GOFLAGS=-v
LDFLAGS=-s -w

# Build directories
BUILD_DIR=build
DIST_DIR=dist

# Platforms
PLATFORMS=darwin linux windows
ARCHITECTURES=amd64 arm64

.PHONY: all build clean test install uninstall run deps format lint help

# Default target
all: clean deps build

# Help target
help:
	@echo "Note2Anki Build System"
	@echo ""
	@echo "Available targets:"
	@echo "  make build       - Build the binary for current platform"
	@echo "  make install     - Build and install binary to GOPATH/bin"
	@echo "  make clean       - Remove build artifacts"
	@echo "  make deps        - Download dependencies"
	@echo "  make test        - Run tests"
	@echo "  make format      - Format code with gofmt"
	@echo "  make lint        - Run golint"
	@echo "  make cross       - Build for all platforms"
	@echo "  make run         - Run with example files"

# Build for current platform
build: deps
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) main.go
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Install binary
install: build
	@echo "Installing $(BINARY_NAME)..."
	@$(GO) install $(GOFLAGS)
	@echo "Installation complete"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR) $(DIST_DIR)
	@$(GO) clean
	@echo "Clean complete"

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	@$(GO) mod download
	@$(GO) mod tidy
	@echo "Dependencies ready"

# Run tests
test:
	@echo "Running tests..."
	@$(GO) test -v ./...

# Format code
format:
	@echo "Formatting code..."
	@gofmt -s -w .
	@echo "Formatting complete"

# Lint code
lint:
	@echo "Running linter..."
	@golangci-lint run --enable-all
	@echo "Linting complete"

# Cross-platform build
cross: deps
	@echo "Building for multiple platforms..."
	@mkdir -p $(DIST_DIR)
	@for platform in $(PLATFORMS); do \
		for arch in $(ARCHITECTURES); do \
			output_name=$(BINARY_NAME)-$$platform-$$arch; \
			if [ $$platform = "windows" ]; then \
				output_name="$$output_name.exe"; \
			fi; \
			echo "Building $$output_name..."; \
			GOOS=$$platform GOARCH=$$arch $(GO) build $(GOFLAGS) \
				-ldflags "$(LDFLAGS)" \
				-o $(DIST_DIR)/$$output_name main.go || true; \
		done; \
	done
	@echo "Cross-platform build complete"

# Run with example (requires example files)
run: build
	@echo "Running example conversion..."
	@$(BUILD_DIR)/$(BINARY_NAME) -dry-run examples/sample.md output.txt

# Docker build (optional)
docker:
	@echo "Building Docker image..."
	@docker build -t note2anki:latest .
	@echo "Docker build complete"

# Generate example config
config:
	@echo "Generating example configuration..."
	@cp config.json.example config.json
	@echo "Configuration file created: config.json"

# Development setup
dev-setup:
	@echo "Setting up development environment..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install golang.org/x/tools/cmd/goimports@latest
	@$(MAKE) deps
	@echo "Development environment ready"

# Version info
version:
	@echo "Note2Anki CLI Tool"
	@echo "Version: 1.0.0"
	@$(GO) version

.DEFAULT_GOAL := help