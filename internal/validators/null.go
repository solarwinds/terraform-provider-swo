package validators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

type nullValidator struct{}

// Null creates a validator that checks that the attribute it applies to is null.
// Validators are not called when attributes are omitted from the config, but they
// are when they are explicitly set, including when set to null.
func Null() validator.String {
	return nullValidator{}
}

func (s nullValidator) Description(ctx context.Context) string {
	return s.MarkdownDescription(ctx)
}

func (s nullValidator) MarkdownDescription(_ context.Context) string {
	return "value must be null"
}

func (s nullValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if !req.ConfigValue.IsNull() && !req.ConfigValue.IsUnknown() {
		resp.Diagnostics.Append(validatordiag.InvalidAttributeValueMatchDiagnostic(
			req.Path,
			s.Description(ctx),
			req.ConfigValue.ValueString(),
		))
	}
}
