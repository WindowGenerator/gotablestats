include .env

.DEFAULT_GOAL:=help
.PHONY: help
help:  ## Display this help
	$(info $(PROJECT_NAME) v$(VERSION))
	awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z0-9_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

reports:
	mkdir reports

.PHONY: test
test: reports ## Run the unit tests
	$(info [+] Running unit tests)
	go test ./... -cover -coverprofile=reports/coverage.out

.PHONY: coverage
coverage: ## Create coverage report
	$(info [+] Running code coverage)
	go tool cover -html=reports/coverage.out -o reports/coverage.html
	open reports/coverage.html

.PHONY: complexity
complexity: gocognit ## Check codebase for complexity

.PHONY: gocognit
gocognit:
	$(info Calculating gocognit score)
	gocognit -over 3 -avg .


.PHONY: lint
lint: ## Check codebase for complexity
	go fmt ./...
	go vet ./...
	staticcheck ./...
	gocritic check ./...

.PHONY: build
build: ## Build the project
	go build

$(VERBOSE).SILENT: