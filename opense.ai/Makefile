.PHONY: build build-go build-web test lint run clean fmt vet tidy dev ui-dev ui-build ui-test ui-lint docker

# Go variables
BINARY_NAME=openseai
GO=go
GOFLAGS=-v
LDFLAGS=-ldflags "-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT?=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE?=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)

# Directories
CMD_DIR=./cmd/openseai
BUILD_DIR=./build
WEB_DIR=./web

## build: Build web frontend + Go binary (full production build)
build: build-web build-go

## build-go: Build only the Go binary (assumes web/out/ exists)
build-go:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)

## build-web: Build Next.js static export into web/out/
build-web:
	@echo "Building web frontend..."
	cd $(WEB_DIR) && npm run build
	@echo "Web build output: $(WEB_DIR)/out/"

## run: Run the application
run: build
	$(BUILD_DIR)/$(BINARY_NAME)

## dev: Run with hot reload (requires air)
dev:
	@which air > /dev/null 2>&1 || (echo "Installing air..." && go install github.com/air-verse/air@latest)
	air

## test: Run all tests
test:
	$(GO) test ./... -v -race -cover

## test-short: Run tests without race detector
test-short:
	$(GO) test ./... -v -short

## coverage: Generate test coverage report
coverage:
	$(GO) test ./... -coverprofile=coverage.out
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

## lint: Run golangci-lint
lint:
	@which golangci-lint > /dev/null 2>&1 || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run ./...

## fmt: Format Go source code
fmt:
	$(GO) fmt ./...
	goimports -w .

## vet: Run go vet
vet:
	$(GO) vet ./...

## tidy: Tidy go.mod
tidy:
	$(GO) mod tidy

## clean: Remove build artifacts
clean:
	@rm -rf $(BUILD_DIR)
	@rm -rf $(WEB_DIR)/out
	@rm -f coverage.out coverage.html

## ui-dev: Start Next.js dev server
ui-dev:
	cd $(WEB_DIR) && npm run dev

## ui-build: Build Next.js production bundle
ui-build:
	cd $(WEB_DIR) && npm run build

## ui-test: Run frontend tests
ui-test:
	cd $(WEB_DIR) && npm test

## ui-lint: Lint frontend code
ui-lint:
	cd $(WEB_DIR) && npm run lint

## docker: Build Docker images
docker:
	docker compose build

## docker-up: Start all services
docker-up:
	docker compose up -d

## docker-down: Stop all services
docker-down:
	docker compose down

## bench: Run Go benchmarks
bench:
	$(GO) test -bench=. -benchmem ./internal/analysis/technical/... ./internal/financeql/...

## e2e: Run Playwright E2E tests
e2e:
	cd $(WEB_DIR) && npx playwright test

## e2e-headed: Run E2E tests in headed mode
e2e-headed:
	cd $(WEB_DIR) && npx playwright test --headed

## setup: Run development environment setup
setup:
	bash scripts/setup.sh

## nse-test: Run NSE data source smoke tests
nse-test:
	bash scripts/test_nse.sh

## help: Show this help
help:
	@echo "OpeNSE.ai — Agentic AI for NSE Stock Analysis & Trading"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@sed -n 's/^## //p' $(MAKEFILE_LIST) | column -t -s ':'
