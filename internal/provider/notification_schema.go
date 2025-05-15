package provider

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/solarwinds/terraform-provider-swo/internal/validators"
)

const (
	emailRegex               = `^(([^<>()[\]\\.,;:\s@"]+(\.[^<>()[\]\\.,;:\s@"]+)*)|(".+"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$`
	phoneNumberRegex         = `^\+(?:[0-9] ?){6,14}[0-9]$`
	snsTopicArnRegex         = `^arn:aws:sns:[^:]+:[0-9]+:[a-zA-Z0-9\-_]+$`
	zapierHooksRegex         = `^https:\/\/hooks\.zapier\.com\/hooks\/catch.*`
	slackHooksRegex          = `^https:\/\/hooks\.slack\.com.*`
	pagerDutyRoutingKeyRegex = `^[a-zA-Z0-9]{32}$`
	msTeamsRegex             = `^https:\/\/.*\.office\.com\/webhook.*`
	httpSchemeRegex          = `^(http|https)`
)

var (
	errParse = errors.New("parser error")
)

func newParseError(msg string) error {
	return fmt.Errorf("%w: %s", errParse, msg)
}

// The main Notification Resource model that is derived from the schema.
type notificationResourceModel struct {
	Id          types.String `tfsdk:"id"`
	Title       types.String `tfsdk:"title"`
	Description types.String `tfsdk:"description"`
	Type        types.String `tfsdk:"type"`
	//Settings    *notificationSettings `tfsdk:"settings"`
	Settings types.Object `tfsdk:"settings"`
}

func ParseNotificationId(id types.String) (idValue string, notificationType string, err error) {
	idParts := strings.Split(id.ValueString(), ":")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		err = newParseError(fmt.Sprintf("expected identifier with format id:type. got: %q", id.ValueString()))
	} else {
		idValue = idParts[0]
		notificationType = idParts[1]
	}
	return idValue, notificationType, err
}

