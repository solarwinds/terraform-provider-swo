package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/attr"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	swoClient "github.com/solarwinds/swo-client-go/pkg/client"
	"github.com/solarwinds/terraform-provider-swo/internal/validators"
)

// The main Alert Resource model that is derived from the schema.
type alertResourceModel struct {
	Id                  types.String `tfsdk:"id"`
	Name                types.String `tfsdk:"name"`
	NotificationActions types.Set    `tfsdk:"notification_actions"` //alertActionInputModel
	Description         types.String `tfsdk:"description"`
	Severity            types.String `tfsdk:"severity"`
	Enabled             types.Bool   `tfsdk:"enabled"`
	Conditions          types.Set    `tfsdk:"conditions"` //alertConditionModel
	Notifications       types.List   `tfsdk:"notifications"`
	TriggerResetActions types.Bool   `tfsdk:"trigger_reset_actions"`
	RunbookLink         types.String `tfsdk:"runbook_link"`
	TriggerDelaySeconds types.Int64  `tfsdk:"trigger_delay_seconds"`
}

type alertConditionModel struct {
	MetricName      types.String `tfsdk:"metric_name"`
	Threshold       types.String `tfsdk:"threshold"`
	Duration        types.String `tfsdk:"duration"`
	AggregationType types.String `tfsdk:"aggregation_type"`

	AttributeName     types.String `tfsdk:"attribute_name"`
	AttributeOperator types.String `tfsdk:"attribute_operator"`
	AttributeValue    types.String `tfsdk:"attribute_value"`
	AttributeValues   types.List   `tfsdk:"attribute_values"`

	EntityIds         types.List   `tfsdk:"entity_ids"`
	QuerySearch       types.String `tfsdk:"query_search"`
	TargetEntityTypes types.List   `tfsdk:"target_entity_types"`
	IncludeTags       types.Set    `tfsdk:"include_tags"`
	ExcludeTags       types.Set    `tfsdk:"exclude_tags"`
	GroupByMetricTag  types.List   `tfsdk:"group_by_metric_tag"`
	NotReporting      types.Bool   `tfsdk:"not_reporting"`
}

func AlertConditionAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"metric_name":      types.StringType,
		"threshold":        types.StringType,
		"duration":         types.StringType,
		"aggregation_type": types.StringType,

		"attribute_names":    types.StringType,
		"attribute_operator": types.StringType,
		"attribute_value":    types.StringType,
		"attribute_values":   types.ListType{ElemType: types.StringType},

		"entity_ids":          types.ListType{ElemType: types.StringType},
		"query_search":        types.StringType,
		"target_entity_types": types.ListType{ElemType: types.StringType},
		"include_tags":        types.SetType{ElemType: types.ObjectType{AttrTypes: AlertTagAttributeTypes()}},
		"exclude_tags":        types.SetType{ElemType: types.ObjectType{AttrTypes: AlertTagAttributeTypes()}},
		"group_by_metric_tag": types.ListType{ElemType: types.StringType},
		"not_reporting":       types.BoolType,
	}
}

type alertTagsModel struct {
	Name   types.String `tfsdk:"name"`
	Values types.List   `tfsdk:"values"`
}

func AlertTagAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":  types.StringType,
		"value": types.ListType{ElemType: types.StringType},
	}
}

