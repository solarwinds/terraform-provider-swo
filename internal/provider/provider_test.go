package provider

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

func providerConfig() string {
	apiToken := os.Getenv("SWO_API_TOKEN")
	if apiToken == "" {
		log.Fatal("SWO_API_TOKEN must be set for acceptance tests")
	}

	baseURL := os.Getenv("SWO_BASE_URL")
	if baseURL == "" {
		log.Fatal("SWO_BASE_URL must be set for acceptance tests")
	}

	return fmt.Sprintf(`provider "swo" {
	api_token = "%s"
	request_timeout = 10
	base_url = "%s"
}`, apiToken, baseURL)
}

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"swo": providerserver.NewProtocol6WithError(New("test", nil)()),
}

func testAccPreCheck(t *testing.T) {
	// You can add code here to run prior to any test case execution, for example assertions
	// about the appropriate environment variables being set are common to see in a pre-check
	// function.
}
