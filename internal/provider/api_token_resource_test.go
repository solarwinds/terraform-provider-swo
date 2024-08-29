package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccApiTokenResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		IsUnitTest:               true,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccApiTokenResourceConfig("test-acc test one"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("swo_apitoken.test", "id"),
					resource.TestCheckResourceAttr("swo_apitoken.test", "name", "test-acc test one"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "swo_apitoken.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccApiTokenResourceConfig("test-acc test two"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("swo_apitoken.test", "name", "test-acc test two"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccApiTokenResourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
	resource "swo_apitoken" "test_uri" {
		name        = %[1]q
		access_level = "READ"
		type = "public-api"
		enabled = true
		attributes = [
		  {
			key   = "attribute-key"
			value = "attribute value"
		  }
		]
	}`, name)
}
