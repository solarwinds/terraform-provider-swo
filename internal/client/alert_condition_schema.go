package client

import "errors"

type AlertOperatorType string

const (
	AlertBinaryOperatorType      AlertOperatorType = "binaryOperator"
	AlertLogicalOperatorType     AlertOperatorType = "logicalOperator"
	AlertUnaryOperatorType       AlertOperatorType = "unaryOperator"
	AlertConstantValueType       AlertOperatorType = "constantValue"
	AlertAggregationOperatorType AlertOperatorType = "aggregationOperator"
	AlertMetricFieldType         AlertOperatorType = "metricField"
	AlertQueryFieldType          AlertOperatorType = "queryField"
)

type AlertAggregationOperator string

const (
	AlertOperatorCount AlertAggregationOperator = "COUNT"
	AlertOperatorMin   AlertAggregationOperator = "MIN"
	AlertOperatorMax   AlertAggregationOperator = "MAX"
	AlertOperatorAvg   AlertAggregationOperator = "AVG"
	AlertOperatorSum   AlertAggregationOperator = "SUM"
	AlertOperatorLast  AlertAggregationOperator = "LAST"
)

type AlertBinaryOperator string

const (
	AlertOperatorEq AlertBinaryOperator = "="
	AlertOperatorNe AlertBinaryOperator = "!="
	AlertOperatorGt AlertBinaryOperator = ">"
	AlertOperatorLt AlertBinaryOperator = "<"
	AlertOperatorGe AlertBinaryOperator = ">="
	AlertOperatorLe AlertBinaryOperator = "<="
)

type AlertLogicalOperator string

const (
	AlertOperatorAnd AlertLogicalOperator = "AND"
	AlertOperatorOr  AlertLogicalOperator = "OR"
)

type AlertUnaryOperator string

const (
	AlertOperatorNot AlertUnaryOperator = "!"
)

var (
	AlertOperators = map[AlertOperatorType][]string{
		AlertAggregationOperatorType: {
			string(AlertOperatorCount),
			string(AlertOperatorMin),
			string(AlertOperatorMax),
			string(AlertOperatorAvg),
			string(AlertOperatorSum),
			string(AlertOperatorLast),
		},
		AlertBinaryOperatorType: {
			string(AlertOperatorEq),
			string(AlertOperatorNe),
			string(AlertOperatorGt),
			string(AlertOperatorLt),
			string(AlertOperatorGe),
			string(AlertOperatorLe),
		},
		AlertLogicalOperatorType: {
			string(AlertOperatorAnd),
			string(AlertOperatorOr),
		},
		AlertUnaryOperatorType: {
			string(AlertOperatorNot),
		},
	}

	FilterOperations = map[string]FilterOperation{
		string(AlertOperatorEq): FilterOperationEq,
		string(AlertOperatorNe): FilterOperationNe,
		string(AlertOperatorGt): FilterOperationGt,
		string(AlertOperatorGe): FilterOperationGe,
		string(AlertOperatorLt): FilterOperationLt,
		string(AlertOperatorLe): FilterOperationLe,
	}
)

func GetAlertConditionType(operator string) (string, error) {
	for operatorType, operatorArray := range AlertOperators {
		exists := SliceValueExists(operatorArray, operator)
		if exists {
			return string(operatorType), nil
		}
	}

	return "", errors.New("alert operation not supported")
}
