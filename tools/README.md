# Development tools

Go tools are pinned in `go.tool.mod`, separate from the provider's dependencies. Run these commands from the repository root; Go downloads the tools automatically.

```sh
make help       # List development commands
make lint       # Check Go code
make lint-fast  # Run fast linters only
make lint-fix   # Apply available lint fixes
make fmt        # Format Go code and Terraform examples (requires Terraform)
make generate   # Format examples and generate documentation (requires Terraform)
make mod        # Tidy and verify provider dependencies
```

## Manage tools

Add or update a tool using its command package and a specific version:

```sh
go -C tools get -modfile=go.tool.mod -tool github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0
```

To remove a tool:

```sh
go -C tools get -modfile=go.tool.mod -tool <command-package>@none
```

Commit the resulting changes to `tools/go.tool.mod` and `tools/go.tool.sum`. Keep development tools out of the root `go.mod`.

GoReleaser is pinned in the release workflow; it is not in the Go tools module because its dependencies conflict with the linter.
