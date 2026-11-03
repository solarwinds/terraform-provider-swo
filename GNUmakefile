default: testacc

GO_TOOL := go tool -modfile=tools/go.tool.mod
GOLANGCI_LINT_CONFIG := tools/.golangci.yml

.PHONY: lint
lint:
	$(GO_TOOL) golangci-lint run --config $(GOLANGCI_LINT_CONFIG)

# Run acceptance tests
.PHONY: testacc
testacc:
	TF_LOG=info TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 180m
