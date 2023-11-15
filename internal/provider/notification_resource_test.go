package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccNotificationResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		IsUnitTest:               true,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccNotificationResourceConfig("test one"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("swo_notification.test", "id"),
					resource.TestCheckResourceAttr("swo_notification.test", "title", "test one"),
					resource.TestCheckResourceAttr("swo_notification.test", "description", "testing..."),
					resource.TestCheckResourceAttr("swo_notification.test", "type", "email"),
					resource.TestCheckResourceAttr("swo_notification.test", "settings.email.addresses.0.email", "test1@host.com"),
					resource.TestCheckResourceAttr("swo_notification.test", "settings.email.addresses.1.email", "test2@host.com"),
					resource.TestCheckResourceAttrSet("swo_notification.test", "created_by"),
					resource.TestCheckResourceAttrSet("swo_notification.test", "created_at"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "swo_notification.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccNotificationResourceConfig("test two"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("swo_notification.test", "title", "test two"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccNotificationResourceConfig(title string) string {
	return providerConfig + fmt.Sprintf(`
	resource "swo_notification" "test" {
		title        = %[1]q
		description = "testing..."
		type = "email"
		settings = {
			email = {
				addresses = [
					{
						email = "test1@host.com"
					},
					{
						email = "test2@host.com"
					},
				]
			}
		}
	}`, title)
}
