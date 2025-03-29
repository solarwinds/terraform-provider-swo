package provider

import (
	"context"
	"errors"
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
func (model alertConditionModel) toAlertConditionInputs(ctx context.Context, rootNodeId int) ([]swoClient.AlertConditionNodeInput, error) {

	thresholdOperatorCondition, thresholdDataCondition, err := model.toThresholdConditionInputs()
	if err != nil {
		return []swoClient.AlertConditionNodeInput{}, err
	}
	thresholdOperatorCondition.Id = rootNodeId
	thresholdOperatorCondition.OperandIds = []int{rootNodeId + 1, rootNodeId + 4}

	aggregationCondition, err := model.toAggregationConditionInput()
	if err != nil {
		return []swoClient.AlertConditionNodeInput{}, err
	}
	aggregationCondition.Id = rootNodeId + 1
	aggregationCondition.OperandIds = []int{rootNodeId + 2, rootNodeId + 3}

	metricFieldCondition := model.toMetricFieldConditionInput(ctx)
	metricFieldCondition.Id = rootNodeId + 2

	durationCondition := model.toDurationConditionInput()
	durationCondition.Id = rootNodeId + 3

	thresholdDataCondition.Id = rootNodeId + 4

	conditions := []swoClient.AlertConditionNodeInput{
		thresholdOperatorCondition,
		aggregationCondition,
		metricFieldCondition,
		durationCondition,
		thresholdDataCondition,
	}

	return conditions, nil
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

func (model alertConditionModel) toMetricFieldConditionInput(ctx context.Context) swoClient.AlertConditionNodeInput {

	var groupByMetricTag []string
	_ = model.GroupByMetricTag.ElementsAs(ctx, &groupByMetricTag, false)
	metricName := model.MetricName.ValueString()

	metricFieldCondition := swoClient.AlertConditionNodeInput{
		Type:             string(swoClient.AlertMetricFieldType),
		FieldName:        &metricName,
		GroupByMetricTag: groupByMetricTag,
	}

	var entityFilterTypes, entityFilterIds []string
	_ = model.TargetEntityTypes.ElementsAs(ctx, &entityFilterTypes, false)
	_ = model.EntityIds.ElementsAs(ctx, &entityFilterIds, false)

	if len(model.EntityIds.Elements()) > 0 {
		entityFilter := &swoClient.AlertConditionNodeEntityFilterInput{
			Types: entityFilterTypes,
			Ids:   entityFilterIds,
		}

		metricFieldCondition.EntityFilter = entityFilter
	}

	var includeTags []alertTagsModel
	var excludeTags []alertTagsModel

	if model.IncludeTags != nil {
		includeTags = *model.IncludeTags
	}

	if model.ExcludeTags != nil {
		excludeTags = *model.ExcludeTags
	}

	if len(includeTags) > 0 || len(excludeTags) > 0 {
		metricFieldCondition.MetricFilter = &swoClient.AlertFilterExpressionInput{
			Operation: swoClient.FilterOperationAnd,
		}
	}

	for _, tag := range includeTags {
		var metricFilterPropertyValues []*string
		_ = tag.Values.ElementsAs(ctx, &metricFilterPropertyValues, false)
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
		_ = tag.Values.ElementsAs(ctx, &propertyValues, false)
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

func GetStringDataType(s string) string {
	dataType := "string"

	if _, err := strconv.Atoi(s); err == nil {
		dataType = "number"
	}

	return dataType
}
