package provider

import (
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func stringPtr(val types.String) *string {
	if val.IsNull() {
		return nil
	}

	result := val.ValueString()
	return &result
}

func boolPtr(val types.Bool) *bool {
	if val.IsNull() {
		return nil
	}

	result := val.ValueBool()
	return &result
}

func convertObject[T any](from any) (*T, error) {
	b, err := json.Marshal(&from)
	if err != nil {
		return nil, err
	}

	var result T
	err = json.Unmarshal(b, &result)
	if err != nil {
		return nil, err
	}

	return &result, err
}
