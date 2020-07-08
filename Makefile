#!/usr/bin/make
# Makefile readme (ru): <http://linux.yaroslavl.ru/docs/prog/gnu_make_3-79_russian_manual.html>
# Makefile readme (en): <https://www.gnu.org/software/make/manual/html_node/index.html#SEC_Contents>

SHELL = /bin/sh

DOCKER_BIN = $(shell command -v docker 2> /dev/null)
APP_NAME = $(notdir $(CURDIR))

.DEFAULT_GOAL : help

help: ## Show this help
	@printf "\033[33m%s:\033[0m\n" 'Available commands'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[32m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

image: ## Build docker image
	$(DOCKER_BIN) build -f ./Dockerfile -t $(APP_NAME):local .
	@printf "\n   \e[30;42m %s \033[0m\n\n" 'Now you can use image like `docker run --rm -p 8080:8080 $(APP_NAME):local ...`'

shell: ## Start shell into container with node
	$(DOCKER_BIN) run --rm -ti -v "$(shell pwd):/src:rw" -w "/src" --user "$(shell id -u):$(shell id -g)" node:12.16.2-alpine sh
