.PHONY: build build-all clean help

# Binary name
BINARY_NAME=gitext
BUILD_DIR=build

# Build info
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date +%Y-%m-%d_%H:%M:%S)
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

# Architectures to build for
ARCHES=amd64 arm64
OSES=linux darwin windows

help: ## Show help message
	@echo "Available commands:"
	@echo "  make build-all    - Build for all architectures and operating systems"
	@echo "  make build        - Build for current platform"
	@echo "  make clean        - Clean build directory"
	@echo "  make help         - Show this help message"

build: ## Build for current platform
	@echo "Building for current platform..."
	@mkdir -p $(BUILD_DIR)
	@go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/gitext

build-all: ## Build for all architectures and operating systems
	@echo "Building for all architectures..."
	@mkdir -p $(BUILD_DIR)
	@for os in $(OSES); do \
		for arch in $(ARCHES); do \
			echo "Building for $$os/$$arch..."; \
			GOOS=$$os GOARCH=$$arch go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-$$os-$$arch$(if $(findstring windows,$$os),.exe,) ./cmd/gitext; \
		done \
	done
	@echo "Build complete! Binaries are in $(BUILD_DIR)/"

clean: ## Clean build directory
	@echo "Cleaning build directory..."
	@rm -rf $(BUILD_DIR)
	@echo "Clean complete!"
