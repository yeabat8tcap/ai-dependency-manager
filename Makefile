# AI Dependency Manager Makefile

# Build variables
BINARY_NAME=ai-dep-manager
VERSION?=dev
GIT_COMMIT?=$(shell git rev-parse --short HEAD)
BUILD_DATE?=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-X github.com/8tcapital/ai-dep-manager/cmd.Version=${VERSION} -X github.com/8tcapital/ai-dep-manager/cmd.GitCommit=${GIT_COMMIT} -X github.com/8tcapital/ai-dep-manager/cmd.BuildDate=${BUILD_DATE}"

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Build targets
.PHONY: all build clean test deps install run help build-frontend build-full-stack serve

all: deps test build-full-stack

build: ## Build the Go binary only
	$(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_NAME) .

build-frontend: ## Build the Angular frontend
	@echo "🚀 Building Angular frontend..."
	@./scripts/build-frontend.sh

build-full-stack: build-frontend ## Build the complete full-stack application
	@echo "🔨 Building unified full-stack application..."
	$(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_NAME) .
	@echo "✅ Full-stack build complete!"
	@echo "📊 Binary size: $$(du -h bin/$(BINARY_NAME) | cut -f1)"
	@echo "🚀 Run with: ./bin/$(BINARY_NAME) serve"

serve: build-full-stack ## Build and start the unified web server
	@echo "🌐 Starting AI Dependency Manager Web Server..."
	./bin/$(BINARY_NAME) serve

build-linux: ## Build for Linux
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_NAME)-linux .

build-windows: ## Build for Windows
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_NAME)-windows.exe .

build-darwin: ## Build for macOS
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin .

clean: ## Clean build artifacts
	$(GOCLEAN)
	rm -rf bin/

test: ## Run all tests
	$(GOTEST) -v ./...

test-unit: ## Run unit tests only
	$(GOTEST) -v -short ./internal/...

test-integration: ## Run integration tests
	$(GOTEST) -v -run Integration ./test/integration/...

test-e2e: ## Run end-to-end tests
	$(GOTEST) -v -timeout=10m ./test/e2e/...

test-coverage: ## Run tests with coverage
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

test-coverage-unit: ## Run unit tests with coverage
	$(GOTEST) -v -short -coverprofile=coverage-unit.out ./internal/...
	$(GOCMD) tool cover -html=coverage-unit.out -o coverage-unit.html

test-race: ## Run tests with race detection
	$(GOTEST) -v -race ./...

test-bench: ## Run benchmark tests
	$(GOTEST) -v -bench=. -benchmem ./...

test-security: ## Run security tests
	$(GOTEST) -v -run Security ./...

test-all: test-unit test-integration test-e2e ## Run all test suites

deps: ## Download dependencies
	$(GOMOD) download
	$(GOMOD) tidy

install: build ## Install the binary to $GOPATH/bin
	cp bin/$(BINARY_NAME) $(GOPATH)/bin/

run: ## Run the application
	$(GOCMD) run . status

run-dev: ## Run with development settings
	$(GOCMD) run . --log-level debug status

lint: ## Run linter
	golangci-lint run

format: ## Format code
	$(GOCMD) fmt ./...

vet: ## Run go vet
	$(GOCMD) vet ./...

security: ## Run security checks
	gosec ./...

docker-build: ## Build Docker image
	docker build -t $(BINARY_NAME):$(VERSION) .

docker-run: ## Run in Docker container
	docker run --rm -it $(BINARY_NAME):$(VERSION)

release: clean deps test build-linux build-windows build-darwin ## Build release binaries

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
