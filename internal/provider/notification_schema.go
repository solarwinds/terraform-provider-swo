package provider

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/solarwindscloud/terraform-provider-swo/internal/envvar"
)

const (
	emailRegex               = `^(([^<>()[\]\\.,;:\s@"]+(\.[^<>()[\]\\.,;:\s@"]+)*)|(".+"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$`
	phoneNumberRegex         = `^\+(?:[0-9] ?){6,14}[0-9]$`
	snsTopicArnRegex         = `^arn:aws:sns:[^:]+:[0-9]+:[a-zA-Z0-9\-_]+$`
	zapierHooksRegex         = `^https:\/\/hooks\.zapier\.com\/hooks\/catch.*`
	slackHooksRegex          = `^https:\/\/hooks\.slack.com.*`
	pagerDutyRoutingKeyRegex = `^[a-zA-Z0-9]{32}$`
	msTeamsRegex             = `^https:\/\/.*\.office\.com\/webhook.*`
	httpSchemeRegex          = `^(http|https)`
)

// The main Notification Resource model that is derived from the schema.
type NotificationResourceModel struct {
	Id          types.String          `tfsdk:"id"`
	Title       types.String          `tfsdk:"title"`
	Description types.String          `tfsdk:"description"`
	Type        types.String          `tfsdk:"type"`
	Settings    *NotificationSettings `tfsdk:"settings"`
	CreatedAt   types.String          `tfsdk:"created_at"`
	CreatedBy   types.String          `tfsdk:"created_by"`
}

func (m *NotificationResourceModel) ParseId() (id string, notificationType string, err error) {
	idParts := strings.Split(m.Id.ValueString(), ":")

	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		err = fmt.Errorf("expected identifier with format: id,type. got: %q", m.Id)
	} else {
		id = idParts[0]
		notificationType = idParts[1]
	}

	return id, notificationType, err
}

