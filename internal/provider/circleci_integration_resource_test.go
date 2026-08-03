package provider

import (
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var (
	errResourceNotFound      = errors.New("resource not found")
	errReceiverUrlMismatch   = errors.New("receiver_url mismatch")
)

func TestAccCircleCIIntegrationResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		IsUnitTest:               true,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccCircleCIIntegrationResourceConfig("test-circleci-one"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("swo_circleciintegration.test", "id"),
					resource.TestCheckResourceAttr("swo_circleciintegration.test", "name", "test-circleci-one"),
					resource.TestCheckResourceAttrSet("swo_circleciintegration.test", "secret_token"),
					resource.TestCheckResourceAttrSet("swo_circleciintegration.test", "receiver_url"),
					resource.TestCheckResourceAttr("swo_circleciintegration.test", "receiver_base", defaultReceiverBase),
					checkReceiverUrlMatchesBase("swo_circleciintegration.test"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "swo_circleciintegration.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"api_token"},
			},
			// Update and Read testing
			{
				Config: testAccCircleCIIntegrationResourceConfig("test-circleci-two"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("swo_circleciintegration.test", "name", "test-circleci-two"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccCircleCIIntegrationResource_CustomReceiverBase(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		IsUnitTest:               true,
		Steps: []resource.TestStep{
			{
				Config: testAccCircleCIIntegrationResourceConfigWithReceiverBase("test-circleci-custom", "https://custom.example.com/webhook"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("swo_circleciintegration.test", "id"),
					resource.TestCheckResourceAttr("swo_circleciintegration.test", "name", "test-circleci-custom"),
					resource.TestCheckResourceAttr("swo_circleciintegration.test", "receiver_base", "https://custom.example.com/webhook"),
				),
			},
		},
	})
}

func testAccCircleCIIntegrationResourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
	resource "swo_circleciintegration" "test" {
		name = %[1]q
	}`, name)
}

func testAccCircleCIIntegrationResourceConfigWithReceiverBase(name, receiverBase string) string {
	return providerConfig() + fmt.Sprintf(`
	resource "swo_circleciintegration" "test" {
		name          = %[1]q
		receiver_base = %[2]q
	}`, name, receiverBase)
}

func checkReceiverUrlMatchesBase(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("%w: %s", errResourceNotFound, resourceName)
		}

		id := rs.Primary.Attributes["id"]
		receiverBase := rs.Primary.Attributes["receiver_base"]
		receiverUrl := rs.Primary.Attributes["receiver_url"]
		expected := fmt.Sprintf("%s?state=%s", receiverBase, id)

		if receiverUrl != expected {
			return fmt.Errorf("%w: got %q, expected %q (receiver_base + ?state= + id)", errReceiverUrlMismatch, receiverUrl, expected)
		}
		return nil
	}
}
