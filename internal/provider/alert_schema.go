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
	Enabled             types.Bool            `tfsdk:"enabled"`
	Conditions          []AlertConditionModel `tfsdk:"conditions"`
	Notifications       []string              `tfsdk:"notifications"`
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
				Description: "Alert definition ID in UUID format. This is a computed value provided by the backend when an alert is created. (Optional)",
				Computed:    true,
				Optional:    true,
			},
			"name": schema.StringAttribute{
				Description: "Alert definition name.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Alert definition description. (Optional)",
				Optional:    true,
			},
			"severity": schema.StringAttribute{
				Description: "Alert definition severity (INFO|WARNING|CRITICAL).",
				Required:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Enabled whether Alert definition shall be evaluated. Default is true. (Optional)",
				Optional:    true,
			},
			"trigger_reset_actions": schema.BoolAttribute{
				Description: "A flag indicating whether to send a notification when active alert returns to normal. It will be set to false if not specified. (Optional)",
				Optional:    true,
			},
			"conditions": schema.SetNestedAttribute{
				Description: "One or more conditions that must be met to tigger the alert.",
				Required:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"metric_name": schema.StringAttribute{
							Description: "The field name of the metric to be filtered on.",
							Required:    true,
						},
						"threshold": schema.StringAttribute{
							Description: "Operator and value that represents the threshold of an the alert. When the threshold is breached it triggers the alert. For Opertator - binaryOperator:(=|!=|>|<|>=|<=), logicalOperator:(AND|OR)",
							Required:    true,
						},
						"duration": schema.StringAttribute{
							Description: "Duration of time that will be used to check if the threshold has been breached.",
							Required:    true,
						},
						"aggregation_type": schema.StringAttribute{
							Description: "The aggregation type that will be applyed to the metric and duration. (MIN|MAX|AVG|SUM|LAST)",
							Required:    true,
						},
						"entity_ids": schema.ListAttribute{
							Description: "A list of Entity IDs that will be used to filter on by the alert. (Optional)",
							Optional:    true,
							ElementType: types.StringType,
						},
						"target_entity_types": schema.ListAttribute{
							Description: "The entity types for scoping this alert (e.g. Website, Host, Database...).",
							Required:    true,
							ElementType: types.StringType,
						},
						"include_tags": schema.SetNestedAttribute{
							Description: "List of Metric values that the metric field will be in. (Optional)",
							Optional:    true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"name": schema.StringAttribute{
										Description: "Name of the metric values the metric field will be in. (Optional)",
										Optional:    true,
									},
									"values": schema.ListAttribute{
										Description: "Metric values the metric field will be in. (Optional)",
										Optional:    true,
										ElementType: types.StringType,
									},
								},
							},
						},
						"exclude_tags": schema.SetNestedAttribute{
							Description: "List of Metric values that the metric field will not be in. (Optional)",
							Optional:    true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"name": schema.StringAttribute{
										Description: "Name of the metric values the metric field will not be in. (Optional)",
										Optional:    true,
									},
									"values": schema.ListAttribute{
										Description: "Metric values the metric field will not be in. (Optional)",
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
				Description: "A list of notifications to assign to this alert.",
				Required:    true,
				ElementType: types.StringType,
			},
		},
	}
}
