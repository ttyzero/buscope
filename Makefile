.PHONY: help build test vet fmt tidy clean

help: ## Show available targets
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-12s\033[0m %s\n", $$1, $$2}'

build: ## Build the buscope binary
	go build -ldflags "-s -w -X main.version=$${VERSION:-dev}" -o buscope ./cmd/buscope

test: ## Run tests (with race detector)
	go test -race ./...

vet: ## go vet
	go vet ./...

fmt: ## go fmt
	gofmt -w .

tidy: ## go mod tidy
	go mod tidy

clean: ## Remove built binary
	rm -f buscope
