#!/usr/bin/make
# Makefile readme (ru): <http://linux.yaroslavl.ru/docs/prog/gnu_make_3-79_russian_manual.html>
# Makefile readme (en): <https://www.gnu.org/software/make/manual/html_node/index.html#SEC_Contents>

SHELL = /bin/sh

DC_RUN_ARGS = --rm --user "$(shell id -u):$(shell id -g)"
APP_NAME = $(notdir $(CURDIR))

.PHONY : help install shell lint test build
.DEFAULT_GOAL : help

help: ## Show this help
	@printf "\033[33m%s:\033[0m\n" 'Available commands'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[32m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

install: ## Install all dependencies
	docker-compose run $(DC_RUN_ARGS) node yarn install --no-progress --non-interactive

shell: ## Start shell into a container with node
	docker-compose run $(DC_RUN_ARGS) node sh

lint: ## Execute provided linters
	docker-compose run $(DC_RUN_ARGS) node yarn lint

test: ## Execute provided tests
	docker-compose run $(DC_RUN_ARGS) node yarn test

build: ## Build frontend
	docker-compose run $(DC_RUN_ARGS) node yarn build
