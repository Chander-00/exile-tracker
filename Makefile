.PHONY: build run test clean deps help docker-build docker-up docker-down docker-logs shrink-tree

# Binary name
BINARY_NAME=exile-tracker
BUILD_DIR=bin

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Tool commands
TEMPL=templ

# Build flags
LDFLAGS=-ldflags "-X main.Version=$(shell git describe --tags --always --dirty)"

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

migrate: ## Run DB migrations
	goose -dir ./migrations sqlite3 ./data.db up

generate: ## Generate templ files
	$(TEMPL) generate

deps: ## Download dependencies
	$(GOMOD) tidy
	$(GOMOD) download

build: deps generate migrate ## Build the application
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd

run: build ## Build and run the application
	@echo "Running $(BINARY_NAME)..."
	./$(BUILD_DIR)/$(BINARY_NAME)

dev: generate migrate ## Run the application in development mode
	@echo "Running in development mode..."
	$(GOCMD) run ./cmd

test: ## Run tests
	@echo "Running tests..."
	$(GOTEST) -v ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out

clean: ## Clean build artifacts
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -f coverage.out

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker compose build

docker-up: ## Start all services
	@echo "Starting services..."
	docker compose up -d --build

docker-down: ## Stop all services
	@echo "Stopping services..."
	docker compose down

docker-logs: ## Show logs from all services
	docker compose logs -f

lint: ## Run linter
	@echo "Running linter..."
	golangci-lint run

shrink-tree: ## Shrink data.json into static/tree.json for the passive tree renderer
	go run cmd/tools/shrinktree/main.go data.json cmd/web/static/tree.json

fmt: ## Format code
	@echo "Formatting code..."
	$(GOCMD) fmt ./...

