package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

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
					resource.TestCheckResourceAttr("swo_alert.test", "trigger_reset_actions", "true"),
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
	trigger_reset_actions = true
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
