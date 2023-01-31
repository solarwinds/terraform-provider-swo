package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// func TestAccCoffeesDataSource(t *testing.T) {
// 	resource.Test(t, resource.TestCase{
// 		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
// 		Steps: []resource.TestStep{
// 			// Read testing
// 			{
// 				Config: providerConfig + `resource "swo_alert" "test" {}`,
// 				Check: resource.ComposeAggregateTestCheckFunc(
// 					// Verify number of coffees returned
// 					resource.TestCheckResourceAttr("data.hashicups_coffees.test", "coffees.#", "6"),
// 					// Verify the first coffee to ensure all attributes are set
// 					resource.TestCheckResourceAttr("data.hashicups_coffees.test", "coffees.0.description", ""),
// 					resource.TestCheckResourceAttr("data.hashicups_coffees.test", "coffees.0.id", "1"),
// 					resource.TestCheckResourceAttr("data.hashicups_coffees.test", "coffees.0.image", "/packer.png"),
// 					resource.TestCheckResourceAttr("data.hashicups_coffees.test", "coffees.0.ingredients.#", "3"),
// 					resource.TestCheckResourceAttr("data.hashicups_coffees.test", "coffees.0.ingredients.0.id", "1"),
// 					resource.TestCheckResourceAttr("data.hashicups_coffees.test", "coffees.0.ingredients.1.id", "2"),
// 					resource.TestCheckResourceAttr("data.hashicups_coffees.test", "coffees.0.ingredients.2.id", "4"),
// 					resource.TestCheckResourceAttr("data.hashicups_coffees.test", "coffees.0.name", "Packer Spiced Latte"),
// 					resource.TestCheckResourceAttr("data.hashicups_coffees.test", "coffees.0.price", "350"),
// 					resource.TestCheckResourceAttr("data.hashicups_coffees.test", "coffees.0.teaser", "Packed with goodness to spice up your images"),
// 					// Verify placeholder id attribute
// 					resource.TestCheckResourceAttr("data.hashicups_coffees.test", "id", "placeholder"),
// 				),
// 			},
// 		},
// 	})
// }

func TestAccAlertResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccAlertResourceConfig("Mock Alert Name"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("swo_alert.test", "id", "0bc4710d-e3b0-4590-9c9b-e5e46d81d912"),
					resource.TestCheckResourceAttr("swo_alert.test", "name", "Mock Alert Name"),
					resource.TestCheckResourceAttr("swo_alert.test", "description", "Mock alert description."),
					resource.TestCheckResourceAttr("swo_alert.test", "severity", "CRITICAL"),
					resource.TestCheckResourceAttr("swo_alert.test", "type", "ENTITY_METRIC"),
					resource.TestCheckResourceAttr("swo_alert.test", "target_entity_types.0", "Website"),
					// Verify number of conditions.
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.#", "1"),
					// Verify the conditions.
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.metric_name", "synthetics.https.response.time"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.threshold", ">=3000ms"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.duration", "5m"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.aggregation_type", "AVG"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.entity_ids.0", "e-1521946194448543744"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.entity_ids.1", "e-1521947552186691584"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.include_tags.0.name", "probe.city"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.include_tags.0.values.0", "Tokyo"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.include_tags.0.values.1", "Sao Paulo"),
					resource.TestCheckResourceAttr("swo_alert.test", "notifications.0", "123"),
					resource.TestCheckResourceAttr("swo_alert.test", "notifications.1", "456"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "swo_alert.test",
				ImportState:       true,
				ImportStateVerify: true,
				// This is not normally necessary, but is here because this
				// example code does not have an actual upstream service.
				// Once the Read method is able to refresh information from
				// the upstream service, this can be removed.
				ImportStateVerifyIgnore: []string{"id"},
			},
			// Update and Read testing
			{
				Config: testAccAlertResourceConfig("test_two"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("swo_alert.test", "name", "test_two"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccAlertResourceConfig(name string) string {
	return providerConfig + fmt.Sprintf(`

resource "swo_alert" "test" {
  name        = %[1]q
  description = "Mock alert description."
  severity    = "CRITICAL"
  type        = "ENTITY_METRIC"
	enabled     = true
  target_entity_types = ["Website"]
  conditions = [
    {
      metric_name      = "synthetics.https.response.time"
      threshold        = ">=3000ms"
      duration         = "5m"
      aggregation_type = "AVG"
      entity_ids = [
        "e-1521946194448543744",
        "e-1521947552186691584"
      ]
      include_tags = [
        {
          name = "probe.city"
          values : [
            "Tokyo",
            "Sao Paulo"
          ]
        }
      ],
      exclude_tags = []
    },
  ]
  notifications = ["123", "456"]
}
`, name)
}
