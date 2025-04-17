package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func Test_ValidateConditions_LengthLessThanOne(t *testing.T) {

	model := alertResourceModel{
		Conditions: []alertConditionModel{},
	}
	result := model.validateConditions()

	expected := diag.Diagnostics{
		diag.NewAttributeErrorDiagnostic(
			path.Root("conditions"),
			"Invalid number of alerting conditions.",
			"Number of alerting conditions must be between 1 and 5."),
	}

	if !result.Equal(expected) {
		t.Errorf("expected %v but got %v", expected, result)
	}
}

func Test_ValidateConditions_LengthGreaterThanFive(t *testing.T) {

	model := alertResourceModel{
		Conditions: []alertConditionModel{
			{}, {}, {}, {}, {}, {},
		},
	}
	result := model.validateConditions()

	expected := diag.Diagnostics{
		diag.NewAttributeErrorDiagnostic(
			path.Root("conditions"),
			"Invalid number of alerting conditions.",
			"Number of alerting conditions must be between 1 and 5."),
	}

	if !result.Equal(expected) {
		t.Errorf("expected %v but got %v", expected, result)
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

	expected := diag.Diagnostics{
		diag.NewAttributeErrorDiagnostic(
			path.Root("threshold"),
			"Cannot set threshold when not_reporting is set to true.",
			"Cannot set threshold when not_reporting is set to true."),
		diag.NewAttributeErrorDiagnostic(
			path.Root("aggregationType"),
			"Aggregation type must be COUNT when not_reporting is set to true.",
			"Aggregation type must be COUNT when not_reporting is set to true."),
	}

	result := model.validateConditions()

	if !result.Equal(expected) {
		t.Errorf("expected %v but got %v", expected, result)
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

	expected := diag.Diagnostics{
		diag.NewAttributeErrorDiagnostic(
			path.Root("threshold"),
			"Required field when not_reporting is set to false.",
			"Required field when not_reporting is set to false."),
	}

	result := model.validateConditions()

	if !result.Equal(expected) {
		t.Errorf("expected %v but got %v", expected, result)
	}
}

func Test_ValidateCondition_CompareLists(t *testing.T) {
	entities0 := []attr.Value{types.StringValue("Website")}
	targetEntityTypes0, _ := types.ListValue(types.StringType, entities0)

	ids0 := []attr.Value{types.StringValue("123")}
	entityIds0, _ := types.ListValue(types.StringType, ids0)

	tags0 := []attr.Value{types.StringValue("tags.names")}
	groupByMetricTag0, _ := types.ListValue(types.StringType, tags0)

	query0 := types.StringValue("healthScore.categoryV2:bad")

	entities1 := []attr.Value{types.StringValue("Uri")}
	targetEntityTypes1, _ := types.ListValue(types.StringType, entities1)

	ids1 := []attr.Value{types.StringValue("456")}
	entityIds1, _ := types.ListValue(types.StringType, ids1)

	tags1 := []attr.Value{types.StringValue("tags.environment")}
	groupByMetricTag1, _ := types.ListValue(types.StringType, tags1)

	query1 := types.StringValue("healthScore.categoryV2:good")

	model := alertResourceModel{
		Conditions: []alertConditionModel{
			{
				NotReporting:      types.BoolValue(false),
				Threshold:         types.StringValue("<300"),
				AggregationType:   types.StringValue("AVG"),
				TargetEntityTypes: targetEntityTypes0,
				EntityIds:         entityIds0,
				QuerySearch:       query0,
				GroupByMetricTag:  groupByMetricTag0,
			},
			{
				NotReporting:    types.BoolValue(true),
				Threshold:       types.StringValue(""),
				AggregationType: types.StringValue("COUNT"),
				// same []types.List as node 0
				TargetEntityTypes: targetEntityTypes0,
				EntityIds:         entityIds0,
				QuerySearch:       query0,
				GroupByMetricTag:  groupByMetricTag0,
			},
			{
				NotReporting:    types.BoolValue(false),
				Threshold:       types.StringValue("<300"),
				AggregationType: types.StringValue("AVG"),
				// different []types.List from node 0
				TargetEntityTypes: targetEntityTypes1,
				EntityIds:         entityIds1,
				QuerySearch:       query1,
				GroupByMetricTag:  groupByMetricTag1,
			},
		},
	}
	result := model.validateConditions()

	expected := diag.Diagnostics{
		diag.NewAttributeErrorDiagnostic(
			path.Root("targetEntityTypes"),
			"The list must be same for all conditions",
			"The list must be same for all conditions, but [\"Website\"] does not match [\"Uri\"]."),
		diag.NewAttributeErrorDiagnostic(
			path.Root("entityIds"),
			"The list must be same for all conditions",
			"The list must be same for all conditions, but [\"123\"] does not match [\"456\"]."),
		diag.NewAttributeErrorDiagnostic(
			path.Root("querySearch"),
			"Query search must be same for all conditions",
			"Query search must be the same for all conditions, but \"healthScore.categoryV2:bad\" does not match \"healthScore.categoryV2:good\"."),
		diag.NewAttributeErrorDiagnostic(
			path.Root("groupByMetricTag"),
			"The list must be same for all conditions",
			"The list must be same for all conditions, but [\"tags.names\"] does not match [\"tags.environment\"]."),
	}

	if !result.Equal(expected) {
		t.Errorf("expected %v but got %v", expected, result)
	}
}
