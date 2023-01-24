package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/solarwindscloud/terraform-provider-swo/internal/envvar"
)

// The main Alert Resource model that is derived from the schema.
type AlertResourceModel struct {
	ID                  types.String          `tfsdk:"id"`
	Name                types.String          `tfsdk:"name"`
	Description         types.String          `tfsdk:"description"`
	Severity            types.String          `tfsdk:"severity"`
	Type                types.String          `tfsdk:"type"`
	EntityType          types.String          `tfsdk:"entity_type"`
	Enabled             types.Bool            `tfsdk:"enabled"`
	Conditions          []AlertConditionModel `tfsdk:"conditions"`
	Notifications       types.List            `tfsdk:"notifications"`
	TriggerResetActions types.Bool            `tfsdk:"trigger_reset_actions"`
}

type AlertConditionModel struct {
	MetricName      types.String      `tfsdk:"metric_name"`
	Threshold       types.String      `tfsdk:"threshold"`
	Duration        types.String      `tfsdk:"duration"`
	AggregationType types.String      `tfsdk:"aggregation_type"`
	EntityIds       types.List        `tfsdk:"entity_ids"`
	IncludeTags     *[]AlertTagsModel `tfsdk:"include_tags"`
	ExcludeTags     *[]AlertTagsModel `tfsdk:"exclude_tags"`
}

type AlertTagsModel struct {
	Name   types.String `tfsdk:"name"`
	Values types.List   `tfsdk:"values"`
}

func (r *AlertResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	tflog.Trace(ctx, "AlertResource: Schema")

	resp.Schema = schema.Schema{
		Description: fmt.Sprintf("A terraform resource for managing %s alerts.", envvar.AppName),
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the alert. This is computed from the backend.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the alert.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "A short description of the alert (optional).",
				Optional:    true,
			},
			"type": schema.StringAttribute{
				Description: "The type of alert (METRICS|LOGS).",
				Required:    true,
			},
			"severity": schema.StringAttribute{
				Description: "The severity of the alert (INFO|WARNING|CRITICAL).",
				Required:    true,
			},
			"entity_type": schema.StringAttribute{
				Description: "???",
				Required:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Is the alert enabled. Default is true.",
				Optional:    true,
			},
			"trigger_reset_actions": schema.BoolAttribute{
				Description: "???",
				Optional:    true,
			},
			"conditions": schema.SetNestedAttribute{
				Description: "One or more conditions that must be met to tigger the alert.",
				Required:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"metric_name": schema.StringAttribute{
							Description: "???",
							Required:    true,
						},
						"threshold": schema.StringAttribute{
							Description: "???",
							Required:    true,
						},
						"duration": schema.StringAttribute{
							Description: "???",
							Required:    true,
						},
						"aggregation_type": schema.StringAttribute{
							Description: "???",
							Required:    true,
						},
						"entity_ids": schema.ListAttribute{
							Description: "???",
							Optional:    true,
							ElementType: types.StringType,
						},
						"include_tags": schema.SetNestedAttribute{
							Description: "???",
							Optional:    true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"name": schema.StringAttribute{
										Description: "???",
										Optional:    true,
									},
									"values": schema.ListAttribute{
										Description: "???",
										Optional:    true,
										ElementType: types.StringType,
									},
								},
							},
						},
						"exclude_tags": schema.SetNestedAttribute{
							Description: "???",
							Optional:    true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"name": schema.StringAttribute{
										Description: "???",
										Optional:    true,
									},
									"values": schema.ListAttribute{
										Description: "???",
										Optional:    true,
										ElementType: types.StringType,
									},
								},
							},
						},
					},
				},
			},
			"notifications": schema.ListAttribute{
				Description: "???",
				Optional:    true,
				ElementType: types.NumberType,
			},
		},
	}
}
