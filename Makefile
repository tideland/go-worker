# Tideland Go Worker - Makefile
#
# Copyright (C) 2021-2025 Frank Mueller / Tideland / Germany
#
# All rights reserved. Use of this source code is governed
# by the new BSD license.

# Shell
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

# Set MAKEFLAGS to suppress entering/leaving directory messages
MAKEFLAGS += --no-print-directory

GO ?= go
COVERAGE_FILE := coverage.out
COVERAGE_HTML := coverage.html

# Colors for output
COLOR_RESET := $(shell printf '\033[0m')
COLOR_BOLD := $(shell printf '\033[1m')
COLOR_GREEN := $(shell printf '\033[32m')
COLOR_YELLOW := $(shell printf '\033[33m')
COLOR_BLUE := $(shell printf '\033[34m')

# Default target
.DEFAULT_GOAL := all

# Phony targets
.PHONY: all help tidy build test bench coverage clean

## all: Run complete build process (tidy, build, test)
all: tidy build test
	@echo "$(COLOR_GREEN)$(COLOR_BOLD)✓ All tasks completed successfully$(COLOR_RESET)"

## help: Display this help message
help:
	@echo "$(COLOR_BOLD)Tideland Go Worker - Available Targets:$(COLOR_RESET)"
	@echo ""
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'
	@echo ""
	@echo "$(COLOR_BLUE)Usage: make [target]$(COLOR_RESET)"
	@echo ""

## tidy: Update go.mod and go.sum files
tidy:
	@echo "$(COLOR_YELLOW)→ Tidying Go modules...$(COLOR_RESET)"
	@$(GO) mod tidy
	@$(GO) mod verify
	@echo "$(COLOR_GREEN)✓ Module dependencies updated$(COLOR_RESET)"

## build: Build the package (verify compilation)
build:
	@echo "$(COLOR_YELLOW)→ Building package...$(COLOR_RESET)"
	@$(GO) build -v ./...
	@echo "$(COLOR_GREEN)✓ Build successful$(COLOR_RESET)"

## test: Run all tests
test:
	@echo "$(COLOR_YELLOW)→ Running tests...$(COLOR_RESET)"
	@$(GO) test -v -race ./...
	@echo "$(COLOR_GREEN)✓ Tests passed$(COLOR_RESET)"

## bench: Run benchmarks
bench:
	@echo "$(COLOR_YELLOW)→ Running benchmarks...$(COLOR_RESET)"
	@$(GO) test -bench=. -benchmem -run=^$$ ./...
	@echo "$(COLOR_GREEN)✓ Benchmarks completed$(COLOR_RESET)"

## coverage: Generate test coverage report
coverage:
	@echo "$(COLOR_YELLOW)→ Generating coverage report...$(COLOR_RESET)"
	@$(GO) test -coverprofile=$(COVERAGE_FILE) -covermode=atomic ./...
	@$(GO) tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@$(GO) tool cover -func=$(COVERAGE_FILE) | grep total | awk '{print "Coverage: " $$3}'
	@echo "$(COLOR_GREEN)✓ Coverage report generated: $(COVERAGE_HTML)$(COLOR_RESET)"

## clean: Remove build artifacts and coverage files
clean:
	@echo "$(COLOR_YELLOW)→ Cleaning build artifacts...$(COLOR_RESET)"
	@rm -f $(COVERAGE_FILE) $(COVERAGE_HTML)
	@$(GO) clean -cache -testcache -modcache
	@echo "$(COLOR_GREEN)✓ Clean completed$(COLOR_RESET)"

## ci: Run CI pipeline (used by GitHub Actions)
ci: tidy build test
	@echo "$(COLOR_GREEN)$(COLOR_BOLD)✓ CI pipeline completed$(COLOR_RESET)"
