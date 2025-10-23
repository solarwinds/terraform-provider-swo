package validators

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// mockValidator is a simple validator for testing purposes
type mockValidator struct {
	description string
	shouldFail  bool
	errorMsg    string
}

func (m mockValidator) Description(_ context.Context) string {
	return m.description
}

func (m mockValidator) MarkdownDescription(_ context.Context) string {
	return m.description
}

func (m mockValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if m.shouldFail {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Validation Error",
			m.errorMsg,
		)
	}
}

func TestWhenValidator_Description(t *testing.T) {
	innerValidator := mockValidator{description: "must be valid"}
	condMsg := "attribute X is set"
	cond := func(ctx context.Context, req validator.StringRequest, diags *diag.Diagnostics) bool {
		return true
	}

	v := When(cond, condMsg, innerValidator)

	desc := v.Description(context.Background())
	expected := "when attribute X is set, must be valid"
	if desc != expected {
		t.Errorf("Expected description '%s', got '%s'", expected, desc)
	}
}

func TestWhenValidator_MarkdownDescription(t *testing.T) {
	innerValidator := mockValidator{description: "must be valid"}
	condMsg := "attribute X is set"
	cond := func(ctx context.Context, req validator.StringRequest, diags *diag.Diagnostics) bool {
		return true
	}

	v := When(cond, condMsg, innerValidator)

	desc := v.MarkdownDescription(context.Background())
	expected := "when attribute X is set, must be valid"
	if desc != expected {
		t.Errorf("Expected markdown description '%s', got '%s'", expected, desc)
	}
}

func TestWhenValidator_ValidateString(t *testing.T) {
	condErrorMsg := "condition evaluation failed"
	innerErrorMsg := "inner validation failed"

	tests := []struct {
		name                string
		value               types.String
		conditionResult     bool
		conditionFails      bool
		innerValidatorFails bool
		expectError         bool
		expectedErrorMsg    string
	}{
		{
			name:                "null value skips validation",
			value:               types.StringNull(),
			conditionResult:     true,
			innerValidatorFails: true,
		},
		{
			name:                "unknown value skips validation",
			value:               types.StringUnknown(),
			conditionResult:     true,
			innerValidatorFails: true,
		},
		{
			name:                "condition returns true, inner validator passes",
			value:               types.StringValue("test"),
			conditionResult:     true,
			innerValidatorFails: false,
		},
		{
			name:                "condition returns true, inner validator fails",
			value:               types.StringValue("test"),
			conditionResult:     true,
			innerValidatorFails: true,
			expectError:         true,
			expectedErrorMsg:    innerErrorMsg,
		},
		{
			name:                "condition returns false, inner validator is skipped",
			value:               types.StringValue("test"),
			conditionResult:     false,
			innerValidatorFails: true,
		},
		{
			name:                "condition has error, inner validator is skipped",
			value:               types.StringValue("test"),
			conditionResult:     false,
			conditionFails:      true,
			innerValidatorFails: true,
			expectError:         true,
			expectedErrorMsg:    condErrorMsg,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			innerValidator := mockValidator{
				description: "must be valid",
				shouldFail:  test.innerValidatorFails,
				errorMsg:    innerErrorMsg,
			}

			cond := func(ctx context.Context, req validator.StringRequest, diags *diag.Diagnostics) bool {
				if test.conditionFails {
					diags.AddError("Condition Error", condErrorMsg)
				}
				return test.conditionResult
			}

			v := When(cond, "test condition", innerValidator)

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
