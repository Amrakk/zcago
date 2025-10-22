GOCMD=go

help: ## display this help screen
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
.PHONY: help

linters-install: ### install golangci-lint if not present
	@golangci-lint --version >/dev/null 2>&1 || { \
		echo "installing linting tools..."; \
		curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh| sh -s v2.0.2; \
	}
.PHONY: linters-install

lint: ### run all linters and auto-fix issues
	golangci-lint run --fix
.PHONY: lint

lint-staged: ### lint and fix only staged changes
	git diff --cached > /tmp/stage.patch
	golangci-lint run --new-from-patch=/tmp/stage.patch
	rm /tmp/stage.patch
.PHONY: lint-staged

format:
	gofumpt -w .
.PHONY: format

test: ### run tests
	$(GOCMD) test -cover -race ./...
.PHONY: test

bench: ### run benchmarks
	$(GOCMD) test -run=NONE -bench=. -benchmem ./...
.PHONY: bench

gen-endpoint: ### generate endpoint skeleton (NAME=messages_send_image)
	$(GOCMD) run ./cmd/gen-endpoint -name=$(NAME)
.PHONY: gen-endpoint

ex-login: ### run login example
	$(GOCMD) run ./examples/login/login.go
.PHONY: ex-login

# Run tests
	
playground: ### Random Go experiments — structs, syntax, whatever
	$(GOCMD) run ./tests/playground.go
.PHONY: playground

loginQR: ### run loginQR test app
	$(GOCMD) run ./cmd/dev-login-qr/main.go
.PHONY: loginQR

api: ### run api test app
	$(GOCMD) run ./cmd/dev-api-test/main.go
.PHONY: api