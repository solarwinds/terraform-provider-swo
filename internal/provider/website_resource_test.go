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
				"test-acc test two [CREATE_TEST]",
				"https://example.com",
				true,
				websiteMonitoringConfig,

				resource.TestCheckResourceAttr("swo_website.test", "monitoring.availability.check_for_string.operator", "CONTAINS"),
				resource.TestCheckResourceAttr("swo_website.test", "monitoring.availability.check_for_string.value", "example-string"),

				resource.TestCheckResourceAttr("swo_website.test", "monitoring.availability.ssl.enabled", "true"),
				resource.TestCheckResourceAttr("swo_website.test", "monitoring.availability.ssl.days_prior_to_expiration", "30"),
				resource.TestCheckResourceAttr("swo_website.test", "monitoring.availability.ssl.ignore_intermediate_certificates", "true"),

				resource.TestCheckResourceAttr("swo_website.test", "monitoring.availability.protocols.#", "2"),
				resource.TestCheckResourceAttr("swo_website.test", "monitoring.availability.protocols.0", "HTTP"),
				resource.TestCheckResourceAttr("swo_website.test", "monitoring.availability.protocols.1", "HTTPS"),

				resource.TestCheckResourceAttr("swo_website.test", "monitoring.availability.test_interval_in_seconds", "300"),
				resource.TestCheckResourceAttr("swo_website.test", "monitoring.availability.test_from_location", "REGION"),

				resource.TestCheckResourceAttr("swo_website.test", "monitoring.availability.location_options.#", "1"),
				resource.TestCheckResourceAttr("swo_website.test", "monitoring.availability.location_options.0.type", "REGION"),
				resource.TestCheckResourceAttr("swo_website.test", "monitoring.availability.location_options.0.value", "NA"),

				resource.TestCheckResourceAttr("swo_website.test", "monitoring.availability.platform_options.test_from_all", "false"),
				resource.TestCheckResourceAttr("swo_website.test", "monitoring.availability.platform_options.platforms.#", "1"),
				resource.TestCheckResourceAttr("swo_website.test", "monitoring.availability.platform_options.platforms.0", "AWS"),

				resource.TestCheckResourceAttr("swo_website.test", "monitoring.custom_headers.#", "1"),
				resource.TestCheckResourceAttr("swo_website.test", "monitoring.custom_headers.0.name", "Custom-Header-1-Deprecated"),
				resource.TestCheckResourceAttr("swo_website.test", "monitoring.custom_headers.0.value", "Custom-Value-1-Deprecated"),

				resource.TestCheckResourceAttr("swo_website.test", "monitoring.availability.outage_configuration.failing_test_locations", "any"),
				resource.TestCheckResourceAttr("swo_website.test", "monitoring.availability.outage_configuration.consecutive_for_down", "5"),

				resource.TestCheckNoResourceAttr("swo_website.test", "monitoring.rum"),
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
				true,
				websiteMonitoringConfig,
				resource.TestCheckResourceAttr("swo_website.test", "monitoring.availability.check_for_string.operator", "CONTAINS"),
				resource.TestCheckResourceAttr("swo_website.test", "monitoring.availability.ssl.enabled", "true"),
				resource.TestCheckResourceAttr("swo_website.test", "monitoring.availability.protocols.0", "HTTP"),
				resource.TestCheckResourceAttr("swo_website.test", "monitoring.availability.protocols.1", "HTTPS"),
			),
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccWebsiteResourceWithoutAvailabilityOptionResources(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		IsUnitTest:               true,
		Steps: []resource.TestStep{
			// Create and Read testing
			createTestStep(
				testAccWebsiteResourceConfig,
				"test-acc create without options [CREATE_TEST]",
				"https://solarwinds.com",
				false,
				websiteMonitoringConfigWithoutAvailabilityOptions,
				resource.TestCheckResourceAttr("swo_website.test", "monitoring.availability.protocols.0", "HTTP"),
				resource.TestCheckResourceAttr("swo_website.test", "monitoring.availability.protocols.1", "HTTPS"),

				resource.TestCheckNoResourceAttr("swo_website.test", "monitoring.availability.check_for_string"),
				resource.TestCheckNoResourceAttr("swo_website.test", "monitoring.availability.ssl"),
				resource.TestCheckNoResourceAttr("swo_website.test", "monitoring.availability.outage_configuration"),
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
				"test-acc update without options [UPDATE_TEST]",
				"https://solarwinds.com",
				false,
				websiteMonitoringConfigWithoutAvailabilityOptions,
				resource.TestCheckResourceAttr("swo_website.test", "monitoring.availability.protocols.0", "HTTP"),
				resource.TestCheckResourceAttr("swo_website.test", "monitoring.availability.protocols.1", "HTTPS"),
			),
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccWebsiteResourceWithoutAvailabilityResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		IsUnitTest:               true,
		Steps: []resource.TestStep{
			// Create and Read testing
			createTestStep(
				testAccWebsiteResourceConfig,
				"test-acc create without availability [CREATE_TEST]",
				"https://solarwinds.com",
				false,
				websiteMonitoringConfigWithoutAvailability,
				resource.TestCheckResourceAttr("swo_website.test", "monitoring.rum.apdex_time_in_seconds", "4"),
				resource.TestCheckResourceAttr("swo_website.test", "monitoring.rum.spa", "true"),

				resource.TestCheckNoResourceAttr("swo_website.test", "monitoring.availability"),
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
				"test-acc update without availability [UPDATE_TEST]",
				"https://solarwinds.com",
				false,
				websiteMonitoringConfigWithoutAvailability,
				resource.TestCheckResourceAttr("swo_website.test", "monitoring.rum.apdex_time_in_seconds", "4"),
				resource.TestCheckResourceAttr("swo_website.test", "monitoring.rum.spa", "true"),
			),
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccWebsiteResourceMonitoringOptionsComputed(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		IsUnitTest:               true,
		Steps: []resource.TestStep{
			// Create with both monitoring types and verify computed options
			createTestStep(
				testAccWebsiteResourceConfig,
				"test-acc monitoring options computed [CREATE_TEST]",
				"https://example.com",
				false,
				websiteMonitoringConfigBothTypes,
				// Verify both monitoring types are active
				resource.TestCheckResourceAttr("swo_website.test", "monitoring.options.is_availability_active", "true"),
				resource.TestCheckResourceAttr("swo_website.test", "monitoring.options.is_rum_active", "true"),
				resource.TestCheckResourceAttr("swo_website.test", "monitoring.availability.check_for_string.operator", "DOES_NOT_CONTAIN"),
				resource.TestCheckResourceAttr("swo_website.test", "monitoring.availability.check_for_string.value", "error"),
				resource.TestCheckResourceAttr("swo_website.test", "monitoring.availability.protocols.0", "HTTPS"),
				resource.TestCheckResourceAttr("swo_website.test", "monitoring.availability.test_from_location", "REGION"),
				resource.TestCheckResourceAttr("swo_website.test", "monitoring.availability.location_options.#", "1"),
				resource.TestCheckResourceAttr("swo_website.test", "monitoring.availability.location_options.0.type", "REGION"),
				resource.TestCheckResourceAttr("swo_website.test", "monitoring.availability.location_options.0.value", "NA"),
				resource.TestCheckResourceAttr("swo_website.test", "monitoring.availability.platform_options.platforms.#", "2"),
				resource.TestCheckResourceAttr("swo_website.test", "monitoring.availability.platform_options.platforms.0", "AWS"),
				resource.TestCheckResourceAttr("swo_website.test", "monitoring.availability.platform_options.platforms.1", "AZURE"),
				resource.TestCheckResourceAttr("swo_website.test", "monitoring.rum.apdex_time_in_seconds", "7"),
				resource.TestCheckResourceAttr("swo_website.test", "monitoring.rum.spa", "false"),
			),
			// ImportState testing
			{
				ResourceName:      "swo_website.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update to remove RUM and verify options reflect the change
			createTestStep(
				testAccWebsiteResourceConfig,
				"test-acc monitoring options updated [UPDATE_TEST]",
				"https://example.com",
				false,
				websiteMonitoringConfigAvailabilityOnly,
				// Verify availability is still active, RUM is now inactive
				resource.TestCheckResourceAttr("swo_website.test", "monitoring.options.is_availability_active", "true"),
				resource.TestCheckResourceAttr("swo_website.test", "monitoring.options.is_rum_active", "false"),
				// Verify availability still works
				resource.TestCheckResourceAttr("swo_website.test", "monitoring.availability.protocols.0", "HTTP"),
				resource.TestCheckResourceAttr("swo_website.test", "monitoring.availability.test_interval_in_seconds", "600"),
				// Verify RUM is removed
				resource.TestCheckNoResourceAttr("swo_website.test", "monitoring.rum"),
			),
			// Update back to both types
			createTestStep(
				testAccWebsiteResourceConfig,
				"test-acc monitoring options full cycle [UPDATE_TEST]",
				"https://example.com",
				false,
				websiteMonitoringConfigBothTypes,
				// Verify both are active again
				resource.TestCheckResourceAttr("swo_website.test", "monitoring.options.is_availability_active", "true"),
				resource.TestCheckResourceAttr("swo_website.test", "monitoring.options.is_rum_active", "true"),
				// Verify RUM snippet is set again
				resource.TestCheckResourceAttrSet("swo_website.test", "monitoring.rum.snippet"),
			),
		},
	})
}

var (
	websiteMonitoringConfig                           = monitoringConfig(availabilityConfig(true, true, true), "null", true)
	websiteMonitoringConfigWithoutAvailability        = monitoringConfig("null", rumConfig(), false)
	websiteMonitoringConfigWithoutAvailabilityOptions = monitoringConfig(availabilityConfig(false, false, false), rumConfig(), true)
	websiteMonitoringConfigBothTypes                  = monitoringConfigWithCustomHeaders(availabilityConfigForMonitoringOptionsTest(), rumConfigForMonitoringOptionsTest(), "X-Test-Header", "test-value")
	websiteMonitoringConfigAvailabilityOnly           = monitoringConfigWithCustomHeaders(availabilityConfigSimpleForMonitoringOptionsTest(), "null", "X-Test-Header", "test-value")
)

func createTestStep(configFunc func(string, string, string, bool) string, name, url string, withTags bool, monitoring string, additionalChecks ...resource.TestCheckFunc) resource.TestStep {

	resourceChecks := []resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet("swo_website.test", "id"),
		resource.TestCheckResourceAttr("swo_website.test", "name", name),
		resource.TestCheckResourceAttr("swo_website.test", "url", url),
	}

	if withTags {
		resourceChecks = append(resourceChecks,
			resource.TestCheckResourceAttr("swo_website.test", "tags.#", "2"),
			//tag object order can be changed. Check for total number and nothing else.
		)
	} else {
		resourceChecks = append(resourceChecks,
			resource.TestCheckNoResourceAttr("swo_website.test", "tags"),
		)
	}

	return resource.TestStep{
		Config: configFunc(name, url, monitoring, withTags),
		Check: resource.ComposeAggregateTestCheckFunc(
			append(resourceChecks, additionalChecks...)...,
		),
	}
}

func testAccWebsiteResourceConfig(name string, url string, monitoring string, includeTags bool) string {

	resourceConfig := fmt.Sprintf(`
    %s
	resource "swo_website" "test" {
    name = %[2]q
    url  = %[3]q `, providerConfig(), name, url)

	if includeTags {
		resourceConfig += `
    		tags = [
				{
					key = "one-key"
					value = "one-value"
				},
				{
					key = "two-key"
					value = "two-value"
				}
			]`
	}

	monitoringStr := fmt.Sprintf(`
		monitoring = %s 
	}`, monitoring)
	resourceConfig += monitoringStr

	return resourceConfig
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

func monitoringConfigWithCustomHeaders(availability, rum, headerName, headerValue string) string {
	monitoringConf := fmt.Sprintf(`{
		availability = %s
		rum = %s
		custom_headers = [
			{
				name  = "%s"
				value = "%s"
			}
		]
	}`, availability, rum, headerName, headerValue)

	return monitoringConf
}

func rumConfig() string {
	return rumConfigWithOptions(4, true)
}

func rumConfigWithOptions(apdexTime int64, spa bool) string {
	return fmt.Sprintf(`{
		apdex_time_in_seconds = %d
		spa                   = %t
    }`, apdexTime, spa)
}

func rumConfigForMonitoringOptionsTest() string {
	return `{
		apdex_time_in_seconds = 7
		spa                   = false
    }`
}

func availabilityConfig(includeCheckForString bool, includeSSL bool, includeOutage bool) string {
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

	if includeOutage {
		availabilityConfig += `
			outage_configuration = {
				failing_test_locations = "any"
				consecutive_for_down   = 5
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

func availabilityConfigForMonitoringOptionsTest() string {
	return `{
		check_for_string = {
			operator = "DOES_NOT_CONTAIN"
			value    = "error"
		}
		protocols                = ["HTTPS"]
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
			platforms     = ["AWS", "AZURE"]
		}
	}`
}

func availabilityConfigSimpleForMonitoringOptionsTest() string {
	return `{
		protocols                = ["HTTP"]
		test_interval_in_seconds = 600
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
}
