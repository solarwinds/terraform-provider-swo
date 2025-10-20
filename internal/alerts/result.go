package alerts

import (
	"fmt"

	swoClient "github.com/solarwinds/swo-client-go/pkg/client"
	"github.com/solarwinds/terraform-provider-swo/internal/typex"
)

// Ensure that linkedResultNode implements linkedValueNode.
var _ linkedValueNode[*swoClient.ReadAlertConditionValueResult, string] = &linkedResultNode[*swoClient.ReadAlertConditionValueResult]{}

// linkedResultNode represents the extra structure that comes from the alerting API, used for
// linking nodes hierarchically in the received flattened representation. The operands slice
// contains the IDs of the children nodes for this node.
type linkedResultNode[T any] struct {
	id       string
	operands []string
	value    T
}

func (n *linkedResultNode[T]) GetId() string         { return n.id }
func (n *linkedResultNode[T]) GetOperands() []string { return n.operands }
func (n *linkedResultNode[T]) GetValue() T           { return n.value }

// ConditionsFromResult converts the flat list of alert condition nodes into a hierarchical
// representation, better suitable for further processing. It's worth noting that the returned
// value is built on top of the original structure, so any change to members of the result slice
// or its nested components will affect the returned Condition. This is done on purpose to
// avoid cloning a potentially big structure.
func ConditionsFromResult(result []swoClient.ReadAlertConditionResult) (Condition, error) {
	conditionNodes, err := typex.MapWithError(result,
		func(node swoClient.ReadAlertConditionResult) (
			linkedValueNode[*swoClient.ReadAlertConditionValueResult, string], error,
		) {
			if node.Value == nil {
				return nil, fmt.Errorf("%w: no value for alert condition node ID: %v", ErrBadCondition, node.Id)
			}
			var operands []string
			for _, link := range node.Links {
				if link.Name == "operands" {
					operands = link.Values
					break
				}
			}
			return &linkedResultNode[*swoClient.ReadAlertConditionValueResult]{
				id:       node.Id,
				operands: operands,
				value:    node.Value,
			}, nil
		})
	if err != nil {
		return nil, err
	}

	return fromFlatNodes(conditionNodes,
		func(value *swoClient.ReadAlertConditionValueResult, operands []Condition) (Condition, error) {
			metricFilter, err := getResultMetricFilter(value)
			if err != nil {
				return nil, err
			}
			if metricFilter != nil {
				// This compensates for a weird behavior in the alerting API. When a metrics
				// filter is provided, the API will echo this back not only as part of the
				// metric filter, but also as a search expression in the query attribute. We
				// zero it when a filter is present to avoid a false positive when checking
				// for differences. Fortunately, the API doesn't let setting both metric
				// and query, so if the former is present we know that the latter was just
				// made up. Otherwise, we honor the query because it could have been set
				// explicitly and we do want to learn about the difference. The copy avoids
				// overwriting the original response. (The hierarchical structure is built
				// on top of the original to avoid needless data copying.)
				clonedValue := *value
				clonedValue.Query = nil
				value = &clonedValue
			}
			return &condition{
				BaseCondition: value,
				entityFilter:  getResultEntityFilter(value),
				metricFilter:  metricFilter,
				operands:      operands,
			}, nil
		})
}

// getResultEntityFilter converts the entity filter from the API representation to the internal
// one. Note that the returned value shares storage with the provided filter, so changes to the
// latter affect the former. This method receives the whole ReadAlertConditionValueResult value
// because github.com/Khan/genqlient does not export the inner types.
func getResultEntityFilter(value *swoClient.ReadAlertConditionValueResult) EntityFilter {
	if value.EntityFilter == nil {
		return nil
	}
	fields := value.EntityFilter.Fields
	newFields := make([]ConditionMatchFieldRule, 0, len(fields))

	for _, field := range fields {
		rules := make([]ConditionMatchRule, 0, len(field.Rules))

		for _, rule := range field.Rules {
			rules = append(rules, &conditionMatchRule{
				Negate: rule.Negate,
				Type:   rule.Type,
				Value:  rule.Value,
			})
		}
		newFields = append(newFields, &conditionMatchFieldRule{
			fieldName: field.FieldName,
			rules:     rules,
		})
	}

	return &entityFilter{
		BaseEntityFilter: value.EntityFilter,
		fields:           newFields,
	}
}

// getResultMetricFilter converts the metric filter from the API representation to the internal
// one. Note that the returned value shares storage with the provided filter, so changes to the
// latter affect the former. This method receives the whole ReadAlertConditionValueResult value
// because github.com/Khan/genqlient does not export the inner types.
func getResultMetricFilter(node *swoClient.ReadAlertConditionValueResult) (MetricFilter, error) {
	flatExpressions := make([]linkedValueNode[BaseMetricFilter, string], 0, len(node.MetricFilter))

	// We cannot do this with typex.Map because the base type is not exported by genqlient.
	for _, filterNode := range node.MetricFilter {
		if filterNode.Value == nil {
			return nil, fmt.Errorf("%w: no value for filter node with ID: %v", ErrBadCondition, filterNode.Id)
		}
		var operands []string
		for _, link := range filterNode.Links {
			if link.Name == "children" {
				operands = link.Values
				break
			}
		}
		flatExpressions = append(flatExpressions, &linkedResultNode[BaseMetricFilter]{
			id:       filterNode.Id,
			operands: operands,
			value:    filterNode.Value,
		})
	}

	return fromFlatNodes(flatExpressions,
		func(value BaseMetricFilter, operands []MetricFilter) (MetricFilter, error) {
			return &metricFilter{
				BaseMetricFilter: value,
				operands:         operands,
			}, nil
		},
	)
}
