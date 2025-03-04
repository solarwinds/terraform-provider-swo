package provider

import (
	"log"
	"reflect"
	"regexp"
	"strconv"

	swoClient "github.com/solarwinds/swo-client-go/pkg/client"
)

type conditionType string

const (
	conditionTypeThresholdData        conditionType = "thresholdData"
	conditionTypeDuration             conditionType = "duration"
	conditionTypeMetric               conditionType = "metric"
	conditionTypeAggregation          conditionType = "aggregation"
	conditionTypeThresholdOperator    conditionType = "thresholdOperator"
	conditionTypeNotReportingOperator conditionType = "notReportingOperator"
	conditionTypeNotReportingData     conditionType = "notReportingData"
)

type ConditionMap struct {
	condition     swoClient.AlertConditionNodeInput
	conditionType conditionType
}

func (model alertConditionModel) toAlertConditionInputs(conditions []swoClient.AlertConditionNodeInput) []swoClient.AlertConditionNodeInput {

	thresholdOperatorCondition, thresholdDataCondition := model.toThresholdConditionInputs()
	notReportingOperatorCondition, notReportingDataCondition := model.toNotReportingConditionInput()

	conditionMaps := []ConditionMap{
		{
			condition:     notReportingDataCondition,
			conditionType: conditionTypeNotReportingData,
		},
		{
			condition:     thresholdDataCondition,
			conditionType: conditionTypeThresholdData,
		},
		{
			condition:     model.toDurationConditionInput(),
			conditionType: conditionTypeDuration,
		},
		{
			condition:     model.toMetricFieldConditionInput(),
			conditionType: conditionTypeMetric,
		},
		{
			condition:     model.toAggregationConditionInput(),
			conditionType: conditionTypeAggregation,
		},
		{
			condition:     notReportingOperatorCondition,
			conditionType: conditionTypeNotReportingOperator,
		},
		{
			condition:     thresholdOperatorCondition,
			conditionType: conditionTypeThresholdOperator,
		},
	}

	var conditionsOrdered []swoClient.AlertConditionNodeInput
	conditionsReturnedLen := len(conditionMaps)
	lastId := len(conditions) + conditionsReturnedLen
	thresholdOperatorKey := conditionsReturnedLen - 1
	notReportingKey := conditionsReturnedLen - 2
	aggregationKey := conditionsReturnedLen - 3

	// Use the conditionsMaps to order our conditions, and assign the correct "OperandIds".
	for _, conditionMap := range conditionMaps {
		condition := conditionMap.condition

		// If the condition is empty don't add it.
		if reflect.DeepEqual(condition, swoClient.AlertConditionNodeInput{}) {
			continue
		}

		condition.Id = lastId
		conditionType := conditionMap.conditionType

		if conditionType == conditionTypeThresholdData {
			operandIds := append([]int{lastId}, conditionMaps[thresholdOperatorKey].condition.OperandIds...)
			conditionMaps[thresholdOperatorKey].condition.OperandIds = operandIds
		} else if conditionType == conditionTypeAggregation {
			// If threshold condition is nil don't update
			thresholdCondition := conditionMaps[thresholdOperatorKey].condition
			if !reflect.DeepEqual(thresholdCondition, swoClient.AlertConditionNodeInput{}) {
				operandIds := append([]int{lastId}, conditionMaps[thresholdOperatorKey].condition.OperandIds...)
				conditionMaps[thresholdOperatorKey].condition.OperandIds = operandIds
			}

			// If not_reporting condition is nil don't update
			notReportingCondition := conditionMaps[notReportingKey].condition
			if !reflect.DeepEqual(notReportingCondition, swoClient.AlertConditionNodeInput{}) {
				operandIds2 := append([]int{lastId}, conditionMaps[notReportingKey].condition.OperandIds...)
				conditionMaps[notReportingKey].condition.OperandIds = operandIds2
			}
		} else if conditionType == conditionTypeMetric || conditionType == conditionTypeDuration {
			operandIds := append([]int{lastId}, conditionMaps[aggregationKey].condition.OperandIds...)
			conditionMaps[aggregationKey].condition.OperandIds = operandIds
		} else if conditionType == conditionTypeNotReportingData {
			operandIds := append([]int{lastId}, conditionMaps[notReportingKey].condition.OperandIds...)
			conditionMaps[notReportingKey].condition.OperandIds = operandIds
		}

		lastId--
		conditionsOrdered = append([]swoClient.AlertConditionNodeInput{condition}, conditionsOrdered...)
	}

	return append(conditions, conditionsOrdered...)
}

func (model *alertConditionModel) toThresholdConditionInputs() (swoClient.AlertConditionNodeInput, swoClient.AlertConditionNodeInput) {
	threshold := model.Threshold.ValueString()
	thresholdOperatorConditions := swoClient.AlertConditionNodeInput{}
	thresholdDataConditions := swoClient.AlertConditionNodeInput{}

	if threshold != "" {

		regex := regexp.MustCompile(`[\W]+`)
		operator := regex.FindString(threshold)
		//Parses threshold into an operator:(>, <, = ...).

		operatorType, err := swoClient.GetAlertConditionType(operator)
		if err == nil {
			thresholdOperatorConditions.Type = operatorType
			thresholdOperatorConditions.Operator = &operator
		}
		regex = regexp.MustCompile("[0-9]+")
		thresholdValue := regex.FindString(threshold)
		//Parses threshold into numbers:(3000, 200, 10...).

		if thresholdValue != "" {
			dataType := GetStringDataType(thresholdValue)

			thresholdDataConditions.Type = string(swoClient.AlertConstantValueType)
			thresholdDataConditions.DataType = &dataType
			thresholdDataConditions.Value = &thresholdValue
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

func (model *alertConditionModel) toNotReportingConditionInput() (swoClient.AlertConditionNodeInput, swoClient.AlertConditionNodeInput) {
	notReportingOperatorCondition := swoClient.AlertConditionNodeInput{}
	notReportingDataCondition := swoClient.AlertConditionNodeInput{}

	if model.NotReporting.ValueBool() {
		operator := string(swoClient.AlertOperatorEq)
		notReportingOperatorCondition.Type = string(swoClient.AlertBinaryOperatorType)
		notReportingOperatorCondition.Operator = &operator

		dataValue := "0"
		dataType := GetStringDataType(dataValue)
		notReportingDataCondition.Type = string(swoClient.AlertConstantValueType)
		notReportingDataCondition.DataType = &dataType
		notReportingDataCondition.Value = &dataValue
	}

	return notReportingOperatorCondition, notReportingDataCondition
}

func (model *alertConditionModel) toAggregationConditionInput() swoClient.AlertConditionNodeInput {
	aggregationCondition := swoClient.AlertConditionNodeInput{}

	operator := model.AggregationType.ValueString()
	operatorType, err := swoClient.GetAlertConditionType(operator)
	if err != nil {
		log.Fatal("Threshold operation not found")
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
