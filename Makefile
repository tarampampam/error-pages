#!/usr/bin/make

.DEFAULT_GOAL : build

gen: ## Generate code
	go generate ./...

build: gen ## Build the application
	CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o ./error-pages ./cmd/error-pages/

test: ## Run tests
	go test -race ./...

lint: ## Run linters (requires https://github.com/golangci/golangci-lint installed)
	golangci-lint run

up: build ## Start the application at http://localhost:8080
	./error-pages --log-level debug serve --show-details --proxy-headers=X-Foo,Bar,Baz_blah