func (r *NotificationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	tflog.Trace(ctx, "NotificationResource: Schema")

	resp.Schema = schema.Schema{
		Description: fmt.Sprintf("A terraform resource for managing %s notifications.", envvar.AppName),
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Notification ID in UUID format. This is a computed value provided by the backend.",
				Computed:    true,
			},
			"title": schema.StringAttribute{
				Description: "Notification title.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "A short description of the Notification. (Optional)",
				Optional:    true,
			},
			"type": schema.StringAttribute{
				Description: "Notification type (email, slack, etc).",
				Required:    true,
			},
			"created_by": schema.StringAttribute{
				Description: "Who created this notification?",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "When was this notification created?",
				Computed:    true,
			},
			"settings": schema.SingleNestedAttribute{
				Description: "The notification settings.",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"email": schema.SingleNestedAttribute{
						Description: "Email settings. (Optional)",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"addresses": schema.SetNestedAttribute{
								Description: "Email addresses for email notifications.",
								Required:    true,
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Description: "The user id associated to the email address. (Optional)",
											Optional:    true,
										},
										"email": schema.StringAttribute{
											Description: "The email address.",
											Required:    true,
											Validators: []validator.String{
												stringvalidator.RegexMatches(
													regexp.MustCompile(emailRegex),
													"Requirement: "+emailRegex,
												),
											},
										},
									},
								},
							},
						},
					},
					"slack": schema.SingleNestedAttribute{
						Description: "Slack settings. (Optional)",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"url": schema.StringAttribute{
								Description: "",
								Required:    true,
								Validators: []validator.String{
									stringvalidator.RegexMatches(
										regexp.MustCompile(slackHooksRegex),
										"Requirement: "+slackHooksRegex,
									),
								},
							},
						},
					},
					"pagerduty": schema.SingleNestedAttribute{
						Description: "PagerDuty settings. (Optional)",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"routing_key": schema.StringAttribute{
								Description: "",
								Required:    true,
								Sensitive:   true,
								Validators: []validator.String{
									stringvalidator.RegexMatches(
										regexp.MustCompile(pagerDutyRoutingKeyRegex),
										"Requirement: "+pagerDutyRoutingKeyRegex,
									),
								},
							},
							"summary": schema.StringAttribute{
								Description: "",
								Required:    true,
								Validators: []validator.String{
									stringvalidator.LengthAtMost(1024),
								},
							},
							"dedup_key": schema.StringAttribute{
								Description: "",
								Required:    true,
							},
						},
					},
					"webhook": schema.SingleNestedAttribute{
						Description: "Webhook settings. (Optional)",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"url": schema.StringAttribute{
								Description: "",
								Required:    true,
								Validators: []validator.String{
									stringvalidator.RegexMatches(
										regexp.MustCompile(httpSchemeRegex),
										"Requirement: "+httpSchemeRegex,
									),
								},
							},
						},
					},
					"victorops": schema.SingleNestedAttribute{
						Description: "VictorOps settings. (Optional)",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"api_key": schema.StringAttribute{
								Description: "",
								Required:    true,
								Sensitive:   true,
							},
							"routing_key": schema.StringAttribute{
								Description: "",
								Optional:    true,
								Sensitive:   true,
							},
						},
					},
					"opsgenie": schema.SingleNestedAttribute{
						Description: "OpsGenie settings. (Optional)",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"hostname": schema.StringAttribute{
								Description: "",
								Optional:    true,
							},
							"api_key": schema.StringAttribute{
								Description: "",
								Required:    true,
								Sensitive:   true,
							},
							"recipients": schema.StringAttribute{
								Description: "",
								Required:    true,
							},
							"teams": schema.StringAttribute{
								Description: "",
								Required:    true,
							},
							"tags": schema.StringAttribute{
								Description: "",
								Required:    true,
							},
						},
					},
					"amazonsns": schema.SingleNestedAttribute{
						Description: "Amazon SNS settings. (Optional)",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"topic_arn": schema.StringAttribute{
								Description: "",
								Required:    true,
								Validators: []validator.String{
									stringvalidator.RegexMatches(
										regexp.MustCompile(snsTopicArnRegex),
										"Requirement: "+snsTopicArnRegex,
									),
								},
							},
							"access_key_id": schema.StringAttribute{
								Description: "",
								Required:    true,
							},
							"secret_access_key": schema.StringAttribute{
								Description: "",
								Required:    true,
								Sensitive:   true,
							},
						},
					},
					"zapier": schema.SingleNestedAttribute{
						Description: "Zapier settings. (Optional)",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"url": schema.StringAttribute{
								Description: "",
								Required:    true,
								Validators: []validator.String{
									stringvalidator.RegexMatches(
										regexp.MustCompile(zapierHooksRegex),
										"Requirement: "+zapierHooksRegex,
									),
								},
							},
						},
					},
					"msteams": schema.SingleNestedAttribute{
						Description: "Microsoft Teams settings. (Optional)",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"url": schema.StringAttribute{
								Description: "",
								Required:    true,
								Validators: []validator.String{
									stringvalidator.RegexMatches(
										regexp.MustCompile(msTeamsRegex),
										"Requirement: "+msTeamsRegex,
									),
								},
							},
						},
					},
					"pushover": schema.SingleNestedAttribute{
						Description: "Pushover settings. (Optional)",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"user_key": schema.StringAttribute{
								Description: "",
								Required:    true,
							},
							"app_token": schema.StringAttribute{
								Description: "",
								Required:    true,
								Sensitive:   true,
							},
						},
					},
					"sms": schema.SingleNestedAttribute{
						Description: "SMS settings. (Optional)",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"phone_numbers": schema.StringAttribute{
								Description: "",
								Required:    true,
								Validators: []validator.String{
									stringvalidator.RegexMatches(
										regexp.MustCompile(phoneNumberRegex),
										"Requirement: "+phoneNumberRegex,
									),
								},
							},
						},
					},
					"swsd": schema.SingleNestedAttribute{
						Description: "SolarWinds Service Desk settings. (Optional)",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"app_token": schema.StringAttribute{
								Description: "",
								Required:    true,
								Sensitive:   true,
							},
							"is_eu": schema.BoolAttribute{
								Description: "",
								Required:    true,
							},
						},
					},
					"servicenow": schema.SingleNestedAttribute{
						Description: "ServiceNow settings. (Optional)",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"app_token": schema.StringAttribute{
								Description: "",
								Required:    true,
								Sensitive:   true,
							},
							"instance": schema.StringAttribute{
								Description: "",
								Required:    true,
							},
						},
					},
				},
			},
		},
	}
}
