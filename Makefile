GOCMD ?= go

.DEFAULT_GOAL := help

##@ General
help: ## Show this help message with available targets
	@awk 'BEGIN {FS=":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\n"} \
	/^[a-zA-Z0-9_.-]+:.*##/ { printf "  \033[36m%-22s\033[0m %s\n", $$1, $$2 } \
	/^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0,5) }' $(MAKEFILE_LIST)
.PHONY: help

##@ Linting & Formatting
linters-install: ## Install golangci-lint locally if missing (latest)
	@command -v golangci-lint >/dev/null 2>&1 || { \
		echo "Installing golangci-lint (latest)..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh \
		  | sh -s -- -b $$(go env GOPATH)/bin; \
	}
.PHONY: linters-install

lint: ## Run all linters and auto-fix issues (golangci-lint run --fix)
	golangci-lint run --fix
.PHONY: lint

lint-staged: ## Lint only staged changes (pre-commit style)
	@tmp=$$(mktemp); \
	git diff --cached > $$tmp; \
	golangci-lint run --new-from-patch="$$tmp"; \
	rm -f "$$tmp"
.PHONY: lint-staged

format: ## Format code with gofumpt (in-place)
	gofumpt -w .
.PHONY: format

##@ Test & Bench
test: ## Run unit tests with race detector and coverage
	$(GOCMD) test -race -cover ./...
.PHONY: test

bench: ## Run benchmarks with memory stats
	$(GOCMD) test -run=NONE -bench=. -benchmem ./...
.PHONY: bench

playground: ## Execute quick playground program (tests/playground.go)
	$(GOCMD) run ./tests/playground.go
.PHONY: playground

##@ Dev Apps
api: ## Run API dev test app (cmd/dev-api-test)
	$(GOCMD) run ./cmd/dev-api-test
.PHONY: api

loginQR: ## Run login-QR dev app (cmd/dev-login-qr)
	$(GOCMD) run ./cmd/dev-login-qr
.PHONY: loginQR

listener: ## Run listener dev app (cmd/dev-listener)
	$(GOCMD) run ./cmd/dev-listener
.PHONY: listener

gen-endpoint: ## Generate endpoint skeleton (usage: make gen-endpoint NAME=SendMessage)
	$(GOCMD) run ./cmd/gen-endpoint -name=$(NAME)
.PHONY: gen-endpoint

##@ Examples
ex-login: ## Run login example (examples/login/login.go)
	$(GOCMD) run ./examples/login/login.go
.PHONY: ex-login

ex-echobot: ## Run echo-bot example (examples/echobot/echobot.go)
	$(GOCMD) run ./examples/echobot/echobot.go
.PHONY: ex-echobot
