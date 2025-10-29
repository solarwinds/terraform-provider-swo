package validators

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestNullValidator_Description(t *testing.T) {
	v := Null()
	ctx := context.Background()

	desc := v.Description(ctx)
	if desc != "value must be null" {
		t.Errorf("Expected description 'value must be null', got '%s'", desc)
	}
}

func TestNullValidator_MarkdownDescription(t *testing.T) {
	v := Null()
	ctx := context.Background()

	desc := v.MarkdownDescription(ctx)
	if desc != "value must be null" {
		t.Errorf("Expected markdown description 'value must be null', got '%s'", desc)
	}
}

func TestNullValidator_ValidateString(t *testing.T) {
	tests := []struct {
		name             string
		value            types.String
		expectError      bool
		expectedErrorMsg string
	}{
		{
			name:  "null value passes validation",
			value: types.StringNull(),
		},
		{
			name:  "unknown value passes validation",
			value: types.StringUnknown(),
		},
		{
			name:             "non-null value fails validation",
			value:            types.StringValue("test"),
			expectError:      true,
			expectedErrorMsg: "value must be null",
		},
		{
			name:             "empty string fails validation",
			value:            types.StringValue(""),
			expectError:      true,
			expectedErrorMsg: "value must be null",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			v := Null()

			req := validator.StringRequest{
				Path:           path.Root("test"),
				PathExpression: path.MatchRoot("test"),
				ConfigValue:    test.value,
			}
			resp := &validator.StringResponse{}

			v.ValidateString(context.Background(), req, resp)

			if !test.expectError {
				if resp.Diagnostics.HasError() {
					t.Errorf("Expected no validation error, but got: %v", resp.Diagnostics.Errors())
				}
				return
			}

			if !resp.Diagnostics.HasError() {
				t.Error("Expected validation error, but got none")
				return
			}
			diagErrors := resp.Diagnostics.Errors()
			if len(diagErrors) == 0 {
				t.Error("unexpected lack of error diagnostics when HasError returned true")
				return
			}

			found := false
			for _, d := range diagErrors {
				if strings.Contains(d.Summary(), test.expectedErrorMsg) || strings.Contains(d.Detail(), test.expectedErrorMsg) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected errors to contain '%s', but got: %q", test.expectedErrorMsg, diagErrors)
			}
		})
	}
}
