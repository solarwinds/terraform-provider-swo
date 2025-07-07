package provider

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"regexp"
	"strconv"

	swoClient "github.com/solarwinds/swo-client-go/pkg/client"
)

var thresholdOperatorError = errors.New("threshold operation not found")
var thresholdValueError = errors.New("threshold value not found")
var aggregationError = errors.New("aggregation operation not found")

// Builds a simple metric condition.
//
// An example of a simple metric condition tree:
//
//	                       >=
//	            (threshold operator, id=0)
//	                      /  \
//	                    AVG    42
//	     (aggregation, id=1)   (threshold data, id=4)
//	         /    \
//	Metric Field   10m
//	    (id=2)    (duration, id=3)
func (model alertConditionModel) toAlertConditionInputs(ctx context.Context, diags *diag.Diagnostics, rootNodeId int) []swoClient.AlertConditionNodeInput {

	if !model.MetricName.IsNull() && !model.AttributeName.IsNull() {
		diags.AddError("Bad input in terraform resource",
			fmt.Sprintf("Alerting condition must be either metric or attribute. Cannot populate both metric_name and attribute_name."))
		return []swoClient.AlertConditionNodeInput{}
	}

	if model.MetricName.IsNull() && model.AttributeName.IsNull() {
		diags.AddError("Bad input in terraform resource",
			fmt.Sprintf("Alerting condition must be either metric or attribute. Must populate either metric_name or attribute_name."))
		return []swoClient.AlertConditionNodeInput{}
	}

	// metric condition node
	// for both metric AND group alerts
	if !model.MetricName.IsNull() {
		//binary/threshold operator
		thresholdOperatorCondition, thresholdDataCondition, err := model.toThresholdConditionInputs()
		if err != nil {
			diags.AddError("Bad input in terraform resource",
				fmt.Sprintf("error parsing terraform resource: %s", err))
			return []swoClient.AlertConditionNodeInput{}
		}
		thresholdOperatorCondition.Id = rootNodeId
		thresholdOperatorCondition.OperandIds = []int{rootNodeId + 1, rootNodeId + 4}

		//aggregation operator
		aggregationCondition, err := model.toAggregationConditionInput()
		if err != nil {
			diags.AddError("Bad input in terraform resource",
				fmt.Sprintf("error parsing terraform resource: %s", err))
			return []swoClient.AlertConditionNodeInput{}
		}
		aggregationCondition.Id = rootNodeId + 1
		aggregationCondition.OperandIds = []int{rootNodeId + 2, rootNodeId + 3}

		//metric field node
		metricFieldCondition := model.toMetricFieldConditionInput(ctx, diags)
		if diags.HasError() {
			return []swoClient.AlertConditionNodeInput{}
		}
		metricFieldCondition.Id = rootNodeId + 2

		//constant value/duration condition
		durationCondition := model.toDurationConditionInput()
		durationCondition.Id = rootNodeId + 3

		//constant value/threshold condition
		thresholdDataCondition.Id = rootNodeId + 4

		conditions := []swoClient.AlertConditionNodeInput{
			thresholdOperatorCondition,
			aggregationCondition,
			metricFieldCondition,
			durationCondition,
			thresholdDataCondition,
		}

		return conditions
	}

	// attribute condition node
	// for metric alerts ONLY
	if !model.AttributeName.IsNull() {

		binaryCondition := swoClient.AlertConditionNodeInput{
			Type:       string(swoClient.AlertBinaryOperatorType),
			Operator:   model.AttributeOperator.ValueStringPointer(),
			OperandIds: []int{rootNodeId + 1, rootNodeId + 2},
		}

		entityFilter := model.buildEntityFilter(ctx, diags)
		attributeField := swoClient.AlertConditionNodeInput{
			Id:           rootNodeId + 1,
			Type:         string(swoClient.AlertAttributeType),
			FieldName:    model.AttributeName.ValueStringPointer(),
			EntityFilter: entityFilter,
		}

		//todo i think we need to do something special for arrays of ints/floats?
		dataType := GetStringDataType(model.AttributeValue.ValueString())
		constantField := swoClient.AlertConditionNodeInput{
			Id:       rootNodeId + 2,
			Type:     string(swoClient.AlertConstantValueType),
			DataType: &dataType,
		}
		operator := model.AttributeOperator.ValueString()
		if operator == "IN" {
			constantField.Values = []string{model.AttributeValue.ValueString()} //todo this isn't right, but good enough for testing?
		} else {
			constantField.Value = model.AttributeValue.ValueStringPointer()
		}

		conditions := []swoClient.AlertConditionNodeInput{
			binaryCondition,
			attributeField,
			constantField,
		}

		return conditions
	}

	return []swoClient.AlertConditionNodeInput{}
}

