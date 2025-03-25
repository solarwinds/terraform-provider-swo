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
			// Validation Testing
			//{
			//	Config:      testAccWebsiteInvalidConfig("invalid"),
			//	ExpectError: regexp.MustCompile(".*At least one of.*"),
			//},
			// Create and Read testing
			//{
			//	Config: testAccWebsiteResourceConfig("test-acc test one [CREATE_TEST]"),
			//	Check: resource.ComposeAggregateTestCheckFunc(
			//		resource.TestCheckResourceAttrSet("swo_website.test", "id"),
			//		resource.TestCheckResourceAttr("swo_website.test", "name", "test-acc test one [CREATE_TEST]"),
			//		resource.TestCheckResourceAttr("swo_website.test", "url", "https://example.com"),
			//	),
			//},
			//{
			//	Config: testAccWebsiteResourceConfigWithoutCheckForString("test-acc create without string check [CREATE_TEST]"),
			//	Check: resource.ComposeAggregateTestCheckFunc(
			//		resource.TestCheckResourceAttrSet("swo_website.test", "id"),
			//		resource.TestCheckResourceAttr("swo_website.test", "name", "test-acc create without string check [CREATE_TEST]"),
			//		resource.TestCheckResourceAttr("swo_website.test", "url", "https://solarwinds.com"),
			//	),
			//},
			{
				Config: testAccWebsiteResourceConfigWithoutAvailability("test-acc create without availability [CREATE_TEST]"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("swo_website.test", "id"),
					resource.TestCheckResourceAttr("swo_website.test", "name", "test-acc create without availability [CREATE_TEST]"),
					resource.TestCheckResourceAttr("swo_website.test", "url", "https://solarwinds.com"),
				),
			},
			//{
			//	Config: testAccWebsiteResourceConfigWithoutRum("test-acc create without rum [CREATE_TEST]"),
			//	Check: resource.ComposeAggregateTestCheckFunc(
			//		resource.TestCheckResourceAttrSet("swo_website.test", "id"),
			//		resource.TestCheckResourceAttr("swo_website.test", "name", "test-acc create without rum [CREATE_TEST]"),
			//		resource.TestCheckResourceAttr("swo_website.test", "url", "https://solarwinds.com"),
			//	),
			//},
			// ImportState testing
			//{
			//	ResourceName:      "swo_website.test",
			//	ImportState:       true,
			//	ImportStateVerify: true,
			//},
			// Update and Read testing
			//{
			//	Config: testAccWebsiteResourceConfig("test-acc test two [UPDATE_TEST]"),
			//	Check: resource.ComposeAggregateTestCheckFunc(
			//		resource.TestCheckResourceAttr("swo_website.test", "name", "test-acc test two [UPDATE_TEST]"),
			//	),
			//},
			//{
			//	Config: testAccWebsiteResourceConfigWithoutCheckForString("test-acc test update without string check [UPDATE_TEST]"),
			//	Check: resource.ComposeAggregateTestCheckFunc(
			//		resource.TestCheckResourceAttr("swo_website.test", "name", "test-acc test update without string check [UPDATE_TEST]"),
			//	),
			//},
			//{
			//	Config: testAccWebsiteResourceConfigWithoutRum("test-acc test update without rum [UPDATE_TEST]"),
			//	Check: resource.ComposeAggregateTestCheckFunc(
			//		resource.TestCheckResourceAttr("swo_website.test", "name", "test-acc test update without rum [UPDATE_TEST]"),
			//	),
			//},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccWebsiteResourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
	resource "swo_website" "test" {
		name        = %[1]q
		url  = "https://example.com"
	
		monitoring = {
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

func testAccWebsiteResourceConfigWithoutCheckForString(name string) string {
	return providerConfig() + fmt.Sprintf(`
	resource "swo_website" "test" {
		name        = %[1]q
		url  = "https://solarwinds.com"
	
		monitoring = {

			availability = {
	
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

				ssl = {
					days_prior_to_expiration         = 30
					enabled                          = false
					ignore_intermediate_certificates = false
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


func testAccWebsiteResourceConfigWithoutAvailability(name string) string {
	return providerConfig() + fmt.Sprintf(`
	resource "swo_website" "test" {
		name        = %[1]q
		url  = "https://solarwinds.com"
	
		monitoring = {

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

func testAccWebsiteResourceConfigWithoutRum(name string) string {
	return providerConfig() + fmt.Sprintf(`
	resource "swo_website" "test" {
		name        = %[1]q
		url  = "https://solarwinds.com"
	
		monitoring = {

			availability = {
	
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

				ssl = {
					days_prior_to_expiration         = 30
					enabled                          = false
					ignore_intermediate_certificates = false
				}
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

func testAccWebsiteInvalidConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
	resource "swo_website" "test" {	
		name        = %[1]q
		url  = "https://example.com"
		monitoring = {
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
	}
	`, name)
}
