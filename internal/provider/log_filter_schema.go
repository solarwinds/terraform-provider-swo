package provider

import (
	"context"

	swoClient "github.com/solarwinds/swo-client-go/pkg/client"
	"github.com/solarwinds/terraform-provider-swo/internal/validators"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// logFilterResourceModel is the main resource structure
type logFilterResourceModel struct {
	Id             types.String          `tfsdk:"id"`
	Name           types.String          `tfsdk:"name"`
	Description    types.String          `tfsdk:"description"`
	TokenSignature *string               `tfsdk:"token_signature"`
	Expressions    []logFilterExpression `tfsdk:"expressions"`
}

// LogFilterResourceOptions represents the options field in the main resource
type logFilterExpression struct {
	Kind       swoClient.ExclusionFilterExpressionKind `tfsdk:"kind"`
	Expression string                                  `tfsdk:"expression"`
}

func (r *logFilterResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A terraform resource for managing log exclusion filters.",
		Attributes: map[string]schema.Attribute{
			"id": resourceIdAttribute(),
			"name": schema.StringAttribute{
				Description: "The name of the log exclusion filter.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "The description of the log exclusion filter.",
				Optional:    true,
			},
			"token_signature": schema.StringAttribute{
				Description: "The ID of the ingestion token to scope the exclusion filter to. If not provided, the filter will be global. If provided, the filter will only apply to logs ingested by the specified token. (NOTE: There may be only one global filter.)",
				Optional:    true,
			},
			"expressions": schema.ListNestedAttribute{
				Description: "The list of exclusions for the log exclusion filter.",
				Required:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"kind": schema.StringAttribute{
							Description: "The kind of the log exclusion filter.",
							Required:    true,
							Validators: []validator.String{
								validators.SingleOption(
									swoClient.ExclusionFilterExpressionKindString,
									swoClient.ExclusionFilterExpressionKindRegex),
							},
						},
						"expression": schema.StringAttribute{
							Description: "The expression of the log exclusion filter.",
							Required:    true,
						},
					},
				},
			},
		},
	}
}
