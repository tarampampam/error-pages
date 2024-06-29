#!/usr/bin/make

DC_RUN_ARGS = --rm --user "$(shell id -u):$(shell id -g)"

.DEFAULT_GOAL : help

help: ## Show this help
	@printf "\033[33m%s:\033[0m\n" 'Available commands'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[32m%-11s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: up
up: ## Start the application in watch mode
	docker compose kill web --remove-orphans 2>/dev/null || true
	docker compose up --detach --wait web
	$$SHELL -c "\
		trap 'docker compose down --remove-orphans --timeout 30' EXIT; \
		docker compose watch --no-up web \
	"

.PHONY: down
down: ## Stop the application
	docker compose down --remove-orphans

.PHONY: shell
shell: ## Start shell into development environment
	docker compose run -ti $(DC_RUN_ARGS) develop bash

.PHONY: test
test: ## Run tests
	docker compose run $(DC_RUN_ARGS) develop gotestsum --format pkgname -- -race -timeout 2m ./...

.PHONY: lint
lint: ## Run linters
	docker compose run $(DC_RUN_ARGS) develop golangci-lint run

.PHONY: gen
gen: ## Generate code
	docker compose run $(DC_RUN_ARGS) develop go generate ./...
