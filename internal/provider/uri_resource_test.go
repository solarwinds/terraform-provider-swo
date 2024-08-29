package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccUriResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		IsUnitTest:               true,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccUriResourceConfig("test one"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("swo_uri.test", "id"),
					resource.TestCheckResourceAttr("swo_uri.test", "name", "test one"),
					resource.TestCheckResourceAttr("swo_uri.test", "host", "www.solarwinds.com"),
					resource.TestCheckResourceAttr("swo_uri.test", "options.is_ping_enabled", "true"),
					resource.TestCheckResourceAttr("swo_uri.test", "options.is_tcp_enabled", "false"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "swo_uri.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccUriResourceConfig("test two"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("swo_uri.test", "name", "test two"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccUriResourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
	resource "swo_uri" "test_uri" {
		name        = %[1]q
		host  = "https://example.com"
	
		options = {
			is_ping_enabled = true
			is_tcp_enabled  = false
		}
	
		tcp_options = {
			port             = 80
			string_to_expect = "string to expect"
			string_to_send   = "string to send"
		}
	
		test_definitions = {
			test_from_location = "REGION"
	
			location_options = [
				{
					type  = "REGION"
					value = "NA"
				},
				{
					type  = "REGION"
					value = "AS"
				},
				{
					type  = "REGION"
					value = "SA"
				},
				{
					type  = "REGION"
					value = "OC"
				}
			]
	
			test_interval_in_seconds = 300
	
			platform_options = {
				test_from_all = false
				platforms     = ["AWS"]
			}
		}
	}`, name)
}
