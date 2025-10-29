package validators

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// WhenCond represents a condition function that determines whether the inner
// validator should be applied.
type WhenCond = func(context.Context, validator.StringRequest, *diag.Diagnostics) bool

type whenValidator struct {
	cond    WhenCond
	condMsg string
	inner   validator.String
}

// When creates a conditional validator, that applies the inner validator only
// when the result of the condition function is true. If the condition function
// returns false, then the inner validator is completely ignored. The condMsg is
// used to describe the condition in the validator description, and it should be
// spelled as a phrase describing what the condition checks for. It should start
// with lowercase, be suitable for following the word "when", and not end with
// punctuation. For example: "attribute X is even".
func When(cond WhenCond, condMsg string, inner validator.String) validator.String {
	return whenValidator{
		cond:    cond,
		condMsg: condMsg,
		inner:   inner,
	}
}

func (s whenValidator) Description(ctx context.Context) string {
	return s.MarkdownDescription(ctx)
}

func (s whenValidator) MarkdownDescription(ctx context.Context) string {
	return fmt.Sprintf("when %s, %s", s.condMsg, s.inner.MarkdownDescription(ctx))
}

func (s whenValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	validate := s.cond(ctx, req, &resp.Diagnostics)
	if resp.Diagnostics.HasError() || !validate {
		return
	}
	s.inner.ValidateString(ctx, req, resp)
}
