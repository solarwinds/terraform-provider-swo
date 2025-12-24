package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccEntityAlertResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		IsUnitTest:               true,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccEntityAlertResourceConfig("test-acc Mock Entity Alert Name"),
				Check: resource.ComposeAggregateTestCheckFunc(
					//resource.TestCheckResourceAttr("swo_alert.test", "id", "0bc4710d-e3b0-4590-9c9b-e5e46d81d912"),
					resource.TestCheckResourceAttr("swo_alert.test", "name", "test-acc Mock Entity Alert Name"),
					resource.TestCheckResourceAttr("swo_alert.test", "description", "Mock alert description."),
					resource.TestCheckResourceAttr("swo_alert.test", "severity", "CRITICAL"),
					resource.TestCheckResourceAttr("swo_alert.test", "trigger_reset_actions", "false"),
					// Verify actions
					resource.TestCheckResourceAttr("swo_alert.test", "notification_actions.0.configuration_ids.0", "333:email"),
					resource.TestCheckResourceAttr("swo_alert.test", "notification_actions.0.configuration_ids.1", "444:msteams"),
					resource.TestCheckResourceAttr("swo_alert.test", "notification_actions.0.resend_interval_seconds", "600"),
					// Verify the number of conditions.
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.#", "1"),
					// Verify the conditions.
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.target_entity_types.0", "Website"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.metric_name", "synthetics.https.response.time"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.threshold", ">=12.3"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.not_reporting", "false"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.duration", "5m"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.aggregation_type", "AVG"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.entity_ids.0", "e-1521946194448543744"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.entity_ids.1", "e-1521947552186691584"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.query_search", "healthScore.categoryV2:bad"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.group_by_metric_tag.0", "host.name"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.include_tags.0.name", "probe.city"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.include_tags.0.values.0", "Tokyo"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.include_tags.0.values.1", "Sao Paulo"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.exclude_tags.0.name", "service.name"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.exclude_tags.0.values.0", "test-service"),

					resource.TestCheckResourceAttr("swo_alert.test", "notifications.0", "123"),
					resource.TestCheckResourceAttr("swo_alert.test", "notifications.1", "456"),
					resource.TestCheckResourceAttr("swo_alert.test", "runbook_link", "https://www.runbooklink.com"),
					resource.TestCheckResourceAttr("swo_alert.test", "trigger_delay_seconds", "300"),
				),
			},
			// ImportState testing
			/*{
				ResourceName:      "swo_alert.test",
				ImportState:       true,
				ImportStateVerify: true,
			}*/
			// Update and Read testing
			{
				Config: testAccEntityAlertResourceConfig("test-acc test_two"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("swo_alert.test", "name", "test-acc test_two"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccMetricGroupAlertResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		IsUnitTest:               true,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccMetricGroupAlertResourceConfig("test-acc Mock Metric Group Alert"),
				Check: resource.ComposeAggregateTestCheckFunc(
					//resource.TestCheckResourceAttr("swo_alert.test", "id", "0bc4710d-e3b0-4590-9c9b-e5e46d81d912"),
					resource.TestCheckResourceAttr("swo_alert.test", "name", "test-acc Mock Metric Group Alert"),
					resource.TestCheckResourceAttr("swo_alert.test", "description", "Mock alert description."),
					resource.TestCheckResourceAttr("swo_alert.test", "severity", "CRITICAL"),
					resource.TestCheckResourceAttr("swo_alert.test", "trigger_reset_actions", "false"),
					// Verify actions
					resource.TestCheckResourceAttr("swo_alert.test", "notification_actions.0.configuration_ids.0", "333:email"),
					resource.TestCheckResourceAttr("swo_alert.test", "notification_actions.0.configuration_ids.1", "444:msteams"),
					resource.TestCheckResourceAttr("swo_alert.test", "notification_actions.0.resend_interval_seconds", "600"),
					// Verify the number of conditions.
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.#", "1"),
					// Verify the conditions.
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.metric_name", "synthetics.https.response.time"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.threshold", ">=10.5"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.not_reporting", "false"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.duration", "5m"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.aggregation_type", "AVG"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.target_entity_types.0", "Uri"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.query_search", "healthScore.categoryV2:bad"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.group_by_metric_tag.0", "host.name"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.include_tags.0.name", "probe.city"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.include_tags.0.values.0", "Tokyo"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.include_tags.0.values.1", "Sao Paulo"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.exclude_tags.0.name", "service.name"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.exclude_tags.0.values.0", "test-service"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.exclude_tags.0.operation", "CONTAINS"),

					resource.TestCheckResourceAttr("swo_alert.test", "runbook_link", "https://www.runbooklink.com"),
					resource.TestCheckResourceAttr("swo_alert.test", "trigger_delay_seconds", "300"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "swo_alert.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"notification_actions"},
			},
			// Update and Read testing
			{
				Config: testAccMetricGroupAlertResourceConfig("test-acc Mock Metric Group Alert 2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("swo_alert.test", "name", "test-acc Mock Metric Group Alert 2"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccAlertResourceNotReporting(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		IsUnitTest:               true,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccAlertResourceNotReportingConfig("test-acc Mock Alert Not Reporting Name"),
				Check: resource.ComposeAggregateTestCheckFunc(
					//resource.TestCheckResourceAttr("swo_alert.test", "id", "0bc4710d-e3b0-4590-9c9b-e5e46d81d912"),
					resource.TestCheckResourceAttr("swo_alert.test", "name", "test-acc Mock Alert Not Reporting Name"),
					resource.TestCheckResourceAttr("swo_alert.test", "description", "Mock alert description."),
					resource.TestCheckResourceAttr("swo_alert.test", "severity", "CRITICAL"),
					resource.TestCheckResourceAttr("swo_alert.test", "trigger_reset_actions", "false"),
					// Verify actions
					resource.TestCheckResourceAttr("swo_alert.test", "notification_actions.0.configuration_ids.0", "333:email"),
					resource.TestCheckResourceAttr("swo_alert.test", "notification_actions.0.configuration_ids.1", "444:msteams"),
					resource.TestCheckResourceAttr("swo_alert.test", "notification_actions.0.resend_interval_seconds", "600"),
					// Verify the number of conditions.
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.#", "1"),
					// Verify the conditions.
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.target_entity_types.0", "Website"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.metric_name", "synthetics.https.response.time"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.not_reporting", "true"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.duration", "10m"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.aggregation_type", "COUNT"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.entity_ids.0", "e-1521946194448543744"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.entity_ids.1", "e-1521947552186691584"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.include_tags.0.name", "probe.city"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.include_tags.0.values.0", "Tokyo"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.include_tags.0.values.1", "Sao Paulo"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.exclude_tags.0.name", "service.name"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.exclude_tags.0.values.0", "test-service"),

					resource.TestCheckResourceAttr("swo_alert.test", "notifications.0", "123"),
					resource.TestCheckResourceAttr("swo_alert.test", "notifications.1", "456"),
					resource.TestCheckResourceAttr("swo_alert.test", "runbook_link", "https://www.runbooklink.com"),
					resource.TestCheckResourceAttr("swo_alert.test", "trigger_delay_seconds", "0"),
				),
			},
			// ImportState testing
			/*{
				ResourceName:      "swo_alert.test",
				ImportState:       true,
				ImportStateVerify: true,
			}*/
			// Update and Read testing
			{
				Config: testAccAlertResourceNotReportingConfig("test-acc test_two_not_reporting"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("swo_alert.test", "name", "test-acc test_two_not_reporting"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAndMetricConditionAlertResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		IsUnitTest:               true,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAndConditionAlertResourceConfig("test-acc Mock AND Multi Condition Alert Name"),
				Check: resource.ComposeAggregateTestCheckFunc(

					resource.TestCheckResourceAttr("swo_alert.test", "name", "test-acc Mock AND Multi Condition Alert Name"),
					resource.TestCheckResourceAttr("swo_alert.test", "description", ""),
					resource.TestCheckResourceAttr("swo_alert.test", "severity", "INFO"),
					resource.TestCheckResourceAttr("swo_alert.test", "enabled", "false"),
					// Verify actions
					resource.TestCheckResourceAttr("swo_alert.test", "notification_actions.0.configuration_ids.0", "333:email"),
					resource.TestCheckResourceAttr("swo_alert.test", "notification_actions.0.configuration_ids.1", "444:msteams"),
					resource.TestCheckResourceAttr("swo_alert.test", "notification_actions.0.resend_interval_seconds", "600"),
					// Verify the number of conditions.
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.#", "3"),
					// Verify the conditions.
					resource.TestCheckResourceAttr("swo_alert.test", "conditions_operation", "AND"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.target_entity_types.0", "Website"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.metric_name", "sw.metrics.healthscore"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.threshold", "<10"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.not_reporting", "false"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.duration", "5m"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.aggregation_type", "AVG"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.entity_ids.0", "e-1521946194448543744"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.entity_ids.1", "e-1521947552186691584"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.query_search", "healthScore.categoryV2:bad"),

					resource.TestCheckResourceAttr("swo_alert.test", "conditions.1.target_entity_types.0", "Website"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.1.metric_name", "synthetics.https.response.time"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.1.threshold", ">=3000"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.1.not_reporting", "false"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.1.duration", "5m"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.1.aggregation_type", "AVG"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.1.entity_ids.0", "e-1521946194448543744"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.1.entity_ids.1", "e-1521947552186691584"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.1.query_search", "healthScore.categoryV2:bad"),

					resource.TestCheckResourceAttr("swo_alert.test", "conditions.2.target_entity_types.0", "Website"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.2.metric_name", "synthetics.status"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.2.not_reporting", "true"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.2.duration", "30m"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.2.aggregation_type", "COUNT"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.2.entity_ids.0", "e-1521946194448543744"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.2.entity_ids.1", "e-1521947552186691584"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.2.query_search", "healthScore.categoryV2:bad"),

					resource.TestCheckResourceAttr("swo_alert.test", "notifications.0", "123"),
					resource.TestCheckResourceAttr("swo_alert.test", "notifications.1", "456"),
					resource.TestCheckResourceAttr("swo_alert.test", "runbook_link", "https://www.runbooklink.com"),
					resource.TestCheckResourceAttr("swo_alert.test", "trigger_delay_seconds", "600"),
				),
			},
			// Update and Read testing
			{
				Config: testAndConditionAlertResourceConfig("test-acc AND test_two"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("swo_alert.test", "name", "test-acc AND test_two"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestOrMetricConditionAlertResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		IsUnitTest:               true,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testOrConditionAlertResourceConfig("test-acc Mock OR Multi Condition Alert Name"),
				Check: resource.ComposeAggregateTestCheckFunc(

					resource.TestCheckResourceAttr("swo_alert.test", "name", "test-acc Mock OR Multi Condition Alert Name"),
					resource.TestCheckResourceAttr("swo_alert.test", "description", ""),
					resource.TestCheckResourceAttr("swo_alert.test", "severity", "INFO"),
					resource.TestCheckResourceAttr("swo_alert.test", "enabled", "false"),
					// Verify actions
					resource.TestCheckResourceAttr("swo_alert.test", "notification_actions.0.configuration_ids.0", "333:email"),
					resource.TestCheckResourceAttr("swo_alert.test", "notification_actions.0.configuration_ids.1", "444:msteams"),
					resource.TestCheckResourceAttr("swo_alert.test", "notification_actions.0.resend_interval_seconds", "600"),
					// Verify the number of conditions.
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.#", "3"),
					// Verify the conditions.
					resource.TestCheckResourceAttr("swo_alert.test", "conditions_operation", "OR"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.target_entity_types.0", "Website"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.metric_name", "sw.metrics.healthscore"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.threshold", "<10"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.not_reporting", "false"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.duration", "5m"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.aggregation_type", "AVG"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.entity_ids.0", "e-1521946194448543744"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.entity_ids.1", "e-1521947552186691584"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.query_search", "healthScore.categoryV2:bad"),

					resource.TestCheckResourceAttr("swo_alert.test", "conditions.1.target_entity_types.0", "Website"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.1.metric_name", "synthetics.https.response.time"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.1.threshold", ">=3000"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.1.not_reporting", "false"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.1.duration", "5m"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.1.aggregation_type", "AVG"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.1.entity_ids.0", "e-1521946194448543744"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.1.entity_ids.1", "e-1521947552186691584"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.1.query_search", "healthScore.categoryV2:bad"),

					resource.TestCheckResourceAttr("swo_alert.test", "conditions.2.target_entity_types.0", "Website"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.2.metric_name", "synthetics.status"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.2.not_reporting", "true"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.2.duration", "30m"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.2.aggregation_type", "COUNT"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.2.entity_ids.0", "e-1521946194448543744"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.2.entity_ids.1", "e-1521947552186691584"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.2.query_search", "healthScore.categoryV2:bad"),

					resource.TestCheckResourceAttr("swo_alert.test", "notifications.0", "123"),
					resource.TestCheckResourceAttr("swo_alert.test", "notifications.1", "456"),
					resource.TestCheckResourceAttr("swo_alert.test", "runbook_link", "https://www.runbooklink.com"),
					resource.TestCheckResourceAttr("swo_alert.test", "trigger_delay_seconds", "600"),
				),
			},
			// Update and Read testing
			{
				Config: testAndConditionAlertResourceConfig("test-acc OR test_two"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("swo_alert.test", "name", "test-acc OR test_two"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccMultiAttributeConditionAlertResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		IsUnitTest:               true,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccMultiAttributeAlertResourceConfig("test-acc Mock Entity Alert Name"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("swo_alert.test", "name", "test-acc Mock Entity Alert Name"),
					resource.TestCheckResourceAttr("swo_alert.test", "description", "Mock alert description."),
					resource.TestCheckResourceAttr("swo_alert.test", "severity", "INFO"),
					// Verify the number of conditions.
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.#", "2"),
					// Verify the conditions.
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.attribute_name", "healthScore.scoreV2"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.attribute_values.0", "0"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.attribute_values.1", "100"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.attribute_operator", "IN"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.target_entity_types.0", "Website"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.1.attribute_name", "inMaintenance"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.1.attribute_value", "true"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.1.attribute_operator", "="),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.1.target_entity_types.0", "Website"),
				),
			},
			// Update and Read testing
			{
				Config: testAccMultiAttributeAlertResourceConfig("test-acc test_two"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("swo_alert.test", "name", "test-acc test_two"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccEntityAlertResourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`

resource "swo_alert" "test" {
 name        = %[1]q
 description = "Mock alert description."
 severity    = "CRITICAL"
 enabled     = true
 notification_actions = [
   {
	  configuration_ids = ["333:email", "444:msteams"]
	  resend_interval_seconds = 600
   },
 ]
 conditions = [
	{
	  metric_name      = "synthetics.https.response.time"
	  threshold        = ">=12.3"
	  duration         = "5m"
	  not_reporting    = false
	  aggregation_type = "AVG"
	  target_entity_types = ["Website"]
	  entity_ids = [
		"e-1521946194448543744",
		"e-1521947552186691584"
	  ]
      query_search = "healthScore.categoryV2:bad"
	  group_by_metric_tag = [
		"host.name"
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
	  exclude_tags = [{
		  name = "service.name"
		  values : [
			"test-service"
		  ]
		}]
	},
 ]
 notifications = ["123", "456"]
 runbook_link = "https://www.runbooklink.com"
 trigger_delay_seconds = 300
}
`, name)
}

func testAndConditionAlertResourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`

resource "swo_alert" "test" {
 name        = %[1]q
 description = ""
 severity    = "INFO"
 enabled     = false
 notification_actions = [
   {
	  configuration_ids = ["333:email", "444:msteams"]
	  resend_interval_seconds = 600
   },
 ]
 conditions = [
	{
	  metric_name      = "synthetics.https.response.time"
	  threshold        = ">=3000"
	  duration         = "5m"
	  not_reporting    = false
	  aggregation_type = "AVG"
	  target_entity_types = ["Website"]
	  entity_ids = [
		"e-1521946194448543744",
		"e-1521947552186691584"
	  ]
      query_search = "healthScore.categoryV2:bad"
	},
	{
	  metric_name      = "sw.metrics.healthscore"
	  threshold        = "<10"
	  duration         = "5m"
	  not_reporting    = false
	  aggregation_type = "AVG"
	  target_entity_types = ["Website"]
	  entity_ids = [
		"e-1521946194448543744",
		"e-1521947552186691584"
	  ]
      query_search = "healthScore.categoryV2:bad"
	},
	{
	  metric_name      = "synthetics.status"
	  not_reporting    = true
	  duration         = "30m"
	  aggregation_type = "COUNT"
	  target_entity_types = ["Website"]
	  entity_ids = [
		"e-1521946194448543744",
		"e-1521947552186691584"
	  ]
      query_search = "healthScore.categoryV2:bad"
	},
 ]
 notifications = ["123", "456"]
 runbook_link = "https://www.runbooklink.com"
 trigger_delay_seconds = 600
}
`, name)
}

func testOrConditionAlertResourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`

resource "swo_alert" "test" {
 name        = %[1]q
 description = ""
 severity    = "INFO"
 enabled     = false
 notification_actions = [
   {
	  configuration_ids = ["333:email", "444:msteams"]
	  resend_interval_seconds = 600
   },
 ]
 conditions_operation  = "OR"
 conditions = [
	{
	  metric_name      = "synthetics.https.response.time"
	  threshold        = ">=3000"
	  duration         = "5m"
	  not_reporting    = false
	  aggregation_type = "AVG"
	  target_entity_types = ["Website"]
	  entity_ids = [
		"e-1521946194448543744",
		"e-1521947552186691584"
	  ]
      query_search = "healthScore.categoryV2:bad"
	},
	{
	  metric_name      = "sw.metrics.healthscore"
	  threshold        = "<10"
	  duration         = "5m"
	  not_reporting    = false
	  aggregation_type = "AVG"
	  target_entity_types = ["Website"]
	  entity_ids = [
		"e-1521946194448543744",
		"e-1521947552186691584"
	  ]
      query_search = "healthScore.categoryV2:bad"
	},
	{
	  metric_name      = "synthetics.status"
	  not_reporting    = true
	  duration         = "30m"
	  aggregation_type = "COUNT"
	  target_entity_types = ["Website"]
	  entity_ids = [
		"e-1521946194448543744",
		"e-1521947552186691584"
	  ]
      query_search = "healthScore.categoryV2:bad"
	},
 ]
 notifications = ["123", "456"]
 runbook_link = "https://www.runbooklink.com"
 trigger_delay_seconds = 600
}
`, name)
}

func testAccAlertResourceNotReportingConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`

resource "swo_alert" "test" {
  name        = %[1]q
  description = "Mock alert description."
  severity    = "CRITICAL"
  enabled     = true
  notification_actions = [
    {
	  configuration_ids = ["333:email", "444:msteams"]
	  resend_interval_seconds = 600
    },
  ]
  conditions = [
	{
	  metric_name      = "synthetics.https.response.time"
      not_reporting    = true
	  duration         = "10m"
	  aggregation_type = "COUNT"
	  target_entity_types = ["Website"]
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
	  exclude_tags = [{
		  name = "service.name"
		  values : [
			"test-service"
		  ]
		}]
	},
  ]
  notifications = ["123", "456"]
  runbook_link = "https://www.runbooklink.com"
}
`, name)
}

func testAccMetricGroupAlertResourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`

resource "swo_alert" "test" {
 name        = %q
 description = "Mock alert description."
 severity    = "CRITICAL"
 enabled     = true
 notification_actions = [
   {
	  configuration_ids = ["333:email", "444:msteams"]
	  resend_interval_seconds = 600
   },
 ]
 conditions = [
	{
	  metric_name      = "synthetics.https.response.time"
	  threshold        = ">=10.5"
	  duration         = "5m"
	  not_reporting    = false
	  aggregation_type = "AVG"
      query_search = "healthScore.categoryV2:bad"
      target_entity_types = ["Uri"]
	  group_by_metric_tag = [
		"host.name"
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
	  exclude_tags = [{
		  name = "service.name"
		  operation = "CONTAINS"
		  values : [
			"test-service"
		  ]
		}]
	},
 ]
 runbook_link = "https://www.runbooklink.com"
 trigger_delay_seconds = 300
}
`, name)
}

func testAccMultiAttributeAlertResourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`

resource "swo_alert" "test" {
 name        = %[1]q
 description = "Mock alert description."
 severity    = "INFO"
 enabled     = false
 conditions = [
	{
	  attribute_name = "inMaintenance"
	  attribute_value = "true"
	  attribute_operator = "="
	  target_entity_types = ["Website"]
	},
	{
	  attribute_name = "healthScore.scoreV2"
	  attribute_values = ["0","100"]
	  attribute_operator = "IN"
	  target_entity_types = ["Website"]
	}
 ]
}
`, name)
}
