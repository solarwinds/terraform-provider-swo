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

					resource.TestCheckResourceAttr("swo_uri.test", "tcp_options.port", "80"),
					resource.TestCheckResourceAttr("swo_uri.test", "tcp_options.string_to_expect", "string to expect"),
					resource.TestCheckResourceAttr("swo_uri.test", "tcp_options.string_to_send", "string to send"),

					resource.TestCheckResourceAttr("swo_uri.test", "test_definitions.test_from_location", "REGION"),
					resource.TestCheckResourceAttr("swo_uri.test", "test_definitions.test_interval_in_seconds", "300"),
					resource.TestCheckResourceAttr("swo_uri.test", "test_definitions.platform_options.test_from_all", "false"),
					resource.TestCheckResourceAttr("swo_uri.test", "test_definitions.platform_options.platforms.#", "1"),
					resource.TestCheckResourceAttr("swo_uri.test", "test_definitions.platform_options.platforms.0", "AWS"),
					resource.TestCheckResourceAttr("swo_uri.test", "test_definitions.location_options.#", "1"),
					resource.TestCheckResourceAttr("swo_uri.test", "test_definitions.location_options.0.type", "REGION"),
					resource.TestCheckResourceAttr("swo_uri.test", "test_definitions.location_options.0.value", "NA"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "swo_uri.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			// This is temporarily disabled because the DEM API is failing with this due to a panic.
			// See NH-122218 for more details.
			//{
			//	Config: testAccUriResourceConfig("test-acc test two [UPDATE_TEST]"),
			//	Check: resource.ComposeAggregateTestCheckFunc(
			//		resource.TestCheckResourceAttr("swo_uri.test", "name", "test-acc test two [UPDATE_TEST]"),
			//	),
			//},
			// Delete testing automatically occurs in TestCase
		},
	})
}

// Only supported regions (location_options.value) in Dev and Stage are NA
// Production supports the following: NA, AS, SA, OC
func testAccUriResourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
	resource "swo_uri" "test" {
		name = %q
		host = "example.com"
	
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
