package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"

	swoClient "github.com/solarwinds/swo-client-go/pkg/client"
	"github.com/solarwinds/terraform-provider-swo/internal/validators"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// logFilterResourceModel is the main resource structure
type logFilterResourceModel struct {
	Id             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	TokenSignature types.String `tfsdk:"token_signature"`
	Expressions    types.List   `tfsdk:"expressions"`
}

// LogFilterResourceOptions represents the options field in the main resource
type logFilterExpression struct {
	Kind       types.String `tfsdk:"kind"`
	Expression types.String `tfsdk:"expression"`
}

func ExpressionAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"kind":       types.StringType,
		"expression": types.StringType,
	}
}

func (r *logFilterResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
							Description: "The kind of the log exclusion filter. Valid values are [`STRING`|`REGEX`].",
							Required:    true,
							Validators: []validator.String{
								validators.OneOf(
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
