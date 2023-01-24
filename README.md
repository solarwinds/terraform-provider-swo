# Solarwinds Observability Terraform Provider (Terraform Plugin Framework)

## terraform-provider-swo
This provider lets you save clicking in the Solarwinds Observability Platform user interface by allowing you to produce SWO configuration with the rest of your cloud infratructure.

The SWO terraform provider enables the automation of:

* Alerts
* Dashboards
* Notification Services

### Example usage
See `example.tf` [in this repo](https://github.com/solarwindscloud/terraform-provider-swo/blob/master/examples/) to understand how to start using the provider.

### Installing
* Grab the latest release binary from the [Releases page](https://github.com/solarwindscloud/terraform-provider-swo/releases).
* Extract and place the binary into `$HOME/.terraform.d/plugins/github.com/solarwindscloud/terraform-provider-swo/<VERSION>/<ARCH>/terraform-provider-swo` (Replace `<VERSION>` with the version downloaded and `<ARCH>` with the machine architecture (eg. `darwin_amd64` or `darwin_arm64`)
* Set the execute flag on the binary
```
chmod 755 $HOME/.terraform.d/plugins/github.com/solarwindscloud/terraform-provider-swo/<VERSION>/<ARCH>/terraform-provider-swo
```
* You should now be able to write TF code for the Solarwinds Observability Platform with the rest of your infrastructure code.

### Usage Notes
In order for the provider to work in a module, you need to add a required_providers block in your module as such:
```hcl
terraform {
  required_providers {
    swo = {
      source  = "github.com/solarwindscloud/terraform-provider-swo"
      version = ">= 0.0.1"
    }
  }
}
```
This needs to be done because this provider has not been published to the Terraform registry, which is the default location that Terraform will look in when searching for providers.

### Issues/Bugs
Please report bugs and request enhancements in the [Issues area](https://github.com/solarwindscloud/terraform-provider-swo/issues) of this repo.

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.18

## Building The Provider

1. Clone the repository
1. Enter the repository directory
1. Build the provider using the Go `install` command:

```shell
go install
```

## Adding Dependencies

This provider uses [Go modules](https://github.com/golang/go/wiki/Modules).
Please see the Go documentation for the most up to date information about using Go modules.

To add a new dependency `github.com/author/dependency` to your Terraform provider:

```shell
go get github.com/author/dependency
go mod tidy
```

Then commit the changes to `go.mod` and `go.sum`.

## Using the provider

Fill this in for each provider

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `go generate`.

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources, and often cost money to run.

```shell
make testacc
```
