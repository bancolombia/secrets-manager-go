GO ?= go
PKG ?= ./...
COVERAGE_OUT ?= coverage.out
COVERAGE_HTML ?= coverage.html

.DEFAULT_GOAL := help

.PHONY: help tidy deps fmt fmt-check vet lint build test test-cover cover-report cover-html check ci clean

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-14s\033[0m %s\n", $$1, $$2}'

tidy: ## Tidy go.mod/go.sum
	$(GO) mod tidy

deps: ## Download module dependencies
	$(GO) mod download

fmt: ## Format all Go source files
	$(GO) fmt $(PKG)

fmt-check: ## Fail if any file is not gofmt-formatted
	@unformatted="$$(gofmt -l .)"; \
	if [ -n "$$unformatted" ]; then \
		echo "gofmt needed for:"; echo "$$unformatted"; exit 1; \
	fi

vet: ## Run go vet
	$(GO) vet $(PKG)

lint: fmt-check vet ## Run format check and go vet

build: ## Compile all packages
	$(GO) build ./...

test: ## Run unit tests
	$(GO) test $(PKG)

test-cover: ## Run tests with race detector and coverage (CI flags)
	$(GO) test -race -covermode=atomic -coverprofile=$(COVERAGE_OUT) $(PKG)

cover-report: test-cover ## Show per-function coverage summary
	$(GO) tool cover -func=$(COVERAGE_OUT)

cover-html: test-cover ## Generate HTML coverage report
	$(GO) tool cover -html=$(COVERAGE_OUT) -o $(COVERAGE_HTML)
	@echo "Coverage report written to $(COVERAGE_HTML)"

check: lint build test ## Run lint, build and tests

ci: lint build test-cover ## Run the same checks as the CI pipeline

clean: ## Remove build and coverage artifacts
	$(GO) clean $(PKG)
	rm -f $(COVERAGE_OUT) $(COVERAGE_HTML)
