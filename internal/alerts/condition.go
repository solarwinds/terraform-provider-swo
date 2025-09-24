package alerts

import (
	"github.com/solarwinds/terraform-provider-swo/internal/typex"
)

type BaseCondition interface {
	GetType() string
	GetOperator() *string
	GetFieldName() *string
	GetDataType() *string
	GetValue() *string
	GetValues() []string
	GetQuery() *string
	GetNamespace() *string
	GetGroupByMetricTag() []string
}

// Condition abstracts alert conditions, to the fullest extent possible in the alerts API.
// Note that not all possible conditions can be expressed in Terraform.
type Condition interface {
	BaseCondition
	GetEntityFilter() EntityFilter
	GetMetricFilter() MetricFilter
	GetOperands() []Condition
	Equals(Condition) bool
}

type condition struct {
	BaseCondition
	entityFilter EntityFilter
	metricFilter MetricFilter
	operands     []Condition
}

func (cond *condition) GetEntityFilter() EntityFilter { return cond.entityFilter }
func (cond *condition) GetMetricFilter() MetricFilter { return cond.metricFilter }
func (cond *condition) GetOperands() []Condition      { return cond.operands }

func (cond *condition) Equals(other Condition) bool {
	return cond.GetType() == other.GetType() &&
		typex.PtrEqual(cond.GetOperator(), other.GetOperator()) &&
		typex.PtrEqual(cond.GetFieldName(), other.GetFieldName()) &&
		typex.PtrEqual(cond.GetDataType(), other.GetDataType()) &&
		typex.PtrEqual(cond.GetValue(), other.GetValue()) &&
		typex.SliceEqual(cond.GetValues(), other.GetValues()) &&
		typex.PtrEqual(cond.GetQuery(), other.GetQuery()) &&
		typex.PtrEqual(cond.GetNamespace(), other.GetNamespace()) &&
		typex.SliceEqual(cond.GetGroupByMetricTag(), other.GetGroupByMetricTag()) &&
		typex.RefCompare(cond.GetEntityFilter(), other.GetEntityFilter(), EntityFilter.Equals) &&
		typex.RefCompare(cond.GetMetricFilter(), other.GetMetricFilter(), MetricFilter.Equals) &&
		typex.SliceEqualFunc(cond.GetOperands(), other.GetOperands(), Condition.Equals)
}