func (r *notificationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A terraform resource for managing notifications.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The Id of the resource provided by the backend in the format of `{id}:{type}`.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"title": schema.StringAttribute{
				Description: "The title of the notification.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "A short description of the notification.",
				Optional:    true,
			},
			"type": schema.StringAttribute{
				Description: "Notification type (email, slack, etc).",
				Required:    true,
			},
			"settings": schema.SingleNestedAttribute{
				Description: "The notification settings.",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"email": schema.SingleNestedAttribute{
						Description: "Email settings.",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"addresses": schema.SetNestedAttribute{
								Description: "Email addresses for email notifications.",
								Required:    true,
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Description: "The user id associated to the email address.",
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
						Description: "Integration for sending static alerts to a Slack channel.",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"url": schema.StringAttribute{
								Description: "Slack Incoming Webhook URL. (https://api.slack.com/messaging/webhooks)",
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
						Description: "Integration for sending events to PagerDuty.",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"routing_key": schema.StringAttribute{
								Description: "Key for live call routing. (https://support.pagerduty.com/docs/live-call-routing)",
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
								Description: "A summary of the issue causing the alert to trigger.",
								Required:    true,
								Validators: []validator.String{
									stringvalidator.LengthAtMost(1024),
								},
							},
							"dedup_key": schema.StringAttribute{
								Description: "deduplication key for correlating trigger conditions. (https://support.pagerduty.com/docs/event-management) ",
								Required:    true,
							},
						},
					},
					"webhook": schema.SingleNestedAttribute{
						Description: "Integration with an existing notification service.",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"url": schema.StringAttribute{
								Description: "Webhook URL to an existing notification service.",
								Required:    true,
								Validators: []validator.String{
									stringvalidator.RegexMatches(
										regexp.MustCompile(httpSchemeRegex),
										"Requirement: "+httpSchemeRegex,
									),
								},
							},
							"method": schema.StringAttribute{
								Description: "HTTP Method for calling the webhook.",
								Required:    true,
								Validators: []validator.String{
									validators.SingleOption("POST", "GET"),
								},
							},
							"auth_type": schema.StringAttribute{
								Description: "Token or username/password auth.",
								Optional:    true,
								Validators: []validator.String{
									validators.SingleOption("basic", "token"),
								},
							},
							"auth_username": schema.StringAttribute{
								Description: "Username for basic auth type.",
								Optional:    true,
							},
							"auth_password": schema.StringAttribute{
								Description: "Password for basic auth type.",
								Optional:    true,
								Sensitive:   true,
							},
							"auth_header_name": schema.StringAttribute{
								Description: "Header name for token auth.",
								Optional:    true,
							},
							"auth_header_value": schema.StringAttribute{
								Description: "Header value for token auth.",
								Optional:    true,
								Sensitive:   true,
							},
						},
					},
					"victorops": schema.SingleNestedAttribute{
						Description: "Integration for sending events to VictorOps.",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"api_key": schema.StringAttribute{
								Description: "API Key for a VictorOps integration. (https://help.victorops.com/knowledge-base/api/)",
								Required:    true,
								Sensitive:   true,
							},
							"routing_key": schema.StringAttribute{
								Description: "Key for live call routing. (https://help.victorops.com/knowledge-base/routing-keys/)",
								Optional:    true,
								Sensitive:   true,
							},
						},
					},
					"opsgenie": schema.SingleNestedAttribute{
						Description: "Integration for sending alerts via email or using a Webhook to OpsGenie.",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"hostname": schema.StringAttribute{
								Description: "API Hostname",
								Optional:    true,
							},
							"api_key": schema.StringAttribute{
								Description: "API key from OpsGenie Integration. (https://support.atlassian.com/opsgenie/docs/api-key-management/)",
								Required:    true,
								Sensitive:   true,
							},
							"recipients": schema.StringAttribute{
								Description: "Specifies who should be notified by email for the alert.",
								Required:    true,
							},
							"teams": schema.StringAttribute{
								Description: "Specifies who should be notified by email for the alert.",
								Required:    true,
							},
							"tags": schema.StringAttribute{
								Description: "Any possible tags.",
								Required:    true,
							},
						},
					},
					"amazonsns": schema.SingleNestedAttribute{
						Description: "Integration for sending alerts to Amazon Simple Notification Service. Provides message delivery from publishers to subscribers.",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"topic_arn": schema.StringAttribute{
								Description: "Resource name that represents a logical access point that acts as a communication channel. (https://docs.aws.amazon.com/sns/latest/dg/sns-create-topic.html)",
								Required:    true,
								Validators: []validator.String{
									stringvalidator.RegexMatches(
										regexp.MustCompile(snsTopicArnRegex),
										"Requirement: "+snsTopicArnRegex,
									),
								},
							},
							"access_key_id": schema.StringAttribute{
								Description: "Access key ID for Amazon SNS.",
								Required:    true,
							},
							"secret_access_key": schema.StringAttribute{
								Description: "Secret access key for Amazon SNS.",
								Required:    true,
								Sensitive:   true,
							},
						},
					},
					"zapier": schema.SingleNestedAttribute{
						Description: "Integration for sending alerts to Zapier.",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"url": schema.StringAttribute{
								Description: "Zapier Webhook URL.",
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
						Description: "Integration for sending static alerts to a Microsoft Teams channel.",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"url": schema.StringAttribute{
								Description: "Microsoft Teams Webhook URL. (https://docs.microsoft.com/en-us/microsoftteams/platform/webhooks-and-connectors/how-to/add-incoming-webhook)",
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
						Description: "Integration for Sending alerts to Pushover.",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"user_key": schema.StringAttribute{
								Description: "User/Group key (or that of your target user), viewable when logged into the Pushover dashboard.",
								Required:    true,
							},
							"app_token": schema.StringAttribute{
								Description: "API token/APP token from registered Pushover application. (https://pushover.net/api)",
								Required:    true,
								Sensitive:   true,
							},
						},
					},
					"sms": schema.SingleNestedAttribute{
						Description: "For sending alerts Through SMS/Text.",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"phone_numbers": schema.StringAttribute{
								Description: "Phone number alerts will be texted to.",
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
						Description: "Integration with SolarWinds Observability creates new incidents based on SolarWinds Observability alerts.",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"app_token": schema.StringAttribute{
								Description: "Token copied from SolarWinds Service Desk",
								Required:    true,
								Sensitive:   true,
							},
							"is_eu": schema.BoolAttribute{
								Description: "Is in the EU.",
								Required:    true,
							},
						},
					},
					"servicenow": schema.SingleNestedAttribute{
						Description: "Integration with SolarWinds Observability creates new incidents based on SolarWinds Observability alerts.",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"app_token": schema.StringAttribute{
								Description: "ServiceNow access token",
								Required:    true,
								Sensitive:   true,
							},
							"instance": schema.StringAttribute{
								Description: "Instance name for this integration",
								Required:    true,
							},
						},
					},
				},
			},
		},
	}
}
