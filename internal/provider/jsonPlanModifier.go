package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// UseStandardizedJson returns a plan modifier that standardizes the serialization
// format of the json string data into the planned value. Use this when json is
// stored as the value of a string attribute to prevent unwanted change detection
// due to added white space, field ordering, or other string value differences that
// aren't different when the json is unmarshaled into an object.
func useStandarizedJson() planmodifier.String {
	return standarizeJson{}
}

// standarizeJson implements the plan modifier.
type standarizeJson struct{}

// Description returns a human-readable description of the plan modifier.
func (m standarizeJson) Description(_ context.Context) string {
	return "Serializes JSON values in a consistent way."
}

// MarkdownDescription returns a markdown description of the plan modifier.
func (m standarizeJson) MarkdownDescription(_ context.Context) string {
	return "Serializes JSON values in a consistent way."
}

// PlanModifyString implements the plan modification logic.
func (m standarizeJson) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// First we unmarshal the plan value json to a standard object.
	var v any
	err := json.Unmarshal([]byte(req.PlanValue.ValueString()), &v)
	if err != nil {
		resp.Diagnostics.AddError("swo provider error", fmt.Sprintf("StandardizeJson plan modifier error: %s",
			err))
		return
	}

	// Now marshal the object back to a json string which we will use as the modified plan value for consistency.
	data, err := json.Marshal(&v)
	if err != nil {
		resp.Diagnostics.AddError("swo provider error", fmt.Sprintf("StandardizeJson plan modifier error: %s",
			err))
		return
	}

	resp.PlanValue = types.StringValue(string(data))
}
