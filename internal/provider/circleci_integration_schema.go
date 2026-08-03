package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/solarwinds/terraform-provider-swo/internal/planmodifier/stringmodifier"
)

const defaultReceiverBase = "https://webhook.swo.cloud.solarwinds.com/webhook"

// circleCIIntegrationResourceModel is the main resource model.
type circleCIIntegrationResourceModel struct {
	Id           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	ApiToken     types.String `tfsdk:"api_token"`
	ApiTokenName types.String `tfsdk:"api_token_name"`
	ReceiverBase types.String `tfsdk:"receiver_base"`
	ReceiverUrl  types.String `tfsdk:"receiver_url"`
	SecretToken  types.String `tfsdk:"secret_token"`
}

func (r *circleCIIntegrationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A terraform resource for managing CircleCI integrations.",
		Attributes: map[string]schema.Attribute{
			"id": resourceIdAttribute(),
			"name": schema.StringAttribute{
				Description: "The name of the CircleCI integration.",
				Required:    true,
			},
			"api_token": schema.StringAttribute{
				Description: "The CircleCI API token. Used to fetch project metadata and logs.",
				Optional:    true,
				Sensitive:   true,
			},
			"api_token_name": schema.StringAttribute{
				Description: "A label for the stored CircleCI API token. Required when api_token is set.",
				Optional:    true,
			},
			"receiver_base": schema.StringAttribute{
				Description: "The receiver URL base. Defaults to the SWO production endpoint.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(defaultReceiverBase),
			},
			"receiver_url": schema.StringAttribute{
				Description: "The receiver URL to configure in CircleCI project webhook.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringmodifier.UseNonNullStateForUnknown(),
				},
			},
			"secret_token": schema.StringAttribute{
				Description: "The secret token to configure in CircleCI project webhook and used to authenticate incoming webhooks.",
				Computed:    true,
				Sensitive:   true,
				PlanModifiers: []planmodifier.String{
					stringmodifier.UseNonNullStateForUnknown(),
				},
			},
		},
	}
}
