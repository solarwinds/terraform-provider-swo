package alerts

import (
	"fmt"

	"github.com/solarwinds/terraform-provider-swo/internal/typex"
)

// linkedValueNode abstracts the flattened representation of a hierarchical structure using
// nodes with IDs that have a primary value and a list of children nodes as operands. Those
// operands are also referenced by their IDs.
type linkedValueNode[T any, U comparable] interface {
	GetId() U
	GetOperands() []U
	GetValue() T
}

type nodeCreator[T, U any] func(value T, operands []U) (U, error)

// fromFlatNodes converts a flat representation of a hierarchical structure into the actual
// hierarchy, rooted at flatNodes[0]. (If flatNodes is empty, the method returns immediately
// with the zero for the type.) The flat representation is defined by a slice of structures
// with a node ID, list of children (also given by ID) and the single value node that will end
// up as part of the hierarchy. The createNode function is used to create each node of the
// resulting structure, given the current node value and all already-resolved children. This
// method will fail if an unknown ID is present, circular dependencies are found, or if the
// createNode function returns an error.
func fromFlatNodes[T, U any, V comparable](flatNodes []linkedValueNode[T, V], createNode nodeCreator[T, U]) (U, error) {
	var result U
	if len(flatNodes) == 0 {
		return result, nil
	}

	flatNodesById := make(map[V]linkedValueNode[T, V])
	for _, flatNode := range flatNodes {
		flatNodesById[flatNode.GetId()] = flatNode
	}
	usedIDs := make(map[V]bool)
	return fromFlatNodeStep(flatNodes[0], flatNodesById, createNode, usedIDs)
}

// fromFlatNodeStep implements the recursive step for fromFlatNodes.
func fromFlatNodeStep[T, U any, V comparable](
	currentNode linkedValueNode[T, V],
	flatNodesByID map[V]linkedValueNode[T, V],
	createNode nodeCreator[T, U],
	usedIDs map[V]bool,
) (U, error) {
	operands, err := typex.MapWithError(currentNode.GetOperands(), func(operandID V) (U, error) {
		if usedIDs[operandID] {
			// Cycle detected. This operand had already been used.
			return typex.Zero[U](),
				fmt.Errorf("%w: cycle in alert condition for operand ID: %v", ErrBadCondition, operandID)
		}
		usedIDs[operandID] = true

		operandNode, ok := flatNodesByID[operandID]
		if !ok {
			// Missing reference. There is no operand by this ID.
			return typex.Zero[U](), fmt.Errorf("%w: unknown alert condition operand ID: %v", ErrBadCondition, operandID)
		}
		return fromFlatNodeStep[T, U, V](operandNode, flatNodesByID, createNode, usedIDs)
	})
	if err != nil {
		return typex.Zero[U](), err
	}
	return createNode(currentNode.GetValue(), operands)
}
