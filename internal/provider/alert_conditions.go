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

func (model AlertConditionModel) toAlertConditionInputs(conditions []swoClient.AlertConditionNodeInput) []swoClient.AlertConditionNodeInput {
	thresholdOperatorCondition, thresholdDataCondition := model.toThresholdConditionInputs()

	conditionMaps := []ConditionMap{
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
			condition:     thresholdOperatorCondition,
			conditionType: conditionTypeThresholdOperator,
		},
	}

	conditionsOrderd := []swoClient.AlertConditionNodeInput{}
	conditionsReturnedLen := len(conditionMaps)
	lastId := len(conditions) + conditionsReturnedLen

	//Use the conditionsMaps to order our conditions, and assign the correct "OperandIds".
	for _, conditionMap := range conditionMaps {
		condition := conditionMap.condition

		//If the condition is empty dont add it.
		if reflect.DeepEqual(condition, swoClient.AlertConditionNodeInput{}) {
			continue
		}

		condition.Id = lastId
		conditionType := conditionMap.conditionType

		if conditionType == conditionTypeThresholdData || conditionType == conditionTypeAggregation {
			thresholdOperatorKey := conditionsReturnedLen - 1
			operandIds := append([]int{lastId}, conditionMaps[thresholdOperatorKey].condition.OperandIds...)
			conditionMaps[thresholdOperatorKey].condition.OperandIds = operandIds
		}

		if conditionType == conditionTypeMetric || conditionType == conditionTypeDuration {
			aggregationKey := conditionsReturnedLen - 2
			operandIds := append([]int{lastId}, conditionMaps[aggregationKey].condition.OperandIds...)
			conditionMaps[aggregationKey].condition.OperandIds = operandIds
		}

		lastId--
		conditionsOrderd = append([]swoClient.AlertConditionNodeInput{condition}, conditionsOrderd...)
	}

	return append(conditions, conditionsOrderd...)
}

func (model *AlertConditionModel) toThresholdConditionInputs() (swoClient.AlertConditionNodeInput, swoClient.AlertConditionNodeInput) {
	threshold := model.Threshold.ValueString()
	thresholdOperatorConditions := swoClient.AlertConditionNodeInput{}
	thresholdDataConditions := swoClient.AlertConditionNodeInput{}

	if threshold != "" {

		regex := regexp.MustCompile(`[\W]+`)
		operator := string(regex.FindString(threshold))
		//Parses threshold into an operator:(>, <, = ...).

		operatorType, err := swoClient.GetAlertConditionType(operator)
		if err == nil {
			thresholdOperatorConditions.Type = operatorType
			thresholdOperatorConditions.Operator = &operator
		}
		regex = regexp.MustCompile("[0-9]+")
		thresholdValue := string(regex.FindString(threshold))
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

func (model *AlertConditionModel) toDurationConditionInput() swoClient.AlertConditionNodeInput {
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

func (model *AlertConditionModel) toAggregationConditionInput() swoClient.AlertConditionNodeInput {
	aggragationCondition := swoClient.AlertConditionNodeInput{}

	operator := model.AggregationType.ValueString()
	operatorType, err := swoClient.GetAlertConditionType(operator)
	if err != nil {
		log.Fatal("Threshold operation not found")
	}

	if operator != "" {
		aggragationCondition.Type = operatorType
		aggragationCondition.Operator = &operator
	}

	return aggragationCondition
}

func (model *AlertConditionModel) toMetricFieldConditionInput() swoClient.AlertConditionNodeInput {
	metricFieldCondition := swoClient.AlertConditionNodeInput{}
	metricName := model.MetricName.ValueString()

	if metricName != "" {
		entityFilter := swoClient.AlertConditionNodeEntityFilterInput{
			Types: model.TargetEntityTypes,
			Ids:   model.EntityIds,
		}

		metricFieldCondition = swoClient.AlertConditionNodeInput{
			Type:         string(swoClient.AlertMetricFieldType),
			FieldName:    &metricName,
			EntityFilter: &entityFilter,
			MetricFilter: &swoClient.AlertFilterExpressionInput{
				Operation: swoClient.FilterOperationAnd,
			},
		}

		includeTags := *model.IncludeTags

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

		excludeTags := *model.ExcludeTags

		for _, tag := range excludeTags {
			propertyName := tag.Name.ValueString()
			metricFilter := swoClient.AlertFilterExpressionInput{
				PropertyName:   &propertyName,
				Operation:      swoClient.FilterOperationNe,
				PropertyValues: tag.Values,
			}

			metricFieldCondition.MetricFilter.Children = append(
				metricFieldCondition.MetricFilter.Children,
				metricFilter,
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
