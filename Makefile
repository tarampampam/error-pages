#!/usr/bin/make
# Makefile readme (ru): <http://linux.yaroslavl.ru/docs/prog/gnu_make_3-79_russian_manual.html>
# Makefile readme (en): <https://www.gnu.org/software/make/manual/html_node/index.html#SEC_Contents>

SHELL = /bin/sh
LDFLAGS = "-s -w -X github.com/tarampampam/error-pages/internal/version.version=$(shell git rev-parse HEAD)"

DC_RUN_ARGS = --rm --user "$(shell id -u):$(shell id -g)"
APP_NAME = $(notdir $(CURDIR))

.PHONY : help \
         image dive build fmt lint gotest int-test test shell \
         up down restart \
         clean
.DEFAULT_GOAL : help
.SILENT : lint gotest

# This will output the help for each task. thanks to https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
help: ## Show this help
	@printf "\033[33m%s:\033[0m\n" 'Available commands'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[32m%-11s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

image: ## Build docker image with app
	docker build -f ./Dockerfile -t $(APP_NAME):local .
	docker run --rm $(APP_NAME):local version
	@printf "\n   \e[30;42m %s \033[0m\n\n" 'Now you can use image like `docker run --rm -p "8080:8080/tcp" $(APP_NAME):local ...`';

dive: image ## Explore the docker image
	docker run --rm -it -v "/var/run/docker.sock:/var/run/docker.sock:ro" wagoodman/dive:latest $(APP_NAME):local

build: ## Build app binary file
	docker-compose run $(DC_RUN_ARGS) -e "CGO_ENABLED=0" --no-deps app go build -trimpath -ldflags $(LDFLAGS) -o ./error-pages ./cmd/error-pages/

fmt: ## Run source code formatter tools
	docker-compose run $(DC_RUN_ARGS) -e "GO111MODULE=off" --no-deps app sh -c 'go get golang.org/x/tools/cmd/goimports && $$GOPATH/bin/goimports -d -w .'
	docker-compose run $(DC_RUN_ARGS) --no-deps app gofmt -s -w -d .
	docker-compose run $(DC_RUN_ARGS) --no-deps app go mod tidy

lint: ## Run app linters
	docker-compose run --rm --no-deps golint golangci-lint run

gotest: ## Run app tests
	docker-compose run $(DC_RUN_ARGS) --no-deps app go test -v -race -timeout 10s ./...

int-test: ## Run integration tests (docs: https://hurl.dev/docs/man-page.html#options)
	docker-compose run --rm hurl --color --no-output --fail-at-end --file-root ./test/hurl --variable host=web --variable port=8080 --summary ./test/hurl/*.hurl

test: lint gotest int-test ## Run app tests and linters

shell: ## Start shell into container with golang
	docker-compose run $(DC_RUN_ARGS) app bash

up: ## Create and start containers
	docker-compose up --detach web
	@printf "\n   \e[30;42m %s \033[0m\n\n" 'Navigate your browser to ⇒ http://127.0.0.1:8080';

down: ## Stop all services
	docker-compose down -t 5

restart: down up ## Restart all containers

clean: ## Make clean
	docker-compose down -v -t 1
	-docker rmi $(APP_NAME):local -f
