# Solarwinds Observability Terraform Provider (Terraform Plugin Framework)

## terraform-provider-swo

This provider lets you manage your SolarWinds Observability configuration (e.g. alerts, websites, API tokens, etc.) using Terraform. Its
Terraform registry page can be found [here](https://registry.terraform.io/providers/solarwinds/swo/latest).

## Using the provider

Full documentation for using the provider can be found at the [Terraform Registry](https://registry.terraform.io/providers/solarwinds/swo/latest/docs) or in the `/docs` folder of this repo.

### Example usage

See `example.tf` [in this repo](https://github.com/solarwinds/terraform-provider-swo/blob/master/examples/) to understand how to start using the provider.

## Issues/Bugs

Please report bugs and request enhancements in the [Issues area](https://github.com/solarwinds/terraform-provider-swo/issues) of this repo.

## Developing the Provider

### Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0.7
- [Go](https://golang.org/doc/install) >= 1.22

### Building

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `go generate`.

Before running Acceptance tests create a .env in the root directory and set your SWO_BASE_URL and SWO_API_TOKEN variables.

In order to run the full suite of Acceptance tests, run `make testacc`.

_Note:_ Acceptance tests create real resources, and often cost money to run.

```shell
make testacc
```

### Adding Dependencies

This provider uses [Go modules](https://github.com/golang/go/wiki/Modules).
Please see the Go documentation for the most up to date information about using Go modules.

To add a new dependency `github.com/author/dependency` to your Terraform provider:

```shell
go get github.com/author/dependency
go mod tidy
```

Then commit the changes to `go.mod` and `go.sum`.

## Terraform Debugging

Follow this: https://opencredo.com/blogs/running-a-terraform-provider-with-a-debugger/
For vscode setup a .vscode/launch.json file that looks like this

```
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "debug tf",
            "type": "go",
            "request": "launch",
            "args": ["--debug"],
            "program": "main.go"
        }
    ],
}
```

Run the debugger. If everything was set up correctly, the plugin will output a message to the Debug Console telling you to set the TF_REATTACH_PROVIDERS environment variable.

Below you see an example of what should be displayed in your Debug Console if the provider stated up correctly. Copy the variable in your Debug Console and export it in your terminal. Don't export what is in the example below.

### Example:

```
Provider started. To attach Terraform CLI, set the TF_REATTACH_PROVIDERS environment variable with the following:

	TF_REATTACH_PROVIDERS='{"github.com/solarwinds/swo":{"Protocol":"grpc","ProtocolVersion":6,"Pid":50111,"Test":true,"Addr":{"Network":"unix","String":"/var/folders/86/21z0bd_x39g177h5nfw2l8w80000gq/T/plugin1234"}}}'
```

export the `TF_REATTACH_PROVIDERS` you get back.

If you're running Apple M1 chip:

```
export GODEBUG=asyncpreemptoff=1
```

Run Terraform:

```
terraform init -upgrade
```
