# Makefile for Istio Traffic Handler

.PHONY: help install build run test clean deps lint

# Variables
BINARY_NAME=traffic-handler
NAMESPACE?=default

help: ## Display this help message
	@echo "Istio Traffic Handler - Makefile Commands"
	@echo "=========================================="
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

deps: ## Download Go dependencies
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy
	@echo "✓ Dependencies installed"

build: ## Build the traffic handler binary
	@echo "Building $(BINARY_NAME)..."
	go build -o $(BINARY_NAME) traffic-handler.go
	@echo "✓ Build complete: $(BINARY_NAME)"

run: ## Run the traffic handler
	@echo "Running traffic handler (namespace: $(NAMESPACE))..."
	NAMESPACE=$(NAMESPACE) go run traffic-handler.go

run-binary: build ## Build and run the binary
	@echo "Running $(BINARY_NAME) (namespace: $(NAMESPACE))..."
	NAMESPACE=$(NAMESPACE) ./$(BINARY_NAME)

test: ## Run tests
	@echo "Running tests..."
	go test -v ./...

lint: ## Run linter
	@echo "Running linter..."
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run

clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -f $(BINARY_NAME)
	go clean
	@echo "✓ Clean complete"

install-istio: ## Install Istio on the cluster
	@echo "Installing Istio..."
	@if [ "$$(uname -s)" = "Darwin" ] || [ "$$(uname -s)" = "Linux" ]; then \
		bash setup.sh; \
	else \
		powershell -File setup.ps1; \
	fi

# Example usage targets
example-canary: build ## Run canary deployment example
	@echo "Running canary deployment example..."
	NAMESPACE=$(NAMESPACE) ./$(BINARY_NAME)

verify-istio: ## Verify Istio installation
	@echo "Verifying Istio installation..."
	@kubectl get pods -n istio-system
	@echo ""
	@echo "VirtualServices:"
	@kubectl get virtualservices -A
	@echo ""
	@echo "DestinationRules:"
	@kubectl get destinationrules -A

check-traffic: ## Check traffic distribution
	@echo "Checking traffic distribution for namespace: $(NAMESPACE)"
	@kubectl get virtualservices -n $(NAMESPACE) -o yaml

list-vs: ## List all VirtualServices
	@kubectl get virtualservices -A

list-dr: ## List all DestinationRules
	@kubectl get destinationrules -A

delete-vs: ## Delete VirtualService (usage: make delete-vs NAME=myapp-canary)
	@if [ -z "$(NAME)" ]; then \
		echo "Error: Please specify NAME=<virtualservice-name>"; \
		exit 1; \
	fi
	kubectl delete virtualservice $(NAME) -n $(NAMESPACE)

delete-dr: ## Delete DestinationRule (usage: make delete-dr NAME=myapp)
	@if [ -z "$(NAME)" ]; then \
		echo "Error: Please specify NAME=<destinationrule-name>"; \
		exit 1; \
	fi
	kubectl delete destinationrule $(NAME) -n $(NAMESPACE)

format: ## Format Go code
	@echo "Formatting code..."
	go fmt ./...
	@echo "✓ Code formatted"

# Development targets
dev-setup: deps ## Setup development environment
	@echo "Setting up development environment..."
	@which golangci-lint > /dev/null || go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "✓ Development environment ready"

all: deps lint test build ## Run all checks and build

.DEFAULT_GOAL := help
