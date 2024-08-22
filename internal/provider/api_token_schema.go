package provider

import (
	"context"

	"github.com/solarwinds/terraform-provider-swo/internal/validators"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	swoClient "github.com/solarwinds/swo-client-go/pkg/client"
)

// apiTokenResourceModel is the main resource model.
type apiTokenResourceModel struct {
	Id          types.String                `tfsdk:"id"`
	Name        types.String                `tfsdk:"name"`
	Enabled     types.Bool                  `tfsdk:"enabled"`
	Type        types.String                `tfsdk:"type"`
	Token       types.String                `tfsdk:"token"`
	AccessLevel *swoClient.TokenAccessLevel `tfsdk:"access_level"`
	Attributes  []apiTokenAttribute         `tfsdk:"attributes"`
}

// apiTokenAttribute is a custom attribute for the ApiTokenResourceModel.
type apiTokenAttribute struct {
	Key   types.String `tfsdk:"key"`
	Value types.String `tfsdk:"value"`
}

func (r *apiTokenResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A terraform resource for managing API tokens.",
		Attributes: map[string]schema.Attribute{
			"id": resourceIdAttribute(),
			"name": schema.StringAttribute{
				Description: "The user provided name of the token.",
				Required:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "True if the token is enabled.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"type": schema.StringAttribute{
				Description: "The type of token.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("public-api"),
			},
			"token": schema.StringAttribute{
				Description: "The plain-text value of the token.",
				Computed:    true,
				Sensitive:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"access_level": schema.StringAttribute{
				Description: "The access level of the token.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(string(swoClient.TokenAccessLevelFull)),
				Validators: []validator.String{
					validators.SingleOption(
						swoClient.TokenAccessLevelFull,
						swoClient.TokenAccessLevelRead,
						swoClient.TokenAccessLevelRecord),
				},
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
