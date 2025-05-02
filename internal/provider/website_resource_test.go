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
			createTestStep(
				testAccWebsiteResourceConfig,
				"test-acc test one [CREATE_TEST]",
				"https://example.com",
				websiteMonitoringConfig,
				resource.TestCheckResourceAttr("swo_website.test", "monitoring.availability.check_for_string.operator", "CONTAINS"),
				resource.TestCheckResourceAttr("swo_website.test", "monitoring.availability.ssl.enabled", "true"),
				resource.TestCheckResourceAttr("swo_website.test", "monitoring.availability.protocols.0", "HTTP"),
				resource.TestCheckResourceAttr("swo_website.test", "monitoring.availability.protocols.1", "HTTPS"),
			),
			createTestStep(
				testAccWebsiteResourceConfig,
				"test-acc create without options [CREATE_TEST]",
				"https://solarwinds.com",
				websiteMonitoringConfigWithoutAvailabilityOptions,
				resource.TestCheckResourceAttr("swo_website.test", "monitoring.availability.protocols.0", "HTTP"),
				resource.TestCheckResourceAttr("swo_website.test", "monitoring.availability.protocols.1", "HTTPS"),
			),
			createTestStep(
				testAccWebsiteResourceConfig,
				"test-acc create without availability [CREATE_TEST]",
				"https://solarwinds.com",
				websiteMonitoringConfigWithoutAvailability,
				resource.TestCheckResourceAttr("swo_website.test", "monitoring.rum.apdex_time_in_seconds", "4"),
				resource.TestCheckResourceAttr("swo_website.test", "monitoring.rum.spa", "true"),
			),
			// ImportState testing
			{
				ResourceName:      "swo_website.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			createTestStep(
				testAccWebsiteResourceConfig,
				"test-acc test two [UPDATE_TEST]",
				"https://example.com",
				websiteMonitoringConfig,
				resource.TestCheckResourceAttr("swo_website.test", "monitoring.availability.check_for_string.operator", "CONTAINS"),
				resource.TestCheckResourceAttr("swo_website.test", "monitoring.availability.ssl.enabled", "true"),
				resource.TestCheckResourceAttr("swo_website.test", "monitoring.availability.protocols.0", "HTTP"),
				resource.TestCheckResourceAttr("swo_website.test", "monitoring.availability.protocols.1", "HTTPS"),
			),
			createTestStep(
				testAccWebsiteResourceConfig,
				"test-acc test update without options [UPDATE_TEST]",
				"https://solarwinds.com",
				websiteMonitoringConfigWithoutAvailabilityOptions,
				resource.TestCheckResourceAttr("swo_website.test", "monitoring.availability.protocols.0", "HTTP"),
				resource.TestCheckResourceAttr("swo_website.test", "monitoring.availability.protocols.1", "HTTPS"),
			),
			createTestStep(
				testAccWebsiteResourceConfig,
				"test-acc test update without availability [UPDATE_TEST]",
				"https://solarwinds.com",
				websiteMonitoringConfigWithoutAvailability,
				resource.TestCheckResourceAttr("swo_website.test", "monitoring.rum.apdex_time_in_seconds", "4"),
				resource.TestCheckResourceAttr("swo_website.test", "monitoring.rum.spa", "true"),
			),
			// Delete testing automatically occurs in TestCase
		},
	})
}

var (
	websiteMonitoringConfig                           = monitoringConfig(availabilityConfig(true, true), "null", true)
	websiteMonitoringConfigWithoutAvailability        = monitoringConfig("null", rumConfig(), false)
	websiteMonitoringConfigWithoutAvailabilityOptions = monitoringConfig(availabilityConfig(false, false), rumConfig(), true)
)

func createTestStep(configFunc func(string, string, string) string, name, url string, monitoring string, additionalChecks ...resource.TestCheckFunc) resource.TestStep {
	return resource.TestStep{
		Config: configFunc(name, url, monitoring),
		Check: resource.ComposeAggregateTestCheckFunc(
			append([]resource.TestCheckFunc{
				resource.TestCheckResourceAttrSet("swo_website.test", "id"),
				resource.TestCheckResourceAttr("swo_website.test", "name", name),
				resource.TestCheckResourceAttr("swo_website.test", "url", url),
			}, additionalChecks...)...,
		),
	}
}

func testAccWebsiteResourceConfig(name string, url string, monitoring string) string {
	return fmt.Sprintf(`
    %s
	resource "swo_website" "test" {
    name = %[2]q
    url  = %[3]q
	monitoring = %s
	}`, providerConfig(), name, url, monitoring)
}

func monitoringConfig(availability, rum string, useDeprecatedCustomHeaders bool) string {
	monitoringConf := fmt.Sprintf(`{
		availability = %s
		rum = %s
	`, availability, rum)

	if useDeprecatedCustomHeaders {
		monitoringConf += `
		custom_headers = [
			{
				name  = "Custom-Header-1-Deprecated"
				value = "Custom-Value-1-Deprecated"
			}
		]`
	}

	monitoringConf += `}`

	return monitoringConf
}

func rumConfig() string {
	return `{
		apdex_time_in_seconds = 4
		spa                   = true
    }`
}

func availabilityConfig(includeCheckForString bool, includeSSL bool) string {
	availabilityConfig := `{`

	if includeCheckForString {
		availabilityConfig += `
			check_for_string = {
				operator = "CONTAINS"
				value    = "example-string"
			}`
	}

	if includeSSL {
		availabilityConfig += `
			ssl = {
				days_prior_to_expiration         = 30
				enabled                          = true
				ignore_intermediate_certificates = true
			}`
	}

	availabilityConfig += `
			protocols                = ["HTTP", "HTTPS"]
			test_interval_in_seconds = 300
			test_from_location       = "REGION"
	
			location_options = [
				{
					type  = "REGION"
					value = "NA"
				}
			]
	
			platform_options = {
				test_from_all = false
				platforms     = ["AWS"]
			}
		}`
	return availabilityConfig
}
