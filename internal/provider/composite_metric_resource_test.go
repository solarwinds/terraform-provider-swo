package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccCompositeMetricResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		IsUnitTest:               true,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccCompositeMetricResourceConfig("composite.testacc", "display name one"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("swo_compositemetric.test", "name", "composite.testacc"),
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
				Config: testAccCompositeUpdateMetricResourceConfig("composite.testacc", "display name two"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("swo_compositemetric.test", "name", "composite.testacc"),
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
