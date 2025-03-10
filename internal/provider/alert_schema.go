package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
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
	Id                  types.String            `tfsdk:"id"`
	Name                types.String            `tfsdk:"name"`
	NotificationActions []alertActionInputModel `tfsdk:"notification_actions"`
	Description         types.String            `tfsdk:"description"`
	Severity            types.String            `tfsdk:"severity"`
	Enabled             types.Bool              `tfsdk:"enabled"`
	Conditions          []alertConditionModel   `tfsdk:"conditions"`
	Notifications       []string                `tfsdk:"notifications"`
	TriggerResetActions types.Bool              `tfsdk:"trigger_reset_actions"`
	RunbookLink         types.String            `tfsdk:"runbook_link"`
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
	GroupByMetricTag  []string          `tfsdk:"group_by_metric_tag"`
	NotReporting      types.Bool        `tfsdk:"not_reporting"`
}

type alertTagsModel struct {
	Name   types.String `tfsdk:"name"`
	Values []*string    `tfsdk:"values"`
}

type alertActionInputModel struct {
	Type                  types.String `tfsdk:"type"`
	ConfigurationIds      []string     `tfsdk:"configuration_ids"`
	ResendIntervalSeconds types.Int64  `tfsdk:"resend_interval_seconds"`
}

var notificationActionTypes = []string{
	"email",
	"amazonsns",
	"msTeams",
	"newRelic",
	"opsgenie",
	"pagerduty",
	"pushover",
	"serviceNow",
	"slack",
	"webhook",
	"zapier",
	"swsd",
}

func (r *alertResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A Terraform resource for managing alerts.",
		Attributes: map[string]schema.Attribute{
			"id": resourceIdAttribute(),
			"name": schema.StringAttribute{
				Description: "Alert name.",
				Required:    true,
			},
			"notification_actions": schema.SetNestedAttribute{
				Description: "List of alert notifications that are sent when an alert triggers.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							Description: "Notification service type (email, msteams, amazonsns, webhook, ...).",
							Required:    true,
							Validators: []validator.String{
								validators.CaseInsensitiveSingleOption(lowerCaseSlice(notificationActionTypes)...),
							},
						},
						"configuration_ids": schema.ListAttribute{
							Description: "A list of notifications ids that should be triggered for this alert.",
							Required:    true,
							ElementType: types.StringType,
						},
						"resend_interval_seconds": schema.Int64Attribute{
							Description: "How often should the notification be resent in case alert keeps being triggered. Null means notification is sent only once. Valid values are 60, 600, 3600, 86400.",
							Optional:    true,
							Validators: []validator.Int64{
								int64validator.OneOf(60, 600, 3600, 86400),
							},
						},
					},
				},
			},
			"description": schema.StringAttribute{
				Description: "Alert description.",
				Optional:    true,
			},
			"severity": schema.StringAttribute{
				Description: "Alert severity.",
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
				Description: "True if the Alert should be evaluated.",
				Computed:    true,
				Optional:    true,
				Default:     booldefault.StaticBool(true),
			},
			"trigger_reset_actions": schema.BoolAttribute{
				Description: "True if a notification should be sent when an active alert returns to normal. Default is false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"conditions": schema.SetNestedAttribute{
				Description: "One or more conditions that must be met to trigger the alert. These conditions are evaluated as a logical AND.",
				Required:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"metric_name": schema.StringAttribute{
							Description: "The field name of the metric to be filtered on.",
							Required:    true,
						},
						"threshold": schema.StringAttribute{
							Description: "Operator and value that represent the threshold of an the alert. When the threshold is breached it triggers the alert. For Operator - binaryOperator:(=|!=|>|<|>=|<=), logicalOperator:(AND|OR) E.g. '>=10'",
							Required:    true,
						},
						"duration": schema.StringAttribute{
							Description: "How long the threshold has been met before triggering an alert.",
							Required:    true,
						},
						"aggregation_type": schema.StringAttribute{
							Description: "The aggregation function that will be applied to the metric.",
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
							Description: "A list of Entity IDs that will be used to filter on the alert. The alert will only trigger if the alert matches one or more of the entity IDs.",
							Optional:    true,
							ElementType: types.StringType,
						},
						"target_entity_types": schema.ListAttribute{
							Description: "The entity types that the alert will be applied to.",
							Required:    true,
							ElementType: types.StringType,
						},
						"include_tags": schema.SetNestedAttribute{
							Description: "Tag key and values to match in order to trigger an alert.",
							Optional:    true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"name": schema.StringAttribute{
										Description: "Tag key to match.",
										Optional:    true,
									},
									"values": schema.ListAttribute{
										Description: "Values to match.",
										Optional:    true,
										ElementType: types.StringType,
									},
								},
							},
						},
						"exclude_tags": schema.SetNestedAttribute{
							Description: "Tag key and values to match in order to not trigger an alert.",
							Optional:    true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"name": schema.StringAttribute{
										Description: "Tag key to match.",
										Optional:    true,
									},
									"values": schema.ListAttribute{
										Description: "Values to match.",
										Optional:    true,
										ElementType: types.StringType,
									},
								},
							},
						},
						"group_by_metric_tag": schema.ListAttribute{
							Description: "Group alert data for selected attribute.",
							Optional:    true,
							ElementType: types.StringType,
						},
						"not_reporting": schema.BoolAttribute{
							Description: "True if the alert should trigger when the metric is not reporting. If true, threshold must be '' and aggregation_type must be 'COUNT'.",
							Computed:    true,
							Optional:    true,
							Default:     booldefault.StaticBool(false),
						},
					},
				},
			},
			"notifications": schema.ListAttribute{
				Description:        "A list of notifications that should be triggered for this alert.",
				Optional:           true,
				ElementType:        types.StringType,
				DeprecationMessage: "This field is deprecated. Please use the notification_actions field instead.",
			},
			"runbook_link": schema.StringAttribute{
				Description: "A runbook is documentation of what steps to follow when something goes wrong.",
				Optional:    true,
			},
		},
	}
}
