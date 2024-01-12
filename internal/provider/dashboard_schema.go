package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/solarwinds/terraform-provider-swo/internal/validators"
)

// The main Dashboard Resource model that is derived from the schema.
type dashboardResourceModel struct {
	Id         types.String           `tfsdk:"id"`
	Name       types.String           `tfsdk:"name"`
	IsPrivate  types.Bool             `tfsdk:"is_private"`
	CategoryId types.String           `tfsdk:"category_id"`
	Widgets    []dashboardWidgetModel `tfsdk:"widgets"`
}

type dashboardWidgetModel struct {
	Id         types.String `tfsdk:"id"`
	Type       types.String `tfsdk:"type"`
	X          types.Int64  `tfsdk:"x"`
	Y          types.Int64  `tfsdk:"y"`
	Width      types.Int64  `tfsdk:"width"`
	Height     types.Int64  `tfsdk:"height"`
	Properties types.String `tfsdk:"properties"`
}

func (r *dashboardResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A terraform resource for managing dashboards.",
		Attributes: map[string]schema.Attribute{
			"id": resourceIdAttribute(),
			"name": schema.StringAttribute{
				Description: "The name of the dashboard.",
				Required:    true,
			},
			"is_private": schema.BoolAttribute{
				Description: "Is this dashboard restricted to the owner?",
				Optional:    true,
			},
			"category_id": schema.StringAttribute{
				Description: "The category that this dashboard is assigned to.",
				Optional:    true,
			},
			"widgets": schema.SetNestedAttribute{
				Description: "The widgets that are placed on the dashboard.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The computed id of the widget.",
							Computed:    true,
						},
						"type": schema.StringAttribute{
							Description: "The type of the widget.",
							Required:    true,
							Validators: []validator.String{
								validators.SingleOption("Kpi", "Proportional", "TimeSeries"),
							},
						},
						"x": schema.Int64Attribute{
							Description: "The X position of the widget.",
							Required:    true,
						},
						"y": schema.Int64Attribute{
							Description: "The Y position of the widget.",
							Required:    true,
						},
						"width": schema.Int64Attribute{
							Description: "The width of the widget.",
							Required:    true,
						},
						"height": schema.Int64Attribute{
							Description: "The height of the widget.",
							Required:    true,
						},
						"properties": schema.StringAttribute{
							Description: "A JSON encoded string that defines the widget configuration.",
							Required:    true,
							PlanModifiers: []planmodifier.String{
								useStandarizedJson(),
							},
						},
					},
				},
			},
		},
	}
}
