package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

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
