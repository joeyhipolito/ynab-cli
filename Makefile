# ynab-cli Makefile
# Standalone YNAB CLI tool in Go

.PHONY: all build install test clean help

# Binary name
BINARY_NAME=ynab

# Build directory
BUILD_DIR=bin

# Source entry point
CMD_DIR=./cmd/ynab-cli

# Installation directory
INSTALL_DIR=$(HOME)/bin

# Go build flags
LDFLAGS=-ldflags "-s -w"

# Default target
all: build

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Install the binary to ~/bin
install: build
	@echo "Installing $(BINARY_NAME) to $(INSTALL_DIR)..."
	@mkdir -p $(INSTALL_DIR)
	@ln -sf $(PWD)/$(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "Installation complete: $(INSTALL_DIR)/$(BINARY_NAME)"
	@echo ""
	@echo "Make sure $(INSTALL_DIR) is in your PATH."
	@echo "Add this to your ~/.zshrc or ~/.bashrc if not already present:"
	@echo '  export PATH="$$HOME/bin:$$PATH"'

# Run all tests
test:
	@echo "Running tests..."
	go test ./... -v

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@mkdir -p $(BUILD_DIR)
	go test ./... -coverprofile=$(BUILD_DIR)/coverage.out
	go tool cover -html=$(BUILD_DIR)/coverage.out -o $(BUILD_DIR)/coverage.html
	@echo "Coverage report: $(BUILD_DIR)/coverage.html"

# Run E2E tests (requires YNAB_ACCESS_TOKEN)
test-e2e:
	@echo "Running E2E tests..."
	@if [ -z "$$YNAB_ACCESS_TOKEN" ]; then \
		echo "Error: YNAB_ACCESS_TOKEN environment variable is required"; \
		exit 1; \
	fi
	go test -v -tags=e2e ./e2e_test.go ./e2e_cli_test.go ./e2e_realtime_test.go

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@echo "Clean complete"

# Uninstall the binary
uninstall:
	@echo "Uninstalling $(BINARY_NAME) from $(INSTALL_DIR)..."
	@rm -f $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "Uninstall complete"

# Build for multiple platforms
build-all: clean
	@echo "Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(CMD_DIR)
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(CMD_DIR)
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(CMD_DIR)
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(CMD_DIR)
	@echo "Multi-platform build complete"
	@ls -lh $(BUILD_DIR)

# Run the binary (for quick testing)
run: build
	@$(BUILD_DIR)/$(BINARY_NAME) $(ARGS)

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Vet code for common issues
vet:
	@echo "Vetting code..."
	go vet ./...

# Run linting and checks
lint: fmt vet
	@echo "Linting complete"

# Show help
help:
	@echo "ynab-cli Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make build          Build the binary"
	@echo "  make install        Build and install to ~/bin"
	@echo "  make test           Run all tests"
	@echo "  make test-coverage  Run tests with coverage report"
	@echo "  make test-e2e       Run E2E tests (requires YNAB_ACCESS_TOKEN)"
	@echo "  make clean          Remove build artifacts"
	@echo "  make uninstall      Remove installed binary"
	@echo "  make build-all      Build for all platforms"
	@echo "  make run ARGS='...' Build and run with arguments"
	@echo "  make fmt            Format code"
	@echo "  make vet            Vet code for issues"
	@echo "  make lint           Run fmt and vet"
	@echo "  make help           Show this help message"
	@echo ""
	@echo "Examples:"
	@echo "  make build"
	@echo "  make install"
	@echo "  make test"
	@echo "  make run ARGS='status'"
	@echo "  make run ARGS='balance checking'"
