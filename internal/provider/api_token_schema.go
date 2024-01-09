package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/solarwinds/terraform-provider-swo/internal/envvar"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	swoClient "github.com/solarwinds/swo-client-go/pkg/client"
)

// ApiTokenResourceModel is the main resource model.
type ApiTokenResourceModel struct {
	Id          types.String                `tfsdk:"id"`
	Name        types.String                `tfsdk:"name"`
	Enabled     types.Bool                  `tfsdk:"enabled"`
	Type        types.String                `tfsdk:"type"`
	Secure      types.Bool                  `tfsdk:"secure"`
	Token       types.String                `tfsdk:"token"`
	AccessLevel *swoClient.TokenAccessLevel `tfsdk:"access_level"`
	Attributes  []ApiTokenAttribute         `tfsdk:"attributes"`
}

// ApiTokenAttribute is a custom attribute for the ApiTokenResourceModel.
type ApiTokenAttribute struct {
	Key   types.String `tfsdk:"key"`
	Value types.String `tfsdk:"value"`
}

func (r *ApiTokenResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	tflog.Trace(ctx, "ApiTokenResource: Schema")

	resp.Schema = schema.Schema{
		Description: fmt.Sprintf("A terraform resource for managing %s ApiTokens.", envvar.AppName),
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "This is a computed ID provided by the backend.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The user provided name of the token.",
				Required:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "The enabled state of the token.",
				Required:    true,
			},
			"type": schema.StringAttribute{
				Description: "The type of the token.",
				Required:    true,
			},
			"secure": schema.BoolAttribute{
				Description: "The secure state of the token. Secure tokens are only revealed to the user once at creation and cannot be unobfuscated.",
				Computed:    true,
			},
			"token": schema.StringAttribute{
				Description: "The plain-text value of the token.",
				Computed:    true,
			},
			"access_level": schema.StringAttribute{
				Description: "The access level of the token [READ|RECORD|FULL].",
				Required:    true,
			},
			"attributes": schema.SetNestedAttribute{
				Description: "The custom attributes assigned to the token.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"key": schema.StringAttribute{
							Description: "The custom attribute key.",
							Required:    true,
						},
						"value": schema.StringAttribute{
							Description: "The custom attribute value.",
							Required:    true,
						},
					},
				},
			},
		},
	}
}
