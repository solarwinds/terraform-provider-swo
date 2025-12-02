package stringmodifier

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

// UseNonNullStateForUnknown returns a plan modifier that copies a known prior state
// value into the planned value, as long as the prior state value is not null. This
// mimics the previous behavior in the Terraform Plugin Framework, updated in 1.15.1
// to copy the state even when it is null (that still qualifies as known). Despite
// the conceptual correctness, this created problems for providers that now failed
// on updates because of an explicit null in the plan.
func UseNonNullStateForUnknown() planmodifier.String {
	return useNonNullStateForUnknownModifier{}
}

// useNonNullStateForUnknownModifier implements the plan modifier.
type useNonNullStateForUnknownModifier struct{}

// Description returns a human-readable description of the plan modifier.
func (m useNonNullStateForUnknownModifier) Description(_ context.Context) string {
	return "Preserves previous non-null state value when planned value is unknown and config is known."
}

// MarkdownDescription returns a markdown description of the plan modifier.
func (m useNonNullStateForUnknownModifier) MarkdownDescription(_ context.Context) string {
	return "Preserves previous non-null state value when planned value is unknown and config is known."
}

// PlanModifyString implements the plan modification logic.
func (m useNonNullStateForUnknownModifier) PlanModifyString(
	_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse,
) {
	// Do nothing if there is no state value.
	if req.StateValue.IsNull() {
		return
	}

	// Do nothing if there is a known planned value.
	if !req.PlanValue.IsUnknown() {
		return
	}

	// Do nothing if there is an unknown configuration value, otherwise interpolation gets messed up.
	if req.ConfigValue.IsUnknown() {
		return
	}

	resp.PlanValue = req.StateValue
}
