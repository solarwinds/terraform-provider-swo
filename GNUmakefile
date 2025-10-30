default: testacc

# Run acceptance tests
.PHONY: testacc
testacc:
	TF_LOG=info TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 180m
