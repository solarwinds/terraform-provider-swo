package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	swoClient "github.com/solarwinds/swo-client-go/pkg/client"
	"github.com/solarwinds/terraform-provider-swo/internal/validators"
)

// The main Alert Resource model that is derived from the schema.
type alertResourceModel struct {
	Id                  types.String          `tfsdk:"id"`
	Name                types.String          `tfsdk:"name"`
	Description         types.String          `tfsdk:"description"`
	Severity            types.String          `tfsdk:"severity"`
	Enabled             types.Bool            `tfsdk:"enabled"`
	Conditions          []alertConditionModel `tfsdk:"conditions"`
	Notifications       []string              `tfsdk:"notifications"`
	TriggerResetActions types.Bool            `tfsdk:"trigger_reset_actions"`
}

type alertConditionModel struct {
	MetricName        types.String      `tfsdk:"metric_name"`
	Threshold         types.String      `tfsdk:"threshold"`
	Duration          types.String      `tfsdk:"duration"`
	AggregationType   types.String      `tfsdk:"aggregation_type"`
	EntityIds         []string          `tfsdk:"entity_ids"`
	TargetEntityTypes []string          `tfsdk:"target_entity_types"`
	IncludeTags       *[]alertTagsModel `tfsdk:"include_tags"`
	ExcludeTags       *[]alertTagsModel `tfsdk:"exclude_tags"`
}

type alertTagsModel struct {
	Name   types.String `tfsdk:"name"`
	Values []*string    `tfsdk:"values"`
}

func (r *alertResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A terraform resource for managing alerts.",
		Attributes: map[string]schema.Attribute{
			"id": resourceIdAttribute(),
			"name": schema.StringAttribute{
				Description: "Alert definition name.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Alert definition description.",
				Optional:    true,
			},
			"severity": schema.StringAttribute{
				Description: "Alert definition severity.",
				Required:    true,
				Validators: []validator.String{
					validators.SingleOption(
						swoClient.AlertSeverityInfo,
						swoClient.AlertSeverityWarning,
						swoClient.AlertSeverityCritical,
					),
				},
			},
			"enabled": schema.BoolAttribute{
				Description: "Enabled whether Alert definition shall be evaluated.",
				Computed:    true,
				Optional:    true,
				Default:     booldefault.StaticBool(true),
			},
			"trigger_reset_actions": schema.BoolAttribute{
				Description: "A flag indicating whether to send a notification when active alert returns to normal. It will be set to false if not specified.",
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
							Description: "The aggregation type that will be applyed to the metric and duration.",
							Required:    true,
							Validators: []validator.String{
								validators.SingleOption(
									swoClient.AlertOperatorAvg,
									swoClient.AlertOperatorCount,
									swoClient.AlertOperatorLast,
									swoClient.AlertOperatorMax,
									swoClient.AlertOperatorMin,
									swoClient.AlertOperatorSum,
								),
							},
						},
						"entity_ids": schema.ListAttribute{
							Description: "A list of Entity IDs that will be used to filter on by the alert.",
							Optional:    true,
							ElementType: types.StringType,
						},
						"target_entity_types": schema.ListAttribute{
							Description: "The entity types for scoping this alert (e.g. Website, Host, Database...).",
							Required:    true,
							ElementType: types.StringType,
						},
						"include_tags": schema.SetNestedAttribute{
							Description: "List of Metric values that the metric field will be in.",
							Optional:    true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"name": schema.StringAttribute{
										Description: "Name of the metric values the metric field will be in.",
										Optional:    true,
									},
									"values": schema.ListAttribute{
										Description: "Metric values the metric field will be in.",
										Optional:    true,
										ElementType: types.StringType,
									},
								},
							},
						},
						"exclude_tags": schema.SetNestedAttribute{
							Description: "List of Metric values that the metric field will not be in.",
							Optional:    true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"name": schema.StringAttribute{
										Description: "Name of the metric values the metric field will not be in.",
										Optional:    true,
									},
									"values": schema.ListAttribute{
										Description: "Metric values the metric field will not be in.",
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