// Creates the threshold operation and threshold data nodes by either:
//  1. If model.not_reporting=true, operator is set to '=' and value to '0'
//  2. Else, parse the model.threshold string into operator and value
//     Ex:">=3000" -> operator '>=' and value '3000'
func (model alertConditionModel) toThresholdConditionInputs() (swoClient.AlertConditionNodeInput, swoClient.AlertConditionNodeInput, error) {
	threshold := model.Threshold.ValueString()
	thresholdOperatorConditions := swoClient.AlertConditionNodeInput{}
	thresholdDataConditions := swoClient.AlertConditionNodeInput{}

	//Not Reporting threshold values
	if model.NotReporting.ValueBool() {

		operator := string(swoClient.AlertOperatorEq)
		thresholdOperatorConditions.Type = string(swoClient.AlertBinaryOperatorType)
		thresholdOperatorConditions.Operator = &operator

		dataValue := "0"
		dataType := GetStringDataType(dataValue)
		thresholdDataConditions.Type = string(swoClient.AlertConstantValueType)
		thresholdDataConditions.DataType = &dataType
		thresholdDataConditions.Value = &dataValue

	} else {

		regex := regexp.MustCompile(`\W+`)
		operator := regex.FindString(threshold)
		//Parses threshold into an operator:(>, <, = ...).

		operatorType, err := swoClient.GetAlertConditionType(operator)
		if err != nil {
			return thresholdOperatorConditions, thresholdDataConditions, thresholdOperatorError
		}
		thresholdOperatorConditions.Type = operatorType
		thresholdOperatorConditions.Operator = &operator

		regex = regexp.MustCompile("[0-9]+")
		thresholdValue := regex.FindString(threshold)
		//Parses threshold into numbers:(3000, 200, 10...).

		if thresholdValue != "" {
			dataType := GetStringDataType(thresholdValue)

			thresholdDataConditions.Type = string(swoClient.AlertConstantValueType)
			thresholdDataConditions.DataType = &dataType
			thresholdDataConditions.Value = &thresholdValue
		} else {
			return thresholdOperatorConditions, thresholdDataConditions, thresholdValueError
		}
	}

	return thresholdOperatorConditions, thresholdDataConditions, nil
}

func (model alertConditionModel) toDurationConditionInput() swoClient.AlertConditionNodeInput {

	duration := model.Duration.ValueString()
	dataType := GetStringDataType(duration)
	durationCondition := swoClient.AlertConditionNodeInput{
		Type:     string(swoClient.AlertConstantValueType),
		DataType: &dataType,
		Value:    &duration,
	}

	return durationCondition
}

func (model alertConditionModel) toAggregationConditionInput() (swoClient.AlertConditionNodeInput, error) {

	operator := model.AggregationType.ValueString()
	operatorType, err := swoClient.GetAlertConditionType(operator)
	if err != nil {
		return swoClient.AlertConditionNodeInput{}, aggregationError
	}

	aggregationCondition := swoClient.AlertConditionNodeInput{
		Type:     operatorType,
		Operator: &operator,
	}

	return aggregationCondition, nil
}

