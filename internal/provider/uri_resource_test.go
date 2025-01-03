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
				Config: testAccUriResourceConfig("test-acc test one [CREATE_TEST]"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("swo_uri.test", "id"),
					resource.TestCheckResourceAttr("swo_uri.test", "name", "test-acc test one [CREATE_TEST]"),
					resource.TestCheckResourceAttr("swo_uri.test", "host", "example.com"),
					resource.TestCheckResourceAttr("swo_uri.test", "options.is_ping_enabled", "false"),
					resource.TestCheckResourceAttr("swo_uri.test", "options.is_tcp_enabled", "true"),
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
				Config: testAccUriResourceConfig("test-acc test two [CREATE_TEST]"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("swo_uri.test", "name", "test-acc test two [CREATE_TEST]"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccUriResourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
	resource "swo_uri" "test" {
		name        = %[1]q
		host  = "example.com"
	
		options = {
			is_ping_enabled = false
			is_tcp_enabled  = true
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
