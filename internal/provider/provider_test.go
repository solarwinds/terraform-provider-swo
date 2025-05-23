package provider

import (
	"fmt"
	"log"
	"path/filepath"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/joho/godotenv"
)

func providerConfig() string {
	//Source the .env file in the root dir if it exists.
	//Set SWO_API_TOKEN, SWO_BASE_URL in the .env
	envPath := filepath.Join("..", "..", ".env")
	if err := godotenv.Load(envPath); err != nil {
		log.Printf("Warning: Couldn't load .env file: %v", err)
	}

	return fmt.Sprintln(`provider "swo" {
		request_timeout = 10
	}`)
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
