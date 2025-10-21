package alerts

import (
	swoClient "github.com/solarwinds/swo-client-go/pkg/client"
	"github.com/solarwinds/terraform-provider-swo/internal/typex"
)

type BaseEntityFilter interface {
	GetTypes() []string
	GetIds() []string
	GetQuery() *string
}

type EntityFilter interface {
	BaseEntityFilter
	GetFields() []ConditionMatchFieldRule
	Equals(EntityFilter) bool
}

type entityFilter struct {
	BaseEntityFilter
	fields []ConditionMatchFieldRule
}

func (f *entityFilter) GetFields() []ConditionMatchFieldRule { return f.fields }

func (f *entityFilter) Equals(other EntityFilter) bool {
	return typex.SliceEqual(f.GetTypes(), other.GetTypes()) &&
		typex.SliceEqual(f.GetIds(), other.GetIds()) &&
		typex.PtrEqual(f.GetQuery(), other.GetQuery()) &&
		typex.SliceEqualFunc(f.GetFields(), other.GetFields(), ConditionMatchFieldRule.Equals)
}

type ConditionMatchFieldRule interface {
	GetFieldName() string
	GetRules() []ConditionMatchRule
	Equals(ConditionMatchFieldRule) bool
}

type conditionMatchFieldRule struct {
	fieldName string
	rules     []ConditionMatchRule
}

func (f *conditionMatchFieldRule) GetFieldName() string           { return f.fieldName }
func (f *conditionMatchFieldRule) GetRules() []ConditionMatchRule { return f.rules }

func (f *conditionMatchFieldRule) Equals(other ConditionMatchFieldRule) bool {
	return f.GetFieldName() == other.GetFieldName() &&
		typex.SliceEqualFunc(f.GetRules(), other.GetRules(), ConditionMatchRule.Equals)
}

type BaseConditionMatchRule interface {
	GetType() swoClient.AlertConditionMatchRuleType
	GetNegate() bool
	GetValue() string
}

type ConditionMatchRule interface {
	BaseConditionMatchRule
	Equals(ConditionMatchRule) bool
}

type conditionMatchRule struct {
	Type   swoClient.AlertConditionMatchRuleType
	Negate bool
	Value  string
}

func (r *conditionMatchRule) GetType() swoClient.AlertConditionMatchRuleType { return r.Type }

func (r *conditionMatchRule) GetNegate() bool  { return r.Negate }
func (r *conditionMatchRule) GetValue() string { return r.Value }

func (r *conditionMatchRule) Equals(other ConditionMatchRule) bool {
	return conditionMatchRuleEquals(r, other)
}

func conditionMatchRuleEquals(a, b ConditionMatchRule) bool {
	return a.GetType() == b.GetType() &&
		a.GetNegate() == b.GetNegate() &&
		a.GetValue() == b.GetValue()
}

type BaseMetricFilter interface {
	GetOperation() swoClient.FilterOperation
	GetPropertyName() *string
	GetPropertyValue() *string
	GetPropertyValues() []*string
}

type MetricFilter interface {
	BaseMetricFilter
	GetOperands() []MetricFilter
	Equals(MetricFilter) bool
}

type metricFilter struct {
	BaseMetricFilter
	operands []MetricFilter
}

func (f *metricFilter) GetOperands() []MetricFilter { return f.operands }

func (f *metricFilter) Equals(other MetricFilter) bool {
	return f.GetOperation() == other.GetOperation() &&
		typex.PtrEqual(f.GetPropertyName(), other.GetPropertyName()) &&
		typex.PtrEqual(f.GetPropertyValue(), other.GetPropertyValue()) &&
		typex.SliceEqualFunc(f.GetPropertyValues(), other.GetPropertyValues(), typex.PtrEqual) &&
		typex.SliceEqualFunc(f.GetOperands(), other.GetOperands(), MetricFilter.Equals)
}
