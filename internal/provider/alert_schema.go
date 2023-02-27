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
	Enabled             types.Bool            `tfsdk:"enabled"`
	Conditions          []AlertConditionModel `tfsdk:"conditions"`
	Notifications       []int                 `tfsdk:"notifications"`
	NotificationType    types.String          `tfsdk:"notification_type"`
	TriggerResetActions types.Bool            `tfsdk:"trigger_reset_actions"`
}

type AlertConditionModel struct {
	MetricName        types.String      `tfsdk:"metric_name"`
	Threshold         types.String      `tfsdk:"threshold"`
	Duration          types.String      `tfsdk:"duration"`
	AggregationType   types.String      `tfsdk:"aggregation_type"`
	EntityIds         []string          `tfsdk:"entity_ids"`
	TargetEntityTypes []string          `tfsdk:"target_entity_types"`
	IncludeTags       *[]AlertTagsModel `tfsdk:"include_tags"`
	ExcludeTags       *[]AlertTagsModel `tfsdk:"exclude_tags"`
}

type AlertTagsModel struct {
	Name   types.String `tfsdk:"name"`
	Values []*string    `tfsdk:"values"`
}

func (r *AlertResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	tflog.Trace(ctx, "AlertResource: Schema")

	resp.Schema = schema.Schema{
		Description: fmt.Sprintf("A terraform resource for managing %s alerts.", envvar.AppName),
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the alert. This is a computed value provided by the backend.",
				Computed:    true,
				Optional:    true,
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
				Description: "The type of alert (ENTITY_METRICS|LOGS).",
				Required:    true,
			},
			"severity": schema.StringAttribute{
				Description: "The severity of the alert (INFO|WARNING|CRITICAL).",
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
							Description: "The name of the metric that is being monitored.",
							Required:    true,
						},
						"threshold": schema.StringAttribute{
							Description: "The threshold value that triggers the alert when breached.",
							Required:    true,
						},
						"duration": schema.StringAttribute{
							Description: "The duration of the time that the ",
							Required:    true,
						},
						"aggregation_type": schema.StringAttribute{
							Description: "The aggregation type (such as average, maximum or minimum) to apply to this metric.",
							Required:    true,
						},
						"entity_ids": schema.ListAttribute{
							Description: "(Optional) A list of entity IDs that will be the scoped targets of the monitoring.",
							Optional:    true,
							ElementType: types.StringType,
						},
						"target_entity_types": schema.ListAttribute{
							Description: "The entity types for scoping this alert (e.g. Website, Host, Database...).",
							Required:    true,
							ElementType: types.StringType,
						},
						"include_tags": schema.SetNestedAttribute{
							Description: "(Optional) Add metric tags to include as part of the scope.",
							Optional:    true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"name": schema.StringAttribute{
										Description: "The tag name.",
										Optional:    true,
									},
									"values": schema.ListAttribute{
										Description: "One or more tag values.",
										Optional:    true,
										ElementType: types.StringType,
									},
								},
							},
						},
						"exclude_tags": schema.SetNestedAttribute{
							Description: "(Optional) Add metric tags to exclude as part of the scope.",
							Optional:    true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"name": schema.StringAttribute{
										Description: "The tag name.",
										Optional:    true,
									},
									"values": schema.ListAttribute{
										Description: "One or more tag values.",
										Optional:    true,
										ElementType: types.StringType,
									},
								},
							},
						},
					},
				},
			},
			"notification_type": schema.StringAttribute{
				Description: "The type of notification service this alert will notify.",
				Required:    true,
			},
			"notifications": schema.ListAttribute{
				Description: "A list of notification configuration IDs to publish to when this alert is triggered.",
				Required:    true,
				ElementType: types.NumberType,
			},
		},
	}
}
