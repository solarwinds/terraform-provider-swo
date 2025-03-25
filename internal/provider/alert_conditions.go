package provider

import (
	"log"
	"regexp"
	"strconv"

	swoClient "github.com/solarwinds/swo-client-go/pkg/client"
)

type conditionType string

const (
	conditionTypeThresholdData     conditionType = "thresholdData"
	conditionTypeDuration          conditionType = "duration"
	conditionTypeMetric            conditionType = "metric"
	conditionTypeAggregation       conditionType = "aggregation"
	conditionTypeThresholdOperator conditionType = "thresholdOperator"
)

type ConditionMap struct {
	condition     swoClient.AlertConditionNodeInput
	conditionType conditionType
}

// Builds a simple metric condition.
//
// An example of a simple metric condition tree:
//
//	  							>=
//					(threshold operator, id=0)
//	       						/  \
//	       		 			AVG  	42
//			(aggregation, id=1)   (threshold data, id=4)
//	     			/  	\
//		Metric Field    10m
//		(id=2) 		   (duration, id=3)
func (model alertConditionModel) toAlertConditionInputs(conditions []swoClient.AlertConditionNodeInput) []swoClient.AlertConditionNodeInput {

	rootNode := 0
	// todo  possible reuse for multi conditions
	//conditionsReturnedLen := len(conditionMaps)
	//lastId := len(conditions) + conditionsReturnedLen
	thresholdOperatorCondition, thresholdDataCondition := model.toThresholdConditionInputs()
	thresholdOperatorCondition.Id = rootNode
	thresholdOperatorCondition.OperandIds = []int{rootNode + 1, rootNode + 4}

	aggregationCondition := model.toAggregationConditionInput()
	aggregationCondition.Id = rootNode + 1
	aggregationCondition.OperandIds = []int{rootNode + 2, rootNode + 3}

	metricFieldCondition := model.toMetricFieldConditionInput()
	metricFieldCondition.Id = rootNode + 2

	durationCondition := model.toDurationConditionInput()
	durationCondition.Id = rootNode + 3

	thresholdDataCondition.Id = rootNode + 4

	conditionsOrdered := []swoClient.AlertConditionNodeInput{
		thresholdOperatorCondition,
		aggregationCondition,
		metricFieldCondition,
		durationCondition,
		thresholdDataCondition,
	}
	// todo keep append to master list we can re-use for multi conditions...
	return append(conditions, conditionsOrdered...)
}

// Creates the threshold operation and threshold data nodes by either:
//  1. If not_reporting=true, operator is set to '=' and value to '0'
//  2. Else, parse the model.threshold string into operator and value
//     Ex:">=3000" -> operator '>=' and value '3000'
func (model *alertConditionModel) toThresholdConditionInputs() (swoClient.AlertConditionNodeInput, swoClient.AlertConditionNodeInput) {
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

	} else if threshold != "" {

		regex := regexp.MustCompile(`[\W]+`)
		operator := regex.FindString(threshold)
		//Parses threshold into an operator:(>, <, = ...).

		operatorType, err := swoClient.GetAlertConditionType(operator)
		if err != nil {
			log.Fatal("Threshold operation not found")
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
			log.Fatal("Threshold value not found")
		}
	}

	return thresholdOperatorConditions, thresholdDataConditions
}

func (model *alertConditionModel) toDurationConditionInput() swoClient.AlertConditionNodeInput {
	durationCondition := swoClient.AlertConditionNodeInput{}

	duration := model.Duration.ValueString()
	dataType := GetStringDataType(duration)

	if duration != "" {
		durationCondition.Type = string(swoClient.AlertConstantValueType)
		durationCondition.DataType = &dataType
		durationCondition.Value = &duration
	}

	return durationCondition
}

func (model *alertConditionModel) toAggregationConditionInput() swoClient.AlertConditionNodeInput {
	aggregationCondition := swoClient.AlertConditionNodeInput{}

	operator := model.AggregationType.ValueString()
	operatorType, err := swoClient.GetAlertConditionType(operator)
	if err != nil {
		log.Fatal("Aggregation operation not found")
	}

	if operator != "" {
		aggregationCondition.Type = operatorType
		aggregationCondition.Operator = &operator
	}

	return aggregationCondition
}

func (model *alertConditionModel) toMetricFieldConditionInput() swoClient.AlertConditionNodeInput {
	metricFieldCondition := swoClient.AlertConditionNodeInput{}
	metricName := model.MetricName.ValueString()

	if metricName != "" {
		metricFieldCondition = swoClient.AlertConditionNodeInput{
			Type:             string(swoClient.AlertMetricFieldType),
			FieldName:        &metricName,
			GroupByMetricTag: model.GroupByMetricTag,
		}

		if len(model.EntityIds) > 0 {
			entityFilter := &swoClient.AlertConditionNodeEntityFilterInput{
				Types: model.TargetEntityTypes,
				Ids:   model.EntityIds,
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
			propertyName := tag.Name.ValueString()
			metricFilter := swoClient.AlertFilterExpressionInput{
				PropertyName:   &propertyName,
				Operation:      swoClient.FilterOperationIn,
				PropertyValues: tag.Values,
			}

			metricFieldCondition.MetricFilter.Children = append(
				metricFieldCondition.MetricFilter.Children,
				metricFilter,
			)
		}

		for _, tag := range excludeTags {
			propertyName := tag.Name.ValueString()
			metricFilter := swoClient.AlertFilterExpressionInput{
				PropertyName:   &propertyName,
				Operation:      swoClient.FilterOperationIn,
				PropertyValues: tag.Values,
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
