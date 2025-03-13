package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type compositeMetricResourceModel struct {
	Id          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	DisplayName types.String `tfsdk:"display_name"`
	Description types.String `tfsdk:"description"`
	Formula     types.String `tfsdk:"formula"`
	Unit        types.String `tfsdk:"unit"`
}

func (r *compositeMetricResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A terraform resource for managing composite metrics.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Description: "The metric name.",
				Required:    true,
			},
			"display_name": schema.StringAttribute{
				Description: "Display name of the composite metric. A short description of the metric.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the composite metric. A detailed description of the metric.",
				Required:    true,
			},
			"formula": schema.StringAttribute{
				Description: "PromQL query to calculate the composite metric. example: rate(system.disk.io[5m])",
				Required:    true,
			},
			"unit": schema.StringAttribute{
				Description: "Unit of the composite metric. example: bytes/s",
				Required:    true,
			},
		},
	}
}
