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
				Config: testAccLogFilterResourceConfig("test-acc test one"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("swo_logfilter.test", "id"),
					resource.TestCheckResourceAttr("swo_logfilter.test", "name", "test-acc test one"),
					resource.TestCheckResourceAttr("swo_logfilter.test", "description", "test description"),
					resource.TestCheckResourceAttr("swo_logfilter.test", "token_signature", "U2aWJEYwSj-pvYegZdH1ozoq3kwapdngO1qDU3a8WXY"),
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
				Config: testAccLogFilterResourceConfig("test-acc test two"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("swo_logfilter.test", "name", "test-acc test two"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccLogFilterResourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
	resource "swo_logfilter" "test" {
		name = %[1]q
		description  = "test description"
		token_signature = "U2aWJEYwSj-pvYegZdH1ozoq3kwapdngO1qDU3a8WXY"
		expressions = [
			{
				kind = "STRING"
				expression = "test expression"
			}
		]
	}`, name)
}
