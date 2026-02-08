.PHONY: help build test clean fmt lint run

help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build: ## Build the application
	go build -o bin/server cmd/server/main.go

test: ## Run tests
	go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...

clean: ## Clean build artifacts
	rm -rf bin/
	rm -f coverage.txt coverage.html

fmt: ## Format code
	go fmt ./...
	goimports -w .

lint: ## Run linter
	golangci-lint run

run: ## Run the server
	go run cmd/server/main.go

deps: ## Download dependencies
	go mod download
	go mod tidy

coverage: test ## Generate coverage report
	go tool cover -html=coverage.txt -o coverage.html
