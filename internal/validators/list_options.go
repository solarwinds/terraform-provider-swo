package validators

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// listOptionsValidator validates that a list of values match one of expected values.
type listOptionsValidator[T any] struct {
	ValidValues     []T
	CaseInsensitive bool
	SplitString     bool
}

func (v listOptionsValidator[T]) Description(ctx context.Context) string {
	return v.MarkdownDescription(ctx)
}

func (v listOptionsValidator[T]) MarkdownDescription(ctx context.Context) string {
	validValues := make([]string, len(v.ValidValues))
	for i, value := range v.ValidValues {
		validValues[i] = getStringValue(value)
	}
	return fmt.Sprintf("value must be one of: %q", strings.Join(validValues, ", "))
}

// ListOption checks that the list of values specified in the attribute is one of the ValidValues.
func ListOptions[T any](validValues ...T) listOptionsValidator[T] {
	return listOptionsValidator[T]{
		ValidValues:     validValues,
		CaseInsensitive: true,
		SplitString:     true,
	}
}

func (v listOptionsValidator[T]) ValidateList(ctx context.Context, req validator.ListRequest, resp *validator.ListResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	var reqValues []string

	diag := req.ConfigValue.ElementsAs(ctx, &reqValues, false)
	if diag.HasError() {
		return
	}

	allValidValues := false
	for _, reqValue := range reqValues {
		floatVal, StringVal, err := SplitStringByDelimiter(reqValue, ":")

		if !isValidFloat(floatVal) || err != nil {
			resp.Diagnostics.Append(validatordiag.InvalidAttributeValueMatchDiagnostic(
				req.Path,
				v.Description(ctx),
				reqValue,
			))
		}

		for _, validValue := range v.ValidValues {
			validStr := fmt.Sprint(validValue)

			if strings.EqualFold(StringVal, validStr) {
				allValidValues = true
				break
			} else {
				allValidValues = false
			}
		}
		if !allValidValues {
			resp.Diagnostics.Append(validatordiag.InvalidAttributeValueMatchDiagnostic(
				req.Path,
				v.Description(ctx),
				reqValue,
			))
		}
	}

	if allValidValues {
		return
	} else {
		resp.Diagnostics.Append(validatordiag.InvalidAttributeValueMatchDiagnostic(
			req.Path,
			v.Description(ctx),
			strings.Join(reqValues, ","),
		))
	}
}

func SplitStringByDelimiter(stringToSplit string, delimiter string) (string, string, error) {
	parts := strings.Split(stringToSplit, delimiter)
	if len(parts) != 2 {
		return "", "", errors.New("invalid ID format")
	}
	return parts[0], parts[1], nil
}

func isValidFloat(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}
