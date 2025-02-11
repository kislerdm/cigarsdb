default: help

help: ## Prints help message.
	@ grep -h -E '^[a-zA-Z0-9_-].+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[1m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: test
test: ## Runs unit tests.
	@ go test -v -cover -timeout=60s -parallel=10 ./...

.PHONY: build
BIN := fetchdata
VERSION?=$(shell git rev-parse --short HEAD)
build: ## Builds the binary.
	@ go build -o ./bin/$(BIN) -a -gcflags=all="-l -B -C" -ldflags="-w -s -X main.version=$(VERSION)" .

.PHONY: lint
lint: ## Runs the linter.
	@ golangci-lint run
