package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccCompositeMetricResource(t *testing.T) {
	// Tests basic CRUD.
	metricName := "composite.testacc.basic"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		IsUnitTest:               true,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccCompositeMetricResourceConfig(metricName, "display name one"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("swo_compositemetric.test", "name", metricName),
					resource.TestCheckResourceAttr("swo_compositemetric.test", "display_name", "display name one"),
					resource.TestCheckResourceAttr("swo_compositemetric.test", "description", "test-acc composite metric description"),
					resource.TestCheckResourceAttr("swo_compositemetric.test", "formula", "rate(system.disk.io[5m])"),
					resource.TestCheckResourceAttr("swo_compositemetric.test", "unit", "bytes/s"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "swo_compositemetric.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"description", "display_name"},
			},
			// Update and Read testing
			{
				Config: testAccCompositeUpdateMetricResourceConfig(metricName, "display name two"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("swo_compositemetric.test", "name", metricName),
					resource.TestCheckResourceAttr("swo_compositemetric.test", "display_name", "display name two"),
					resource.TestCheckResourceAttr("swo_compositemetric.test", "description", "Update metric description"),
					resource.TestCheckResourceAttr("swo_compositemetric.test", "formula", "SUM(synthetics.https.response.time)"),
					resource.TestCheckResourceAttr("swo_compositemetric.test", "unit", "m/s"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccCompositeMetricResourceOptionalFields(t *testing.T) {
	// Tests that the provider handles optional fields being nil/present gracefully
	metricName := "composite.testacc.optional.fields"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		IsUnitTest:               true,
		Steps: []resource.TestStep{
			// Create metric with minimal required fields only
			{
				Config: testAccCompositeMetricMinimalConfig(metricName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("swo_compositemetric.test", "name", metricName),
					resource.TestCheckResourceAttr("swo_compositemetric.test", "formula", "sum(synthetics.https.response.time)"),
					resource.TestCheckResourceAttrSet("swo_compositemetric.test", "id"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "swo_compositemetric.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"description", "display_name", "unit"},
			},
			// Add optional fields via update
			{
				Config: testAccCompositeMetricWithOptionalFieldsConfig(metricName, "Added Display Name", "Added Description", "ms"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("swo_compositemetric.test", "name", metricName),
					resource.TestCheckResourceAttr("swo_compositemetric.test", "formula", "sum(synthetics.https.response.time)"),
					resource.TestCheckResourceAttr("swo_compositemetric.test", "display_name", "Added Display Name"),
					resource.TestCheckResourceAttr("swo_compositemetric.test", "description", "Added Description"),
					resource.TestCheckResourceAttr("swo_compositemetric.test", "unit", "ms"),
				),
			},
			// Tests that the provider handles API responses with nil optional fields gracefully
			// (this could happen after UI edits)
			{
				RefreshState: true,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("swo_compositemetric.test", "name", metricName),
					resource.TestCheckResourceAttr("swo_compositemetric.test", "formula", "sum(synthetics.https.response.time)"),
					resource.TestCheckResourceAttrSet("swo_compositemetric.test", "display_name"),
					resource.TestCheckResourceAttrSet("swo_compositemetric.test", "description"),
					resource.TestCheckResourceAttrSet("swo_compositemetric.test", "unit"),
				),
			},
			// Verifies optional fields are removed
			{
				Config: testAccCompositeMetricMinimalConfig(metricName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("swo_compositemetric.test", "name", metricName),
					resource.TestCheckResourceAttr("swo_compositemetric.test", "formula", "sum(synthetics.https.response.time)"),
					resource.TestCheckResourceAttrSet("swo_compositemetric.test", "id"),
					// Verify optional fields are completely absent after removal
					resource.TestCheckNoResourceAttr("swo_compositemetric.test", "display_name"),
					resource.TestCheckNoResourceAttr("swo_compositemetric.test", "description"),
					resource.TestCheckNoResourceAttr("swo_compositemetric.test", "unit"),
				),
			},
		},
	})
}

func testAccCompositeMetricResourceConfig(name string, displayName string) string {
	return providerConfig() + fmt.Sprintf(`
	resource "swo_compositemetric" "test" {
		name        = %[1]q
		display_name = %[2]q
		description = "test-acc composite metric description"
		formula = "rate(system.disk.io[5m])"
		unit = "bytes/s"
	}`, name, displayName)
}

func testAccCompositeUpdateMetricResourceConfig(name string, displayName string) string {
	return providerConfig() + fmt.Sprintf(`
	resource "swo_compositemetric" "test" {
		name        = %[1]q
		display_name = %[2]q
		description = "Update metric description"
		formula = "SUM(synthetics.https.response.time)"
		unit = "m/s"
	}`, name, displayName)
}

func testAccCompositeMetricWithOptionalFieldsConfig(name string, displayName string, description string, unit string) string {
	return providerConfig() + fmt.Sprintf(`
	resource "swo_compositemetric" "test" {
		name        = %[1]q
		display_name = %[2]q
		description = %[3]q
		formula = "sum(synthetics.https.response.time)"
		unit = %[4]q
	}`, name, displayName, description, unit)
}

func testAccCompositeMetricMinimalConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
	resource "swo_compositemetric" "test" {
		name    = %[1]q
		formula = "sum(synthetics.https.response.time)"
		# Omitting optional fields: display_name, description, unit
	}`, name)
}
