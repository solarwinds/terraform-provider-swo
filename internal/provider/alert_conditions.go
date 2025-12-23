package provider

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/diag"

	swoClient "github.com/solarwinds/swo-client-go/pkg/client"
)

var (
	errThresholdParse    = errors.New("cannot parse threshold")
	errThresholdOperator = errors.New("threshold operator is not valid")
	errThresholdValue    = errors.New("threshold value not found")
	errAggregation       = errors.New("aggregation operation not found")
)

// Builds a simple alert condition.
//
// An example of a metric condition tree:
//
//	                       >=
//	            (threshold operator, id=0)
//	                      /  \
//	                    AVG    42
//	     (aggregation, id=1)   (threshold data, id=4)
//	         /    \
//	Metric Field   10m
//	    (id=2)    (duration, id=3)
//
// An example of an attribute condition tree:
//
//	                  >=
//	       (binary operator, id=0)
//	               /      \
//	   attribute.name     42
//	(attribute, id=1)     (constant, id=2)
func (model alertConditionModel) toAlertConditionInputs(ctx context.Context, diags *diag.Diagnostics, rootNodeId int) []swoClient.AlertConditionNodeInput {

	if !model.MetricName.IsNull() && !model.AttributeName.IsNull() {
		diags.AddError("Bad input in terraform resource",
			"Alerting condition must be either metric or attribute. Cannot populate both metric_name and attribute_name.")
		return []swoClient.AlertConditionNodeInput{}
	}

	if model.MetricName.IsNull() && model.AttributeName.IsNull() {
		diags.AddError("Bad input in terraform resource",
			"Alerting condition must be either metric or attribute. Must populate either metric_name or attribute_name.")
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
			Id:         rootNodeId,
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

		constantField := swoClient.AlertConditionNodeInput{
			Id:   rootNodeId + 2,
			Type: string(swoClient.AlertConstantValueType),
		}
		operator := model.AttributeOperator.ValueString()
		if operator == "IN" {
			var constantValues []string
			d := model.AttributeValues.ElementsAs(ctx, &constantValues, false)
			diags.Append(d...)
			if diags.HasError() {
				return []swoClient.AlertConditionNodeInput{}
			}
			// get the first value and determine its data type
			if len(constantValues) == 0 {
				diags.AddError("Bad input in terraform resource",
					"attribute_values is a required field when attribute_operator is 'IN'")
				return []swoClient.AlertConditionNodeInput{}
			}
			cv := constantValues[0]
			dataType := GetStringDataType(cv)
			constantField.DataType = &dataType
			constantField.Values = constantValues
		} else {
			dataType := GetStringDataType(model.AttributeValue.ValueString())
			constantField.DataType = &dataType
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
//     Ex:">=3000.123" -> operator '>=' and value '3000.123'
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
		match := alertThresholdRegex.FindStringSubmatch(threshold)
		if match == nil {
			return thresholdOperatorConditions, thresholdDataConditions, errThresholdParse
		}
		result := make(map[string]string)
		for i, name := range alertThresholdRegex.SubexpNames() {
			if i != 0 && name != "" {
				result[name] = match[i]
			}
		}

		operator := result["operator"]
		if !isValidThresholdOperator(operator) {
			return thresholdOperatorConditions, thresholdDataConditions, errThresholdOperator
		}

		thresholdOperatorConditions.Type = string(swoClient.AlertBinaryOperatorType)
		thresholdOperatorConditions.Operator = &operator

		thresholdValue := result["threshold"]
		if thresholdValue != "" {
			dataType := GetStringDataType(thresholdValue)

			thresholdDataConditions.Type = string(swoClient.AlertConstantValueType)
			thresholdDataConditions.DataType = &dataType
			thresholdDataConditions.Value = &thresholdValue
		} else {
			return thresholdOperatorConditions, thresholdDataConditions, errThresholdValue
		}
	}

	return thresholdOperatorConditions, thresholdDataConditions, nil
}

func isValidThresholdOperator(operator string) bool {
	validOperators := []swoClient.AlertBinaryOperator{
		swoClient.AlertOperatorEq,
		swoClient.AlertOperatorNe,
		swoClient.AlertOperatorLt,
		swoClient.AlertOperatorGt,
		swoClient.AlertOperatorLe,
		swoClient.AlertOperatorGe,
	}

	for _, op := range validOperators {
		if operator == string(op) {
			return true
		}
	}
	return false
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
		return swoClient.AlertConditionNodeInput{}, errAggregation
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
		tagFilter := tagToFilter(ctx, tag, diags)
		if diags.HasError() {
			return swoClient.AlertConditionNodeInput{}
		}
		metricFieldCondition.MetricFilter.Children = append(metricFieldCondition.MetricFilter.Children, tagFilter)
	}

	for _, tag := range excludeTags {
		tagFilter := tagToFilter(ctx, tag, diags)
		if diags.HasError() {
			return swoClient.AlertConditionNodeInput{}
		}
		metricFieldCondition.MetricFilter.Children =
			append(metricFieldCondition.MetricFilter.Children, swoClient.AlertFilterExpressionInput{
				Operation: swoClient.FilterOperationNot,
				Children:  []swoClient.AlertFilterExpressionInput{tagFilter},
			})
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

func tagToFilter(ctx context.Context, tag alertTagsModel, diags *diag.Diagnostics) swoClient.AlertFilterExpressionInput {
	var propertyValues []*string
	dFilter := tag.Values.ElementsAs(ctx, &propertyValues, false)
	diags.Append(dFilter...)
	if diags.HasError() {
		return swoClient.AlertFilterExpressionInput{}
	}

	var operation = swoClient.FilterOperationIn
	if !tag.Operation.IsNull() && !tag.Operation.IsUnknown() {
		// Operation is guaranteed to be valid due to schema validation.
		operation = swoClient.FilterOperation(tag.Operation.ValueString())
	}
	propertyName := tag.Name.ValueString()

	if operation == swoClient.FilterOperationContains {
		if len(propertyValues) != 1 {
			diags.AddError("Invalid Tag Filter",
				fmt.Sprintf("Exactly one value expected for 'CONTAINS'; found %d.", len(propertyValues)),
			)
			return swoClient.AlertFilterExpressionInput{}
		}

		return swoClient.AlertFilterExpressionInput{
			PropertyName:  &propertyName,
			Operation:     operation,
			PropertyValue: propertyValues[0],
		}
	}

	return swoClient.AlertFilterExpressionInput{
		PropertyName:   &propertyName,
		Operation:      operation,
		PropertyValues: propertyValues,
	}
}
