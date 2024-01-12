package validators

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// singleOptionValidator validates that the value matches one of expected values.
type singleOptionValidator[T any] struct {
	ValidValues []T
}

func (v singleOptionValidator[T]) Description(ctx context.Context) string {
	return v.MarkdownDescription(ctx)
}

func (v singleOptionValidator[T]) MarkdownDescription(ctx context.Context) string {
	validValues := make([]string, len(v.ValidValues))
	for i, value := range v.ValidValues {
		validValues[i] = getStringValue(value)
	}
	return fmt.Sprintf("value must be one of: %q", strings.Join(validValues, ", "))
}

// SingleOption checks that the value specified in the attribute is one of the ValidValues.
func SingleOption[T any](validValues ...T) singleOptionValidator[T] {
	return singleOptionValidator[T]{
		ValidValues: validValues,
	}
}

func (v singleOptionValidator[T]) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue

	for _, valid := range v.ValidValues {
		if strings.Compare(value.ValueString(), getStringValue(valid)) == 0 {
			return
		}
	}

	resp.Diagnostics.Append(validatordiag.InvalidAttributeValueMatchDiagnostic(
		req.Path,
		v.Description(ctx),
		value.String(),
	))
}

func getStringValue[T any](input T) string {
	return fmt.Sprintf("%v", input)
}
