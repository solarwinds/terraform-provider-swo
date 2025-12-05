package alerts

import (
	swoClient "github.com/solarwinds/swo-client-go/pkg/client"
	"github.com/solarwinds/terraform-provider-swo/internal/typex"
)

var _ linkedValueNode[*swoClient.AlertConditionNodeInput, int] = &linkedInputNode{}

// linkedInputNode is used to represent the links composing a node hierarchy in the
// flattened version of condition nodes. It's implemented directly on top of the input
// node and, thanks to embedding, it automatically forwards the call to get the ID to
// the underlying input node.
type linkedInputNode struct {
	*swoClient.AlertConditionNodeInput
}

func (n *linkedInputNode) GetOperands() []int { return n.GetOperandIds() }

func (n *linkedInputNode) GetValue() *swoClient.AlertConditionNodeInput {
	return n.AlertConditionNodeInput
}

// conditionMatchRuleInput is used as a tiny wrapper, so that the type can include the
// Equals operation. If this was not required, we wouldn't even have to create this
// type, as the input structure provided by genqlient already has all the required
// members.
type conditionMatchRuleInput struct {
	BaseConditionMatchRule
}

func (r *conditionMatchRuleInput) Equals(other ConditionMatchRule) bool {
	return conditionMatchRuleEquals(r, other)
}

// ConditionsFromInput converts the list representation of an alert condition slice into
// the hierarchical representation. It's worth noting that the returned value is built on
// top of the original structure, so any change to members of the result slice or its
// nested components will affect the returned Condition. This is done on purpose to avoid
// cloning a potentially big structure.
func ConditionsFromInput(result []swoClient.AlertConditionNodeInput) (Condition, error) {
	conditionNodes := typex.Map(result,
		func(node swoClient.AlertConditionNodeInput) linkedValueNode[*swoClient.AlertConditionNodeInput, int] {
			return &linkedInputNode{&node}
		})

	return fromFlatNodes(conditionNodes,
		func(value *swoClient.AlertConditionNodeInput, operands []Condition) (Condition, error) {
			result := &condition{
				BaseCondition: value,
				operands:      operands,
			}
			if value.EntityFilter != nil {
				result.entityFilter = getInputEntityFilter(value.EntityFilter)
			}
			if value.MetricFilter != nil {
				result.metricFilter = getInputMetricFilter(value.MetricFilter)
			}
			return result, nil
		})
}

// getInputEntityFilter converts the (non-nil) entity filter from an input structure to
// the internal one. Note that the returned value shares storage with the provided filter,
// so changes to the latter affect the former.
func getInputEntityFilter(value *swoClient.AlertConditionNodeEntityFilterInput) EntityFilter {
	fields := typex.Map(value.Fields,
		func(field swoClient.AlertConditionMatchFieldRuleInput) ConditionMatchFieldRule {
			return &conditionMatchFieldRule{
				fieldName: field.FieldName,
				rules: typex.Map(field.Rules,
					func(rule swoClient.AlertConditionMatchRuleInput) ConditionMatchRule {
						return &conditionMatchRuleInput{&rule}
					}),
			}
		})

	return &entityFilter{
		BaseEntityFilter: value,
		fields:           fields,
	}
}

// getInputMetricFilter converts the (non-nil) metric filter from an input structure to
// the internal one. Note that the returned value shares storage with the provided filter,
// so changes to the latter affect the former.
func getInputMetricFilter(value *swoClient.AlertFilterExpressionInput) MetricFilter {
	return &metricFilter{
		BaseMetricFilter: value,
		operands: typex.Map(value.Children, func(child swoClient.AlertFilterExpressionInput) MetricFilter {
			return getInputMetricFilter(&child)
		}),
	}
}
