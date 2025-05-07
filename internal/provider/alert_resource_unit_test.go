package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func Test_ValidateConditions_LengthLessThanOne(t *testing.T) {

	ctx := context.Background()
	var alertConditions []alertConditionModel
	conditions, _ := types.SetValueFrom(ctx, types.ObjectType{AttrTypes: AlertConditionAttributeTypes()}, alertConditions)

	model := alertResourceModel{
		Conditions: conditions,
	}
	result := model.validateConditions(ctx)

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

	ctx := context.Background()
	alertConditions := []alertConditionModel{
		{}, {}, {}, {}, {}, {},
	}
	conditions, _ := types.SetValueFrom(ctx, types.ObjectType{AttrTypes: AlertConditionAttributeTypes()}, alertConditions)

	model := alertResourceModel{
		Conditions: conditions,
	}
	result := model.validateConditions(ctx)

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

	alertConditions := []alertConditionModel{
		{
			MetricName:        types.StringValue("metric_name"),
			Threshold:         types.StringValue("<300"),
			Duration:          types.StringValue("10s"),
			AggregationType:   types.StringValue("AVG"),
			EntityIds:         types.ListNull(attr.Type(types.StringType)),
			QuerySearch:       types.StringNull(),
			TargetEntityTypes: targetEntityTypes,
			IncludeTags:       types.SetNull(types.ObjectType{AttrTypes: AlertTagAttributeTypes()}),
			ExcludeTags:       types.SetNull(types.ObjectType{AttrTypes: AlertTagAttributeTypes()}),
			GroupByMetricTag:  types.ListNull(attr.Type(types.StringType)),
			NotReporting:      types.BoolValue(false),
		},
	}
	ctx := context.Background()
	conditions, _ := types.SetValueFrom(ctx, types.ObjectType{AttrTypes: AlertConditionAttributeTypes()}, alertConditions)
	model := alertResourceModel{
		Conditions: conditions,
	}

	diagnosticError := model.validateConditions(ctx)

	if len(diagnosticError) != 0 {
		t.Fatal("expected 0 diagnosticError")
	}
}

func Test_ValidateCondition_NotReporting(t *testing.T) {
	entities := []attr.Value{types.StringValue("Website")}
	targetEntityTypes, _ := types.ListValue(types.StringType, entities)
	alertConditions := []alertConditionModel{
		{
			MetricName:        types.StringValue("metric_name"),
			Threshold:         types.StringValue("<300"), // should be ""
			Duration:          types.StringValue("10s"),
			AggregationType:   types.StringValue("AVG"), // should be COUNT
			EntityIds:         types.ListNull(attr.Type(types.StringType)),
			QuerySearch:       types.StringNull(),
			TargetEntityTypes: targetEntityTypes,
			IncludeTags:       types.SetNull(types.ObjectType{AttrTypes: AlertTagAttributeTypes()}),
			ExcludeTags:       types.SetNull(types.ObjectType{AttrTypes: AlertTagAttributeTypes()}),
			GroupByMetricTag:  types.ListNull(attr.Type(types.StringType)),
			NotReporting:      types.BoolValue(true),
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

	ctx := context.Background()
	conditions, _ := types.SetValueFrom(ctx, types.ObjectType{AttrTypes: AlertConditionAttributeTypes()}, alertConditions)
	model := alertResourceModel{
		Conditions: conditions,
	}
	result := model.validateConditions(ctx)

	if !result.Equal(expected) {
		t.Errorf("expected %v but got %v", expected, result)
	}
}

func Test_ValidateCondition_Reporting(t *testing.T) {
	entities := []attr.Value{types.StringValue("Website")}
	targetEntityTypes, _ := types.ListValue(types.StringType, entities)
	alertConditions := []alertConditionModel{
		{
			MetricName:        types.StringValue("metric_name"),
			Threshold:         types.StringValue(""), // is required
			Duration:          types.StringValue("10s"),
			AggregationType:   types.StringValue("AVG"),
			EntityIds:         types.ListNull(attr.Type(types.StringType)),
			QuerySearch:       types.StringNull(),
			TargetEntityTypes: targetEntityTypes,
			IncludeTags:       types.SetNull(types.ObjectType{AttrTypes: AlertTagAttributeTypes()}),
			ExcludeTags:       types.SetNull(types.ObjectType{AttrTypes: AlertTagAttributeTypes()}),
			GroupByMetricTag:  types.ListNull(attr.Type(types.StringType)),
			NotReporting:      types.BoolValue(false),
		},
	}

	expected := diag.Diagnostics{
		diag.NewAttributeErrorDiagnostic(
			path.Root("threshold"),
			"Required field when not_reporting is set to false.",
			"Required field when not_reporting is set to false."),
	}

	ctx := context.Background()
	conditions, _ := types.SetValueFrom(ctx, types.ObjectType{AttrTypes: AlertConditionAttributeTypes()}, alertConditions)
	model := alertResourceModel{
		Conditions: conditions,
	}
	result := model.validateConditions(ctx)

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

	alertConditions := []alertConditionModel{
		{
			MetricName:        types.StringValue("metric_name"),
			Threshold:         types.StringValue("<300"),
			Duration:          types.StringValue("10s"),
			AggregationType:   types.StringValue("AVG"),
			EntityIds:         entityIds0,
			QuerySearch:       query0,
			TargetEntityTypes: targetEntityTypes0,
			IncludeTags:       types.SetNull(types.ObjectType{AttrTypes: AlertTagAttributeTypes()}),
			ExcludeTags:       types.SetNull(types.ObjectType{AttrTypes: AlertTagAttributeTypes()}),
			GroupByMetricTag:  groupByMetricTag0,
			NotReporting:      types.BoolValue(false),
		},
		{
			// same values as node 0
			MetricName:        types.StringValue("metric_name"),
			Threshold:         types.StringValue(""),
			Duration:          types.StringValue("10s"),
			AggregationType:   types.StringValue("COUNT"),
			EntityIds:         entityIds0,
			QuerySearch:       query0,
			TargetEntityTypes: targetEntityTypes0,
			IncludeTags:       types.SetNull(types.ObjectType{AttrTypes: AlertTagAttributeTypes()}),
			ExcludeTags:       types.SetNull(types.ObjectType{AttrTypes: AlertTagAttributeTypes()}),
			GroupByMetricTag:  groupByMetricTag0,
			NotReporting:      types.BoolValue(true),
		},
		{
			// different values from node 0
			MetricName:        types.StringValue("metric_name"),
			Threshold:         types.StringValue("<300"),
			Duration:          types.StringValue("10s"),
			AggregationType:   types.StringValue("AVG"),
			EntityIds:         entityIds1,
			QuerySearch:       query1,
			TargetEntityTypes: targetEntityTypes1,
			IncludeTags:       types.SetNull(types.ObjectType{AttrTypes: AlertTagAttributeTypes()}),
			ExcludeTags:       types.SetNull(types.ObjectType{AttrTypes: AlertTagAttributeTypes()}),
			GroupByMetricTag:  groupByMetricTag1,
			NotReporting:      types.BoolValue(false),
		},
	}

	ctx := context.Background()
	conditions, _ := types.SetValueFrom(ctx, types.ObjectType{AttrTypes: AlertConditionAttributeTypes()}, alertConditions)
	model := alertResourceModel{Conditions: conditions}
	result := model.validateConditions(ctx)

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
