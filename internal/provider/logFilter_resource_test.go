package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccLogFilterResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		IsUnitTest:               true,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccLogFilterResourceConfig("test one"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("swo_logfilter.test", "id"),
					resource.TestCheckResourceAttr("swo_logfilter.test", "name", "test one"),
					resource.TestCheckResourceAttr("swo_logfilter.test", "description", "test description"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "swo_logfilter.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccLogFilterResourceConfig("test two"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("swo_logfilter.test", "name", "test two"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccLogFilterResourceConfig(name string) string {
	return providerConfig + fmt.Sprintf(`
	resource "swo_logfilter" "test_logfilter" {
		name = %[1]q
		description  = "test description"
		token_signature = null
		expressions = [
			{
				kind = "STRING"
				expression = "test expression"
			}
		]
	}`, name)
}
