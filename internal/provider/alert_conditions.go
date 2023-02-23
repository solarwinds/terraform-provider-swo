package provider

import (
	"log"
	"regexp"

	swoClient "github.com/solarwindscloud/terraform-provider-swo/internal/client"
)

func (model AlertConditionModel) ToAlertConditionInputs(conditions []swoClient.AlertConditionNodeInput) []swoClient.AlertConditionNodeInput {
	thresholdOperatorCondition, thresholdDataCondition := model.ToThresholdConditionInputs()
	durationCondition := model.ToDurationConditionInput()
	aggregationCondition := model.ToAggregationConditionInput()
	metricCondition := model.ToMetricFieldConditionInput()

	conditionExists := map[string]bool{
		"thresholdData":      *thresholdDataCondition.GetValue() != "",
		"duration":           *durationCondition.GetValue() != "",
		"metric":             *metricCondition.GetFieldName() != "",
		"aggregation":        *aggregationCondition.GetOperator() != "",
		"thresholdOperation": *thresholdOperatorCondition.GetOperator() != "",
	}

	conditionsOrderd := []swoClient.AlertConditionNodeInput{}
	lastId := len(conditions)

	for _, exists := range conditionExists {
		if exists {
			lastId++
		}
	}

	if conditionExists["thresholdData"] {
		thresholdDataCondition.Id = lastId
		thresholdOperatorCondition.OperandIds = append(thresholdOperatorCondition.OperandIds, lastId)
		lastId--
		conditionsOrderd = PrependConditions(thresholdDataCondition, conditionsOrderd)
	}

	if conditionExists["duration"] {
		durationCondition.Id = lastId
		aggregationCondition.OperandIds = append([]int{lastId}, aggregationCondition.OperandIds...)
		lastId--
		conditionsOrderd = PrependConditions(durationCondition, conditionsOrderd)
	}

	if conditionExists["metric"] {
		metricCondition.Id = lastId
		aggregationCondition.OperandIds = append([]int{lastId}, aggregationCondition.OperandIds...)
		lastId--
		conditionsOrderd = PrependConditions(metricCondition, conditionsOrderd)
	}

	if conditionExists["aggregation"] {
		aggregationCondition.Id = lastId
		thresholdOperatorCondition.OperandIds = append([]int{lastId}, thresholdOperatorCondition.OperandIds...)
		lastId--
		conditionsOrderd = PrependConditions(aggregationCondition, conditionsOrderd)
	}

	if conditionExists["thresholdOperation"] {
		thresholdOperatorCondition.Id = lastId
		conditionsOrderd = PrependConditions(thresholdOperatorCondition, conditionsOrderd)
	}

	return append(conditions, conditionsOrderd...)
}

func (model *AlertConditionModel) ToThresholdConditionInputs() (swoClient.AlertConditionNodeInput, swoClient.AlertConditionNodeInput) {
	threshold := Trim(model.Threshold.String())
	thresholdOperatorConditions := swoClient.AlertConditionNodeInput{}
	thresholdDataConditions := swoClient.AlertConditionNodeInput{}

	if threshold != "" {
		regex := regexp.MustCompile(`[\W]+`)
		operator := string(regex.FindString(threshold))

		operatorType, err := swoClient.GetAlertConditionType(operator)
		if err == nil {
			thresholdOperatorConditions.Type = operatorType
			thresholdOperatorConditions.Operator = &operator
		}

		regex = regexp.MustCompile("[0-9]+")
		thresholdValue := string(regex.FindString(threshold))

		if thresholdValue != "" {
			dataType := GetDataType(thresholdValue)

			thresholdDataConditions.Type = string(swoClient.AlertConstantValueType)
			thresholdDataConditions.DataType = &dataType
			thresholdDataConditions.Value = &thresholdValue
		}
	}

	return thresholdOperatorConditions, thresholdDataConditions
}

func (model *AlertConditionModel) ToDurationConditionInput() swoClient.AlertConditionNodeInput {
	durationCondition := swoClient.AlertConditionNodeInput{}

	duration := Trim(model.Duration.String())
	dataType := GetDataType(duration)

	if duration != "" {
		durationCondition.Type = string(swoClient.AlertConstantValueType)
		durationCondition.DataType = &dataType
		durationCondition.Value = &duration
	}

	return durationCondition
}

func (model *AlertConditionModel) ToAggregationConditionInput() swoClient.AlertConditionNodeInput {
	aggragationCondition := swoClient.AlertConditionNodeInput{}

	operator := Trim(model.AggregationType.String())
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

func (model *AlertConditionModel) ToMetricFieldConditionInput() swoClient.AlertConditionNodeInput {
	metricFieldCondition := swoClient.AlertConditionNodeInput{}
	metricName := Trim(model.MetricName.String())

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
			propertyName := Trim(tag.Name.String())
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
			propertyName := Trim(tag.Name.String())
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

func PrependConditions(condition swoClient.AlertConditionNodeInput, conditions []swoClient.AlertConditionNodeInput) []swoClient.AlertConditionNodeInput {
	return append([]swoClient.AlertConditionNodeInput{condition}, conditions...)
}
