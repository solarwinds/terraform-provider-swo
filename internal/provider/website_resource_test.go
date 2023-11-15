package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccWebsiteResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		IsUnitTest:               true,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccWebsiteResourceConfig("test one"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("swo_website.test", "id"),
					resource.TestCheckResourceAttr("swo_website.test", "name", "test one"),
					resource.TestCheckResourceAttr("swo_website.test", "url", "www.solarwinds.com"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "swo_website.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccWebsiteResourceConfig("test two"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("swo_website.test", "name", "test two"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccWebsiteResourceConfig(name string) string {
	return providerConfig + fmt.Sprintf(`
	resource "swo_website" "test_website" {
		name        = %[1]q
		url  = "https://example.com"
	
		monitoring = {
			options = {
				is_availability_active = true
				is_rum_active          = true
			}
	
			availability = {
				check_for_string = {
					operator = "CONTAINS"
					value    = "example-string"
				}
	
				ssl = {
					days_prior_to_expiration         = 30
					enabled                          = true
					ignore_intermediate_certificates = true
				}
	
				protocols                = ["HTTP", "HTTPS"]
				test_interval_in_seconds = 300
				test_from_location       = "REGION"
	
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
	
				platform_options = {
					test_from_all = false
					platforms     = ["AWS"]
				}
			}
	
			rum = {
				apdex_time_in_seconds = 4
				spa                   = true
			}
	
			custom_headers = [
				{
					name  = "Custom-Header-1"
					value = "Custom-Value-1"
				},
				{
					name  = "Custom-Header-2"
					value = "Custom-Value-2"
				}
			]
		}
	}`, name)
}
