package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	swoClient "github.com/solarwinds/swo-client-go/pkg/client"
	"github.com/solarwinds/terraform-provider-swo/internal/envvar"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// LogFilterResourceModel is the main resource structure
type LogFilterResourceModel struct {
	Id             types.String          `tfsdk:"id"`
	Name           types.String          `tfsdk:"name"`
	Description    types.String          `tfsdk:"description"`
	TokenSignature *string               `tfsdk:"token_signature"`
	Expressions    []LogFilterExpression `tfsdk:"expressions"`
}

// LogFilterResourceOptions represents the options field in the main resource
type LogFilterExpression struct {
	Kind       swoClient.ExclusionFilterExpressionKind `tfsdk:"kind"`
	Expression string                                  `tfsdk:"expression"`
}

func (r *LogFilterResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	tflog.Trace(ctx, "LogFilterResource: Schema")

	resp.Schema = schema.Schema{
		Description: fmt.Sprintf("A terraform resource for managing %s LogFilters.", envvar.AppName),
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "This is a computed ID provided by the backend.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the log exclusion filter.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "The description of the log exclusion filter.",
				Optional:    true,
			},
			"token_signature": schema.StringAttribute{
				Description: "The token signature of the log exclusion filter.",
				Optional:    true,
			},
			"expressions": schema.SetNestedAttribute{
				Description: "The list of exclusions for the log exclusion filter.",
				Required:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"kind": schema.StringAttribute{
							Description: "The kind of the log exclusion filter [STRING|REGEX].",
							Required:    true,
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
