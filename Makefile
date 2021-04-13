#!/usr/bin/make
# Makefile readme (ru): <http://linux.yaroslavl.ru/docs/prog/gnu_make_3-79_russian_manual.html>
# Makefile readme (en): <https://www.gnu.org/software/make/manual/html_node/index.html#SEC_Contents>

SHELL = /bin/sh

DC_RUN_ARGS = --rm --user "$(shell id -u):$(shell id -g)"
APP_NAME = $(notdir $(CURDIR))

.PHONY : help install gen preview
.DEFAULT_GOAL : help

help: ## Show this help
	@printf "\033[33m%s:\033[0m\n" 'Available commands'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[32m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

install: ## Install all dependencies
	docker-compose run $(DC_RUN_ARGS) -w "/src/generator" node yarn install --frozen-lockfile --no-progress --non-interactive

gen: ## Generate error pages
	docker-compose run $(DC_RUN_ARGS) node nodejs ./generator/generator.js -i -c ./config.json -o ./out

preview: ## Build docker image and start preview
	docker build -f ./Dockerfile -t $(APP_NAME):local .
	@printf "\n   \e[30;42m %s \033[0m\n\n" 'Now open in your favorite browser <http://127.0.0.1:8081> and press CTRL+C for stopping'
	docker run --rm -i -p 8081:8080 -e "TEMPLATE_NAME=random" $(APP_NAME):local

shell: ## Start shell into container with node
	docker-compose run $(DC_RUN_ARGS) node sh