func (model alertConditionModel) toMetricFieldConditionInput(ctx context.Context, diags *diag.Diagnostics) swoClient.AlertConditionNodeInput {

	var groupByMetricTag []string
	d := model.GroupByMetricTag.ElementsAs(ctx, &groupByMetricTag, false)
	diags.Append(d...)
	if diags.HasError() {
		return swoClient.AlertConditionNodeInput{}
	}
	metricName := model.MetricName.ValueString()
	entityFilter := model.buildEntityFilter(ctx, diags)

	metricFieldCondition := swoClient.AlertConditionNodeInput{
		Type:             string(swoClient.AlertMetricFieldType),
		FieldName:        &metricName,
		GroupByMetricTag: groupByMetricTag,
		EntityFilter:     entityFilter,
	}

	var includeTags []alertTagsModel
	var excludeTags []alertTagsModel
	d = model.IncludeTags.ElementsAs(ctx, &includeTags, false)
	diags.Append(d...)
	if diags.HasError() {
		return swoClient.AlertConditionNodeInput{}
	}
	d = model.ExcludeTags.ElementsAs(ctx, &excludeTags, false)
	diags.Append(d...)
	if diags.HasError() {
		return swoClient.AlertConditionNodeInput{}
	}

	if len(includeTags) > 0 || len(excludeTags) > 0 {
		metricFieldCondition.MetricFilter = &swoClient.AlertFilterExpressionInput{
			Operation: swoClient.FilterOperationAnd,
		}
	}

	for _, tag := range includeTags {
		var metricFilterPropertyValues []*string
		dFilter := tag.Values.ElementsAs(ctx, &metricFilterPropertyValues, false)
		diags.Append(dFilter...)
		if diags.HasError() {
			return swoClient.AlertConditionNodeInput{}
		}
		propertyName := tag.Name.ValueString()
		metricFilter := swoClient.AlertFilterExpressionInput{
			PropertyName:   &propertyName,
			Operation:      swoClient.FilterOperationIn,
			PropertyValues: metricFilterPropertyValues,
		}

		metricFieldCondition.MetricFilter.Children = append(
			metricFieldCondition.MetricFilter.Children,
			metricFilter,
		)
	}

	for _, tag := range excludeTags {
		var propertyValues []*string
		dValues := tag.Values.ElementsAs(ctx, &propertyValues, false)
		diags.Append(dValues...)
		if diags.HasError() {
			return swoClient.AlertConditionNodeInput{}
		}
		propertyName := tag.Name.ValueString()

		metricFilter := swoClient.AlertFilterExpressionInput{
			PropertyName:   &propertyName,
			Operation:      swoClient.FilterOperationIn,
			PropertyValues: propertyValues,
		}

		metricFilterNotOp := swoClient.AlertFilterExpressionInput{
			Operation: swoClient.FilterOperationNot,
		}

		metricFilterNotOp.Children = append(metricFilterNotOp.Children, metricFilter)

		metricFieldCondition.MetricFilter.Children = append(
			metricFieldCondition.MetricFilter.Children,
			metricFilterNotOp,
		)
	}

	return metricFieldCondition
}

func (model alertConditionModel) buildEntityFilter(ctx context.Context, diags *diag.Diagnostics) *swoClient.AlertConditionNodeEntityFilterInput {
	// Metric Group alerts do not use entity types. It is necessary to drop the entire entityFilter field
	// when calling the Alerting API because presence/absence of this field determines the type of the
	// Alert definition (Entity vs. Metric Group).

	if !model.TargetEntityTypes.IsNull() {
		var entityFilterTypes, entityFilterIds []string
		d := model.TargetEntityTypes.ElementsAs(ctx, &entityFilterTypes, false)
		diags.Append(d...)

		if diags.HasError() {
			return nil
		}
		d = model.EntityIds.ElementsAs(ctx, &entityFilterIds, false)
		diags.Append(d...)
		if diags.HasError() {
			return nil
		}
		querySearch := model.QuerySearch.ValueString()

		entityFilter := swoClient.AlertConditionNodeEntityFilterInput{
			Types: entityFilterTypes,
			Ids:   entityFilterIds,
			Query: &querySearch,
		}
		return &entityFilter
	} else {
		return nil
	}
}

func GetStringDataType(s string) string {

	if _, err := strconv.Atoi(s); err == nil {
		return "number"
	}

	if _, err := strconv.ParseFloat(s, 64); err == nil {
		return "number"
	}

	if _, err := strconv.ParseBool(s); err == nil {
		return "boolean"
	}

	return "string"
}
