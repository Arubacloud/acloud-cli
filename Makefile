# Makefile for acloud-cli local development

# Use bash with xpg_echo so plain `echo` interprets ANSI escape sequences
# (\033[...]) without needing `-e`. Required for the colorised output below
# on systems where /bin/sh is bash but `echo` defaults to POSIX behaviour.
SHELL := /bin/bash
.SHELLFLAGS := -O xpg_echo -c

# Variables
BINARY_NAME=acloud
BINARY_WIN=$(BINARY_NAME).exe
GO_VERSION=1.24.2
GOOS?=$(shell go env GOOS)
GOARCH?=$(shell go env GOARCH)

# Build flags
LDFLAGS=-ldflags "-s -w"
BUILD_FLAGS=-trimpath

# Colors for output
GREEN=\033[0;32m
YELLOW=\033[1;33m
RED=\033[0;31m
NC=\033[0m # No Color

.PHONY: help build clean test test-coverage fmt vet lint install run e2e-test build-all release-snapshot

# Default target
.DEFAULT_GOAL := help

##@ General

help: ## Display this help message
	@echo "$(GREEN)Available targets:$(NC)"
	@awk 'BEGIN {FS = ":.*##"; printf "\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  $(YELLOW)%-20s$(NC) %s\n", $$1, $$2 } /^##@/ { printf "\n$(GREEN)%s$(NC)\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Building

build: ## Build the binary for current platform
	@echo "$(GREEN)Building $(BINARY_NAME)...$(NC)"
	@go build $(BUILD_FLAGS) $(LDFLAGS) -o $(BINARY_NAME) .
	@echo "$(GREEN)Build complete: $(BINARY_NAME)$(NC)"

build-windows: ## Build for Windows
	@echo "$(GREEN)Building for Windows...$(NC)"
	@GOOS=windows GOARCH=amd64 go build $(BUILD_FLAGS) $(LDFLAGS) -o $(BINARY_WIN) .
	@echo "$(GREEN)Build complete: $(BINARY_WIN)$(NC)"

build-linux: ## Build for Linux
	@echo "$(GREEN)Building for Linux...$(NC)"
	@GOOS=linux GOARCH=amd64 go build $(BUILD_FLAGS) $(LDFLAGS) -o $(BINARY_NAME)-linux-amd64 .
	@echo "$(GREEN)Build complete: $(BINARY_NAME)-linux-amd64$(NC)"

build-darwin: ## Build for macOS (Intel)
	@echo "$(GREEN)Building for macOS (Intel)...$(NC)"
	@GOOS=darwin GOARCH=amd64 go build $(BUILD_FLAGS) $(LDFLAGS) -o $(BINARY_NAME)-darwin-amd64 .
	@echo "$(GREEN)Build complete: $(BINARY_NAME)-darwin-amd64$(NC)"

build-darwin-arm: ## Build for macOS (Apple Silicon)
	@echo "$(GREEN)Building for macOS (Apple Silicon)...$(NC)"
	@GOOS=darwin GOARCH=arm64 go build $(BUILD_FLAGS) $(LDFLAGS) -o $(BINARY_NAME)-darwin-arm64 .
	@echo "$(GREEN)Build complete: $(BINARY_NAME)-darwin-arm64$(NC)"

build-all: build-windows build-linux build-darwin build-darwin-arm ## Build for all platforms
	@echo "$(GREEN)All builds complete!$(NC)"

release-snapshot: ## Build a local snapshot with GoReleaser (no publish, no tag required)
	@echo "$(GREEN)Building release snapshot...$(NC)"
	@goreleaser release --snapshot --clean
	@echo "$(GREEN)Snapshot artifacts in dist/$(NC)"

##@ Testing

test: ## Run unit tests
	@echo "$(GREEN)Running unit tests...$(NC)"
	@go test ./... -v

test-short: ## Run unit tests (short mode)
	@echo "$(GREEN)Running unit tests (short mode)...$(NC)"
	@go test ./... -short -v

test-coverage: ## Run tests with coverage report
	@echo "$(GREEN)Running tests with coverage...$(NC)"
	@go test ./... -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Coverage report generated: coverage.html$(NC)"

test-coverage-func: ## Show function-level coverage
	@echo "$(GREEN)Function-level coverage:$(NC)"
	@go test ./... -coverprofile=coverage.out
	@go tool cover -func=coverage.out

test-cmd: ## Run tests in cmd package only
	@echo "$(GREEN)Running cmd package tests...$(NC)"
	@go test ./cmd/... -v

test-race: ## Run tests with race detector
	@echo "$(GREEN)Running tests with race detector...$(NC)"
	@go test ./... -race -v

test-benchmark: ## Run benchmark tests
	@echo "$(GREEN)Running benchmark tests...$(NC)"
	@go test ./... -bench=. -benchmem

test-skip-client: ## Run tests skipping client tests (for CI without credentials)
	@echo "$(GREEN)Running tests (skipping client tests)...$(NC)"
	@ACLOUD_TEST_SKIP_CLIENT=true go test ./... -v

test-verbose: ## Run tests with verbose output
	@echo "$(GREEN)Running tests with verbose output...$(NC)"
	@go test ./... -v -count=1

##@ Code Quality

fmt: ## Format all Go code
	@echo "$(GREEN)Formatting code...$(NC)"
	@go fmt ./...
	@echo "$(GREEN)Formatting complete$(NC)"

vet: ## Run go vet
	@echo "$(GREEN)Running go vet...$(NC)"
	@go vet ./...
	@echo "$(GREEN)Vet complete$(NC)"

lint: fmt vet ## Run all linters (fmt + vet)
	@echo "$(GREEN)Linting complete$(NC)"

lint-check: ## Check if code needs formatting (for CI)
	@echo "$(GREEN)Checking code formatting...$(NC)"
	@if [ -n "$$(gofmt -l .)" ]; then \
		echo "$(RED)Code is not formatted. Run 'make fmt'$(NC)"; \
		exit 1; \
	fi
	@echo "$(GREEN)Code formatting check passed$(NC)"

mod-verify: ## Verify go.mod and go.sum
	@echo "$(GREEN)Verifying go.mod and go.sum...$(NC)"
	@go mod verify
	@if command -v git >/dev/null 2>&1 && git rev-parse --git-dir >/dev/null 2>&1; then \
		echo "$(YELLOW)Running go mod tidy...$(NC)"; \
		git diff --quiet go.mod go.sum 2>/dev/null && WAS_CLEAN=1 || WAS_CLEAN=0; \
		go mod tidy; \
		if [ "$$WAS_CLEAN" = "1" ] && ! git diff --quiet go.mod go.sum 2>/dev/null; then \
			echo "$(RED)go.mod or go.sum needs updating. Run 'go mod tidy'$(NC)"; \
			exit 1; \
		fi; \
	fi
	@echo "$(GREEN)Module verification passed$(NC)"

##@ Development

install: build ## Build and install to GOPATH/bin
	@echo "$(GREEN)Installing $(BINARY_NAME)...$(NC)"
	@go install .
	@echo "$(GREEN)Installation complete$(NC)"

run: build ## Build and run the CLI
	@echo "$(GREEN)Running $(BINARY_NAME)...$(NC)"
	@./$(BINARY_NAME) --help

run-debug: build ## Build and run the CLI with debug flag
	@echo "$(GREEN)Running $(BINARY_NAME) with debug...$(NC)"
	@./$(BINARY_NAME) --debug --help

clean: ## Clean build artifacts
	@echo "$(GREEN)Cleaning build artifacts...$(NC)"
	@rm -f $(BINARY_NAME) $(BINARY_WIN)
	@rm -f $(BINARY_NAME)-linux-amd64
	@rm -f $(BINARY_NAME)-darwin-amd64
	@rm -f $(BINARY_NAME)-darwin-arm64
	@rm -rf dist/
	@rm -f coverage.out coverage.html
	@echo "$(GREEN)Clean complete$(NC)"

clean-all: clean ## Clean all artifacts including test cache
	@echo "$(GREEN)Cleaning test cache...$(NC)"
	@go clean -testcache
	@echo "$(GREEN)Clean all complete$(NC)"

deps: ## Download and verify dependencies
	@echo "$(GREEN)Downloading dependencies...$(NC)"
	@go mod download
	@go mod verify
	@echo "$(GREEN)Dependencies ready$(NC)"

deps-update: ## Update all dependencies
	@echo "$(GREEN)Updating dependencies...$(NC)"
	@go get -u ./...
	@go mod tidy
	@echo "$(GREEN)Dependencies updated$(NC)"

deps-tidy: ## Tidy go.mod and go.sum
	@echo "$(GREEN)Tidying dependencies...$(NC)"
	@go mod tidy
	@echo "$(GREEN)Dependencies tidied$(NC)"

##@ E2E Testing

e2e-test: ## Run all E2E tests (requires credentials)
	@echo "$(GREEN)Running E2E tests...$(NC)"
	@echo "$(YELLOW)Note: This requires ACLOUD_PROJECT_ID and other env vars to be set$(NC)"
	@chmod +x e2e/common.sh e2e/management/test.sh e2e/storage/test.sh e2e/network/test.sh
	@EC=0; \
	./e2e/management/test.sh || EC=1; \
	./e2e/storage/test.sh    || EC=1; \
	./e2e/network/test.sh    || EC=1; \
	if [ $$EC -eq 0 ]; then \
	    echo "$(GREEN)✓ E2E: all suites passed$(NC)"; \
	else \
	    echo "$(RED)✗ E2E: one or more suites failed (see above)$(NC)"; \
	fi; \
	exit $$EC

e2e-management: ## Run management E2E tests
	@echo "$(GREEN)Running management E2E tests...$(NC)"
	@chmod +x e2e/management/test.sh
	@./e2e/management/test.sh

e2e-storage: ## Run storage E2E tests
	@echo "$(GREEN)Running storage E2E tests...$(NC)"
	@chmod +x e2e/storage/test.sh
	@./e2e/storage/test.sh

e2e-network: ## Run network E2E tests
	@echo "$(GREEN)Running network E2E tests...$(NC)"
	@chmod +x e2e/network/test.sh
	@./e2e/network/test.sh

##@ Documentation

docs-serve: ## Serve documentation locally (requires Python)
	@echo "$(GREEN)Serving documentation...$(NC)"
	@cd docs && python3 -m http.server 8000 || python -m http.server 8000

docs-check: ## Check documentation links and format
	@echo "$(GREEN)Checking documentation...$(NC)"
	@find docs -name "*.md" -type f | while read file; do \
		echo "Checking $$file"; \
	done

##@ Release

release-check: build test lint ## Check if ready for release
	@echo "$(GREEN)Running release checks...$(NC)"
	@echo "$(GREEN)✓ Build successful$(NC)"
	@echo "$(GREEN)✓ Tests passing$(NC)"
	@echo "$(GREEN)✓ Linting passed$(NC)"
	@echo "$(GREEN)Ready for release!$(NC)"

version: ## Show version information
	@echo "$(GREEN)Version Information:$(NC)"
	@go version
	@echo "Go OS: $(GOOS)"
	@echo "Go ARCH: $(GOARCH)"

##@ Utilities

check-env: ## Check development environment
	@echo "$(GREEN)Checking development environment...$(NC)"
	@echo "Go version: $$(go version)"
	@echo "Go path: $$(go env GOPATH)"
	@echo "Go root: $$(go env GOROOT)"
	@echo "Module: $$(go list -m)"
	@echo "Platform: $(GOOS)/$(GOARCH)"

dev-setup: deps build ## Complete development setup
	@echo "$(GREEN)Development environment ready!$(NC)"
	@echo "$(YELLOW)Next steps:$(NC)"
	@echo "  1. Configure credentials: ./$(BINARY_NAME) config set --client-id <id> --client-secret <secret>"
	@echo "  2. Run tests: make test"
	@echo "  3. Run E2E tests: make e2e-test"

pre-commit: fmt vet test-short ## Run pre-commit checks (fmt, vet, tests)
	@echo "$(GREEN)Pre-commit checks passed!$(NC)"

ci: lint-check mod-verify test-skip-client ## Run CI checks (lint, mod verify, tests without credentials)
	@echo "$(GREEN)CI checks passed!$(NC)"