type alertActionInputModel struct {
	ConfigurationIds      types.List  `tfsdk:"configuration_ids"`
	ResendIntervalSeconds types.Int64 `tfsdk:"resend_interval_seconds"`
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
						"configuration_ids": schema.ListAttribute{
							Description: "List of configuration_ids in `id:type` format. " +
								"Example: `[\"4661:email\", \"8112:webhook\", \"2456:newrelic\"]`. " +
								"Valid `type` values are [`email`|`amazonsns`|`msteams`|`newrelic`|`opsgenie`|`pagerduty`|`pushover`|`servicenow`|`slack`|`webhook`|`zapier`|`swsd`].",
							Required:    true,
							ElementType: types.StringType,
							Validators: []validator.List{
								validators.ListOptions(lowerCaseSlice(notificationActionTypes)...),
							},
						},
						"resend_interval_seconds": schema.Int64Attribute{
							Description: "How often should the notification be resent in case alert keeps being triggered. " +
								"Null means notification is sent only once. Value must be between 60 and 86400 seconds, and value must be divisible by 60.",
							Optional: true,
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
							Description: "The field name of the metric to be filtered on." +
								"Required field when condition is for a metric.",
							Optional: true,
						},
						"threshold": schema.StringAttribute{
							Description: "Operator and value that represent the threshold of an the alert. " +
								"When the threshold is breached it triggers the alert. " +
								"For Operator - binaryOperator:(=|!=|>|<|>=|<=) E.g. '>=10'" +
								"Required field when condition is for a metric.",
							Optional: true,
						},
						"duration": schema.StringAttribute{
							Description: "The duration window determines how frequently the alert is evaluated." +
								"Required field when condition is for a metric.",
							Optional: true,
						},
						"aggregation_type": schema.StringAttribute{
							Description: "The aggregation function that will be applied to the metric." +
								"Required field when condition is for a metric.",
							Optional: true,
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
						"not_reporting": schema.BoolAttribute{
							Description: "True if the alert should trigger when the metric is not reporting. " +
								"If true, threshold must be '' and aggregation_type must be 'COUNT'." +
								"Applies only when condition is for a metric.",
							Computed: true,
							Optional: true,
							Default:  booldefault.StaticBool(false),
						},

						"attribute_name": schema.StringAttribute{
							Description: "The attribute name of the entity to be filtered on." +
								"Required field when condition is for a attribute.",
							Optional: true,
						},
						"attribute_operator": schema.StringAttribute{
							Description: "Select an operator, and then specify the values that trigger this alert." +
								"Required field when condition is for a attribute.",
							Optional: true,
							Validators: []validator.String{
								validators.SingleOption(
									swoClient.AlertOperatorEq,
									swoClient.AlertOperatorNe,
									swoClient.AlertOperatorGt,
									swoClient.AlertOperatorLt,
									swoClient.AlertOperatorGe,
									swoClient.AlertOperatorLe,
									swoClient.AlertOperatorIn,
								),
							},
						},
						"attribute_value": schema.StringAttribute{
							Description: "Specify the value that trigger this alert." +
								"Required field when condition is for a attribute, and attribute_operator is not 'IN'.",
							Optional: true,
						},
						"attribute_values": schema.ListAttribute{
							Description: "Specify the set of values that trigger this alert." +
								"Required field when condition is for a attribute, and attribute_operator is 'IN'.",
							Optional:    true,
							ElementType: types.StringType,
						},

						"entity_ids": schema.ListAttribute{
							Description: "A list of Entity IDs that will be used to filter on the alert. " +
								"The alert will only trigger if the alert matches one or more of the entity IDs. " +
								"Must match across all alert conditions.",
							Optional:    true,
							ElementType: types.StringType,
						},
						"query_search": schema.StringAttribute{
							Description: "Case-sensitive. System will automatically match existing and newly added " +
								"entities matching the following query string.",
							Optional: true,
						},
						"target_entity_types": schema.ListAttribute{
							Description: "The entity types that the alert will be applied to. Must match across all alert conditions.",
							Optional:    true,
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
							Description: "Group alert data for selected attribute. Must match across all alert conditions.",
							Optional:    true,
							ElementType: types.StringType,
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
			"trigger_delay_seconds": schema.Int64Attribute{
				Description: "Trigger the alert after the alert condition persists for a specific duration. This prevents false positives. " +
					"Value must be between 60 and 86400 seconds, and value must be divisible by 60.",
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(0),
			},
		},
	}
}
