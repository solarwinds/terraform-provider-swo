.DEFAULT_GOAL := help

GOLANGCI_LINT := go tool -modfile=tools/go.mod golangci-lint
LINT_CONFIG := --config tools/.golangci.yml

.PHONY: help build install test testacc lint lint-fast lint-fix fmt fmt-go fmt-examples generate mod

help: ## Show available commands
	@grep -E '^[a-zA-Z_-]+:.*## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build: ## Build all Go packages
	go build ./...

install: ## Install the provider in GOPATH/bin
	go install .

# Some acceptance tests are not named TestAcc and run even without TF_ACC.
test: ## Run unit tests, excluding acceptance tests
	go test ./... -skip '^(TestAcc|TestAndMetricConditionAlertResource$$|TestOrMetricConditionAlertResource$$)' $(TESTARGS)

testacc: ## Run acceptance tests (creates real resources; requires credentials)
	TF_LOG=info TF_ACC=1 go test ./... -v -timeout 180m $(TESTARGS)

lint: ## Run the pinned Go linter
	$(GOLANGCI_LINT) run $(LINT_CONFIG)

lint-fast: ## Run only fast linters
	$(GOLANGCI_LINT) run $(LINT_CONFIG) --fast-only

lint-fix: ## Run the linter and apply available fixes
	$(GOLANGCI_LINT) run $(LINT_CONFIG) --fix

fmt: fmt-go fmt-examples ## Format Go code and Terraform examples

fmt-go: ## Format Go code using the pinned linter's formatters
	$(GOLANGCI_LINT) fmt $(LINT_CONFIG)

fmt-examples: ## Format Terraform examples (requires Terraform)
	terraform fmt -recursive ./examples/

generate: ## Format examples and generate provider documentation (requires Terraform)
	go generate ./...

mod: ## Tidy and verify provider dependencies
	go mod tidy
	go mod verify
