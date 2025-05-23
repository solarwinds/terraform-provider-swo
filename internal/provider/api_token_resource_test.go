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
					resource.TestCheckResourceAttr("swo_apitoken.test", "access_level", "API_FULL"),
					resource.TestCheckResourceAttr("swo_apitoken.test", "type", "public-api"),
					resource.TestCheckResourceAttr("swo_apitoken.test", "enabled", "true"),
					resource.TestCheckResourceAttr("swo_apitoken.test", "attributes.0.key", "attribute-key"),
					resource.TestCheckResourceAttr("swo_apitoken.test", "attributes.0.value", "attribute value"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "swo_apitoken.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"token"},
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
	resource "swo_apitoken" "test" {
		name        = %[1]q
		access_level = "API_FULL"
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
