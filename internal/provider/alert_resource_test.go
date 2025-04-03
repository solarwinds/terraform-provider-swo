package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAlertResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		IsUnitTest:               true,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccAlertResourceConfig("test-acc Mock Alert Name"),
				Check: resource.ComposeAggregateTestCheckFunc(
					//resource.TestCheckResourceAttr("swo_alert.test", "id", "0bc4710d-e3b0-4590-9c9b-e5e46d81d912"),
					resource.TestCheckResourceAttr("swo_alert.test", "name", "test-acc Mock Alert Name"),
					resource.TestCheckResourceAttr("swo_alert.test", "description", "Mock alert description."),
					resource.TestCheckResourceAttr("swo_alert.test", "severity", "CRITICAL"),
					resource.TestCheckResourceAttr("swo_alert.test", "trigger_reset_actions", "false"),
					// Verify actions
					resource.TestCheckResourceAttr("swo_alert.test", "notification_actions.0.configuration_ids.0", "333:email"),
					resource.TestCheckResourceAttr("swo_alert.test", "notification_actions.0.configuration_ids.1", "444:msteams"),
					resource.TestCheckResourceAttr("swo_alert.test", "notification_actions.0.resend_interval_seconds", "600"),
					// Verify number of conditions.
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.#", "1"),
					// Verify the conditions.
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.target_entity_types.0", "Website"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.metric_name", "synthetics.https.response.time"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.threshold", ">=3000ms"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.not_reporting", "false"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.duration", "5m"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.aggregation_type", "AVG"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.entity_ids.0", "e-1521946194448543744"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.entity_ids.1", "e-1521947552186691584"),
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
				Config: testAccAlertResourceConfig("test-acc test_two"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("swo_alert.test", "name", "test-acc test_two"),
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
					// Verify number of conditions.
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.#", "1"),
					// Verify the conditions.
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.target_entity_types.0", "Website"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.metric_name", "synthetics.https.response.time"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.threshold", ""),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.not_reporting", "true"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.duration", "10m"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.aggregation_type", "COUNT"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.entity_ids.0", "e-1521946194448543744"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.entity_ids.1", "e-1521947552186691584"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.group_by_metric_tag.0", "host.name"),
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

func TestMultiConditionAlertResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		IsUnitTest:               true,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testMultiConditionAlertResourceConfig("test-acc Mock Multi Condition Alert Name"),
				Check: resource.ComposeAggregateTestCheckFunc(

					resource.TestCheckResourceAttr("swo_alert.test", "name", "test-acc Mock Multi Condition Alert Name"),
					resource.TestCheckResourceAttr("swo_alert.test", "description", ""),
					resource.TestCheckResourceAttr("swo_alert.test", "severity", "INFO"),
					resource.TestCheckResourceAttr("swo_alert.test", "enabled", "false"),
					// Verify actions
					resource.TestCheckResourceAttr("swo_alert.test", "notification_actions.0.configuration_ids.0", "333:email"),
					resource.TestCheckResourceAttr("swo_alert.test", "notification_actions.0.configuration_ids.1", "444:msteams"),
					resource.TestCheckResourceAttr("swo_alert.test", "notification_actions.0.resend_interval_seconds", "600"),
					// Verify number of conditions.
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.#", "3"),
					// Verify the conditions.
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.target_entity_types.0", "Website"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.metric_name", "sw.metrics.healthscore"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.threshold", "<10"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.not_reporting", "false"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.duration", "5m"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.aggregation_type", "AVG"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.entity_ids.0", "e-1521946194448543744"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.entity_ids.1", "e-1521947552186691584"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.0.group_by_metric_tag.0", "host.name"),

					resource.TestCheckResourceAttr("swo_alert.test", "conditions.1.target_entity_types.0", "Website"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.1.metric_name", "synthetics.https.response.time"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.1.threshold", ">=3000ms"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.1.not_reporting", "false"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.1.duration", "5m"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.1.aggregation_type", "AVG"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.1.entity_ids.0", "e-1521946194448543744"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.1.entity_ids.1", "e-1521947552186691584"),

					resource.TestCheckResourceAttr("swo_alert.test", "conditions.2.target_entity_types.0", "Website"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.2.metric_name", "synthetics.status"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.2.threshold", ""),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.2.not_reporting", "true"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.2.duration", "30m"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.2.aggregation_type", "COUNT"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.2.entity_ids.0", "e-1521946194448543744"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.2.entity_ids.1", "e-1521947552186691584"),
					resource.TestCheckResourceAttr("swo_alert.test", "conditions.2.group_by_metric_tag.0", "host.name"),

					resource.TestCheckResourceAttr("swo_alert.test", "notifications.0", "123"),
					resource.TestCheckResourceAttr("swo_alert.test", "notifications.1", "456"),
					resource.TestCheckResourceAttr("swo_alert.test", "runbook_link", "https://www.runbooklink.com"),
					resource.TestCheckResourceAttr("swo_alert.test", "trigger_delay_seconds", "600"),
				),
			},
			// Update and Read testing
			{
				Config: testMultiConditionAlertResourceConfig("test-acc test_two"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("swo_alert.test", "name", "test-acc test_two"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccAlertResourceConfig(name string) string {
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
	  threshold        = ">=3000ms"
	  duration         = "5m"
	  not_reporting    = false
	  aggregation_type = "AVG"
	  target_entity_types = ["Website"]
	  entity_ids = [
		"e-1521946194448543744",
		"e-1521947552186691584"
	  ]
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

func testMultiConditionAlertResourceConfig(name string) string {
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
	  threshold        = ">=3000ms"
	  duration         = "5m"
	  not_reporting    = false
	  aggregation_type = "AVG"
	  target_entity_types = ["Website"]
	  entity_ids = [
		"e-1521946194448543744",
		"e-1521947552186691584"
	  ]
	  group_by_metric_tag = [
		"host.name"
	  ]
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
	  group_by_metric_tag = [
		"host.name"
	  ]
	},
	{
	  metric_name      = "synthetics.status"
	  threshold        = ""
	  not_reporting    = true
	  duration         = "30m"
	  aggregation_type = "COUNT"
	  target_entity_types = ["Website"]
	  entity_ids = [
		"e-1521946194448543744",
		"e-1521947552186691584"
	  ]
	  group_by_metric_tag = [
		"host.name"
	  ]
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
	  threshold        = ""
      not_reporting    = true
	  duration         = "10m"
	  aggregation_type = "COUNT"
	  target_entity_types = ["Website"]
	  entity_ids = [
		"e-1521946194448543744",
		"e-1521947552186691584"
	  ]
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
}
`, name)
}

func Test_ValidateConditions_LengthLessThanOne(t *testing.T) {

	model := alertResourceModel{
		Conditions: []alertConditionModel{},
	}
	diagnosticError := model.validateConditions()

	expected := []diagnosticsError{
		{
			attributeName: "conditions",
			summary:       "Invalid number of alerting conditions.",
			details:       "Number of alerting conditions must be between 1 and 5.",
		},
	}
	if len(diagnosticError) != len(expected) {
		t.Fatalf("expected %v diagnosticErrors", len(expected))
	}

	for i := 0; i < len(expected); i++ {
		if diagnosticError[i] != expected[i] {
			t.Fatalf("expected(%v, %v, %v) unexpected(%v, %v, %v) ",
				expected[i].attributeName, expected[i].summary, expected[i].details,
				diagnosticError[i].attributeName, diagnosticError[i].summary, diagnosticError[i].details)
		}
	}
}

func Test_ValidateConditions_LengthGreaterThanFive(t *testing.T) {

	model := alertResourceModel{
		Conditions: []alertConditionModel{
			{}, {}, {}, {}, {}, {},
		},
	}
	diagnosticError := model.validateConditions()

	expected := []diagnosticsError{
		{
			attributeName: "conditions",
			summary:       "Invalid number of alerting conditions.",
			details:       "Number of alerting conditions must be between 1 and 5.",
		},
	}
	if len(diagnosticError) != len(expected) {
		t.Fatalf("expected %v diagnosticErrors", len(expected))
	}

	for i := 0; i < len(expected); i++ {
		if diagnosticError[i] != expected[i] {
			t.Fatalf("expected(%v, %v, %v) unexpected(%v, %v, %v) ",
				expected[i].attributeName, expected[i].summary, expected[i].details,
				diagnosticError[i].attributeName, diagnosticError[i].summary, diagnosticError[i].details)
		}
	}
}

func Test_ValidateCondition_HappyPath(t *testing.T) {
	entities := []attr.Value{types.StringValue("Website")}
	targetEntityTypes, _ := types.ListValue(types.StringType, entities)

	model := alertResourceModel{
		Conditions: []alertConditionModel{
			{
				NotReporting:      types.BoolValue(false),
				Threshold:         types.StringValue("<300"),
				AggregationType:   types.StringValue("AVG"),
				TargetEntityTypes: targetEntityTypes,
				EntityIds:         types.ListNull(attr.Type(types.StringType)),
				GroupByMetricTag:  types.ListNull(attr.Type(types.StringType)),
			},
		},
	}
	diagnosticError := model.validateConditions()

	if len(diagnosticError) != 0 {
		t.Fatal("expected 0 diagnosticError")
	}
}

func Test_ValidateCondition_NotReporting(t *testing.T) {
	entities := []attr.Value{types.StringValue("Website")}
	targetEntityTypes, _ := types.ListValue(types.StringType, entities)
	model := alertResourceModel{
		Conditions: []alertConditionModel{
			{
				NotReporting:      types.BoolValue(true),
				Threshold:         types.StringValue("<300"), // should be ""
				AggregationType:   types.StringValue("AVG"),  // should be COUNT
				TargetEntityTypes: targetEntityTypes,
				EntityIds:         types.ListNull(attr.Type(types.StringType)),
				GroupByMetricTag:  types.ListNull(attr.Type(types.StringType)),
			},
		},
	}
	expected := []diagnosticsError{
		{
			attributeName: "threshold",
			summary:       "Cannot set threshold when not_reporting is set to true.",
			details:       "Cannot set threshold when not_reporting is set to true.",
		},
		{
			attributeName: "aggregationType",
			summary:       "Aggregation type must be COUNT when not_reporting is set to true.",
			details:       "Aggregation type must be COUNT when not_reporting is set to true.",
		},
	}

	diagnosticError := model.validateConditions()

	if len(diagnosticError) != len(expected) {
		t.Fatalf("expected %v diagnosticErrors", len(expected))
	}

	for i := 0; i < len(expected); i++ {
		if diagnosticError[i] != expected[i] {
			t.Fatalf("expected(%v, %v, %v) unexpected(%v, %v, %v) ",
				expected[i].attributeName, expected[i].summary, expected[i].details,
				diagnosticError[i].attributeName, diagnosticError[i].summary, diagnosticError[i].details)
		}
	}
}

func Test_ValidateCondition_Reporting(t *testing.T) {
	entities := []attr.Value{types.StringValue("Website")}
	targetEntityTypes, _ := types.ListValue(types.StringType, entities)
	model := alertResourceModel{
		Conditions: []alertConditionModel{
			{
				NotReporting:      types.BoolValue(false),
				Threshold:         types.StringValue(""), // is required
				AggregationType:   types.StringValue("AVG"),
				TargetEntityTypes: targetEntityTypes,
				EntityIds:         types.ListNull(attr.Type(types.StringType)),
				GroupByMetricTag:  types.ListNull(attr.Type(types.StringType)),
			},
		},
	}
	diagnosticError := model.validateConditions()

	expected := []diagnosticsError{
		{
			attributeName: "threshold",
			summary:       "Required field when not_reporting is set to false.",
			details:       "Required field when not_reporting is set to false.",
		},
	}
	if len(diagnosticError) != len(expected) {
		t.Fatalf("expected %v diagnosticErrors", len(expected))
	}

	for i := 0; i < len(expected); i++ {
		if diagnosticError[i] != expected[i] {
			t.Fatalf("expected(%v, %v, %v) unexpected(%v, %v, %v) ",
				expected[i].attributeName, expected[i].summary, expected[i].details,
				diagnosticError[i].attributeName, diagnosticError[i].summary, diagnosticError[i].details)
		}
	}
}

func Test_ValidateCondition_CompareLists(t *testing.T) {
	entities0 := []attr.Value{types.StringValue("Website")}
	targetEntityTypes0, _ := types.ListValue(types.StringType, entities0)

	ids0 := []attr.Value{types.StringValue("123")}
	entityIds0, _ := types.ListValue(types.StringType, ids0)

	tags0 := []attr.Value{types.StringValue("tags.names")}
	groupByMetricTag0, _ := types.ListValue(types.StringType, tags0)

	entities1 := []attr.Value{types.StringValue("Uri")}
	targetEntityTypes1, _ := types.ListValue(types.StringType, entities1)

	ids1 := []attr.Value{types.StringValue("456")}
	entityIds1, _ := types.ListValue(types.StringType, ids1)

	tags1 := []attr.Value{types.StringValue("tags.environment")}
	groupByMetricTag1, _ := types.ListValue(types.StringType, tags1)

	model := alertResourceModel{
		Conditions: []alertConditionModel{
			{
				NotReporting:      types.BoolValue(false),
				Threshold:         types.StringValue("<300"),
				AggregationType:   types.StringValue("AVG"),
				TargetEntityTypes: targetEntityTypes0,
				EntityIds:         entityIds0,
				GroupByMetricTag:  groupByMetricTag0,
			},
			{
				NotReporting:    types.BoolValue(true),
				Threshold:       types.StringValue(""),
				AggregationType: types.StringValue("COUNT"),
				// same []types.List as node 0
				TargetEntityTypes: targetEntityTypes0,
				EntityIds:         entityIds0,
				GroupByMetricTag:  groupByMetricTag0,
			},
			{
				NotReporting:    types.BoolValue(false),
				Threshold:       types.StringValue("<300"),
				AggregationType: types.StringValue("AVG"),
				// different []types.List from node 0
				TargetEntityTypes: targetEntityTypes1,
				EntityIds:         entityIds1,
				GroupByMetricTag:  groupByMetricTag1,
			},
		},
	}
	diagnosticError := model.validateConditions()

	expected := []diagnosticsError{
		{
			attributeName: "targetEntityTypes",
			summary:       "The list must be same for all conditions",
			details:       "The list must be same for all conditions, but [\"Website\"] does not match [\"Uri\"].",
		},
		{
			attributeName: "entityIds",
			summary:       "The list must be same for all conditions",
			details:       "The list must be same for all conditions, but [\"123\"] does not match [\"456\"].",
		},
		{
			attributeName: "groupByMetricTag",
			summary:       "The list must be same for all conditions",
			details:       "The list must be same for all conditions, but [\"tags.names\"] does not match [\"tags.environment\"].",
		},
	}

	if len(diagnosticError) != len(expected) {
		t.Fatalf("expected %v diagnosticErrors", len(expected))
	}

	for i := 0; i < len(expected); i++ {
		if diagnosticError[i] != expected[i] {
			t.Fatalf("expected(%v, %v, %v) unexpected(%v, %v, %v) ",
				expected[i].attributeName, expected[i].summary, expected[i].details,
				diagnosticError[i].attributeName, diagnosticError[i].summary, diagnosticError[i].details)
		}
	}
}
