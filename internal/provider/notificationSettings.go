package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	errUnsupportedNotificationType = errors.New("unsupported notification type")
)

type notificationSettings struct {
	Email                 types.Object `tfsdk:"email"`
	Slack                 types.Object `tfsdk:"slack"`
	PagerDuty             types.Object `tfsdk:"pagerduty"`
	MsTeams               types.Object `tfsdk:"msteams"`
	Webhook               types.Object `tfsdk:"webhook"`
	OpsGenie              types.Object `tfsdk:"opsgenie"`
	AmazonSNS             types.Object `tfsdk:"amazonsns"`
	Zapier                types.Object `tfsdk:"zapier"`
	Pushover              types.Object `tfsdk:"pushover"`
	SolarWindsServiceDesk types.Object `tfsdk:"swsd"`
	ServiceNow            types.Object `tfsdk:"servicenow"`
}

func NotificationSettingsAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"email":      types.ObjectType{AttrTypes: EmailAttributeTypes()},
		"slack":      types.ObjectType{AttrTypes: SlackAttributeTypes()},
		"pagerduty":  types.ObjectType{AttrTypes: PagerDutyAttributeTypes()},
		"msteams":    types.ObjectType{AttrTypes: MsTeamsAttributeTypes()},
		"webhook":    types.ObjectType{AttrTypes: WebhookAttributeTypes()},
		"opsgenie":   types.ObjectType{AttrTypes: OpsGenieAttributeTypes()},
		"amazonsns":  types.ObjectType{AttrTypes: AmazonSNSAttributeTypes()},
		"zapier":     types.ObjectType{AttrTypes: ZapierAttributeTypes()},
		"pushover":   types.ObjectType{AttrTypes: PushoverAttributeTypes()},
		"swsd":       types.ObjectType{AttrTypes: SolarWindsServiceDeskAttributeTypes()},
		"servicenow": types.ObjectType{AttrTypes: ServiceNowAttributeTypes()},
	}
}

type notificationSettingsEmail struct {
	Addresses types.Set `tfsdk:"addresses"`
}

type clientEmail struct {
	Addresses []clientEmailAddress `tfsdk:"addresses" json:"addresses"`
}

func EmailAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"addresses": types.SetType{ElemType: types.ObjectType{AttrTypes: EmailAddressAttributeTypes()}},
	}
}

type notificationSettingsEmailAddress struct {
	Id    types.String `tfsdk:"id"`
	Email types.String `tfsdk:"email"`
}

type clientEmailAddress struct {
	Id    *string `tfsdk:"id" json:"id"`
	Email string  `tfsdk:"email" json:"email"`
}

func EmailAddressAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":    types.StringType,
		"email": types.StringType,
	}
}

type notificationSettingsOpsGenie struct {
	HostName   types.String `tfsdk:"hostname"`
	ApiKey     types.String `tfsdk:"api_key"`
	Recipients types.String `tfsdk:"recipients"`
	Teams      types.String `tfsdk:"teams"`
	Tags       types.String `tfsdk:"tags"`
}

type clientOpsGenie struct {
	HostName   string `tfsdk:"hostname" json:"hostname"`
	ApiKey     string `tfsdk:"api_key" json:"apiKey"`
	Recipients string `tfsdk:"recipients" json:"recipients"`
	Teams      string `tfsdk:"teams" json:"teams"`
	Tags       string `tfsdk:"tags" json:"tags"`
}

func OpsGenieAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"hostname":   types.StringType,
		"api_key":    types.StringType,
		"recipients": types.StringType,
		"teams":      types.StringType,
		"tags":       types.StringType,
	}
}

type notificationSettingsSlack struct {
	Url types.String `tfsdk:"url"`
}

type clientSlack struct {
	Url string `tfsdk:"url" json:"url"`
}

func SlackAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"url": types.StringType,
	}
}

type notificationSettingsMsTeams struct {
	Url types.String `tfsdk:"url"`
}

type clientMsTeams struct {
	Url string `tfsdk:"url" json:"url"`
}

func MsTeamsAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"url": types.StringType,
	}
}

type notificationSettingsPagerDuty struct {
	RoutingKey types.String `tfsdk:"routing_key"`
	Summary    types.String `tfsdk:"summary"`
	DedupKey   types.String `tfsdk:"dedup_key"`
}

type clientPagerDuty struct {
	RoutingKey string `tfsdk:"routing_key" json:"routingKey"`
	Summary    string `tfsdk:"summary" json:"summary"`
	DedupKey   string `tfsdk:"dedup_key" json:"dedupKey"`
}

func PagerDutyAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"routing_key": types.StringType,
		"summary":     types.StringType,
		"dedup_key":   types.StringType,
	}
}

type notificationSettingsWebhook struct {
	Url             types.String `tfsdk:"url" `
	Method          types.String `tfsdk:"method"`
	AuthType        types.String `tfsdk:"auth_type"`
	AuthUsername    types.String `tfsdk:"auth_username"`
	AuthPassword    types.String `tfsdk:"auth_password"`
	AuthHeaderName  types.String `tfsdk:"auth_header_name"`
	AuthHeaderValue types.String `tfsdk:"auth_header_value"`
}

type clientWebhook struct {
	Url             string `tfsdk:"url" json:"url"`
	Method          string `tfsdk:"method" json:"method"`
	AuthType        string `tfsdk:"auth_type" json:"authType"`
	AuthUsername    string `tfsdk:"auth_username" json:"authUsername"`
	AuthPassword    string `tfsdk:"auth_password" json:"authPassword"`
	AuthHeaderName  string `tfsdk:"auth_header_name" json:"authHeaderName"`
	AuthHeaderValue string `tfsdk:"auth_header_value" json:"authHeaderValue"`
}

func WebhookAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"url":               types.StringType,
		"method":            types.StringType,
		"auth_type":         types.StringType,
		"auth_username":     types.StringType,
		"auth_password":     types.StringType,
		"auth_header_name":  types.StringType,
		"auth_header_value": types.StringType,
	}
}

type notificationSettingsAmazonSNS struct {
	TopicARN        types.String `tfsdk:"topic_arn"`
	AccessKeyID     types.String `tfsdk:"access_key_id"`
	SecretAccessKey types.String `tfsdk:"secret_access_key"`
}

type clientAmazonSNS struct {
	TopicARN        string `tfsdk:"topic_arn" json:"topicARN"`
	AccessKeyID     string `tfsdk:"access_key_id" json:"accessKeyId"`
	SecretAccessKey string `tfsdk:"secret_access_key" json:"secretAccessKey"`
}

func AmazonSNSAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"topic_arn":         types.StringType,
		"access_key_id":     types.StringType,
		"secret_access_key": types.StringType,
	}
}

type notificationSettingsZapier struct {
	Url types.String `tfsdk:"url"`
}

type clientZapier struct {
	Url string `tfsdk:"url" json:"url"`
}

func ZapierAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"url": types.StringType,
	}
}

type notificationSettingsPushover struct {
	UserKey  types.String `tfsdk:"user_key"`
	AppToken types.String `tfsdk:"app_token"`
}

type clientPushover struct {
	UserKey  string `tfsdk:"user_key" json:"userKey"`
	AppToken string `tfsdk:"app_token" json:"appToken"`
}

func PushoverAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"user_key":  types.StringType,
		"app_token": types.StringType,
	}
}

type notificationSettingsSolarWindsServiceDesk struct {
	AppToken types.String `tfsdk:"app_token"`
	IsEU     types.Bool   `tfsdk:"is_eu"`
}

type clientSolarWindsServiceDesk struct {
	AppToken string `tfsdk:"app_token" json:"appToken"`
	IsEU     bool   `tfsdk:"is_eu" json:"isEu"`
}

func SolarWindsServiceDeskAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"app_token": types.StringType,
		"is_eu":     types.BoolType,
	}
}

type notificationSettingsServiceNow struct {
	AppToken types.String `tfsdk:"app_token"`
	Instance types.String `tfsdk:"instance"`
}

type clientServiceNow struct {
	AppToken string `tfsdk:"app_token" json:"appToken"`
	Instance string `tfsdk:"instance" json:"instance"`
}

func ServiceNowAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"app_token": types.StringType,
		"instance":  types.StringType,
	}
}

type notificationSettingsAccessor struct {
	/// Translates the TF type notification model into the JSON client model
	Get func(m *notificationSettings, ctx context.Context, diags *diag.Diagnostics) any
	/// Copy non-sensitive properties from the JSON client model into the TF type notification model
	Set func(m *notificationSettings, settings any, ctx context.Context, diags *diag.Diagnostics)
}

var settingsAccessors = map[string]notificationSettingsAccessor{
	"amazonsns": {
		Get: func(m *notificationSettings, ctx context.Context, diags *diag.Diagnostics) any {
			var amazonSns notificationSettingsAmazonSNS
			d := m.AmazonSNS.As(ctx, &amazonSns, basetypes.ObjectAsOptions{})
			if d.HasError() {
				diags.Append(d...)
				return nil
			}
			return clientAmazonSNS{
				TopicARN:        amazonSns.TopicARN.ValueString(),
				AccessKeyID:     amazonSns.AccessKeyID.ValueString(),
				SecretAccessKey: amazonSns.SecretAccessKey.ValueString(),
			}
		},
		Set: func(m *notificationSettings, settings any, ctx context.Context, diags *diag.Diagnostics) {
			settingsStruct, err := toSettingsStruct[clientAmazonSNS](settings)
			if err != nil {
				diags.AddError("Marshal Error",
					fmt.Sprintf("Error marshalling 'amazonsns' settings: %s", err))
				return
			}

			if !m.AmazonSNS.IsNull() {
				var amazonSns notificationSettingsAmazonSNS
				d := m.AmazonSNS.As(ctx, &amazonSns, basetypes.ObjectAsOptions{})
				if d.HasError() {
					diags.Append(d...)
					return
				}

				amazonSns.TopicARN = types.StringValue(settingsStruct.TopicARN)
				amazonSns.AccessKeyID = types.StringValue(settingsStruct.AccessKeyID)
				tfObject, d := types.ObjectValueFrom(ctx, AmazonSNSAttributeTypes(), amazonSns)
				if d.HasError() {
					diags.Append(d...)
					return
				}
				m.AmazonSNS = tfObject
			} else {
				tfObject, d := types.ObjectValueFrom(ctx, AmazonSNSAttributeTypes(), settingsStruct)
				if d.HasError() {
					diags.Append(d...)
					return
				}
				m.AmazonSNS = tfObject
			}
		},
	},
	"email": {
		Get: func(m *notificationSettings, ctx context.Context, diags *diag.Diagnostics) any {
			var email notificationSettingsEmail
			d := m.Email.As(ctx, &email, basetypes.ObjectAsOptions{})
			if d.HasError() {
				diags.Append(d...)
				return nil
			}
			var addresses []notificationSettingsEmailAddress
			d = email.Addresses.ElementsAs(ctx, &addresses, false)
			if d.HasError() {
				diags.Append(d...)
				return nil
			}

			var clientAddresses []clientEmailAddress
			clientAddresses = convertArray(addresses, func(h notificationSettingsEmailAddress) clientEmailAddress {
				return clientEmailAddress{
					Id:    h.Id.ValueStringPointer(),
					Email: h.Email.ValueString(),
				}
			})
			return clientEmail{
				Addresses: clientAddresses,
			}
		},
		Set: func(m *notificationSettings, settings any, ctx context.Context, diags *diag.Diagnostics) {
			settingsStruct, err := toSettingsStruct[clientEmail](settings)
			if err != nil {
				diags.AddError("Marshal Error",
					fmt.Sprintf("Error marshalling 'email' settings: %s", err))
				return
			}

			var elements []attr.Value
			for _, ss := range settingsStruct.Addresses {
				objectValue, d := types.ObjectValueFrom(
					ctx,
					EmailAddressAttributeTypes(),
					notificationSettingsEmailAddress{
						Id:    types.StringPointerValue(ss.Id),
						Email: types.StringValue(ss.Email),
					},
				)
				if d.HasError() {
					diags.Append(d...)
					return
				}
				elements = append(elements, objectValue)
			}

			setValue, d := types.SetValueFrom(ctx, types.ObjectType{AttrTypes: EmailAddressAttributeTypes()}, elements)
			if d.HasError() {
				diags.Append(d...)
				return
			}
			var email notificationSettingsEmail
			m.Email.As(ctx, &email, basetypes.ObjectAsOptions{})
			email.Addresses = setValue
			tfObject, d := types.ObjectValueFrom(ctx, EmailAttributeTypes(), email)
			if d.HasError() {
				diags.Append(d...)
				return
			}
			m.Email = tfObject
		},
	},
	"msTeams": {
		Get: func(m *notificationSettings, ctx context.Context, diags *diag.Diagnostics) any {
			var msTeams notificationSettingsMsTeams
			d := m.MsTeams.As(ctx, &msTeams, basetypes.ObjectAsOptions{})
			if d.HasError() {
				diags.Append(d...)
				return nil
			}
			return clientMsTeams{
				Url: msTeams.Url.ValueString(),
			}
		},
		Set: func(m *notificationSettings, settings any, ctx context.Context, diags *diag.Diagnostics) {
			settingsStruct, err := toSettingsStruct[clientMsTeams](settings)
			if err != nil {
				diags.AddError("Marshal Error",
					fmt.Sprintf("Error marshalling 'msTeams' settings: %s", err))
				return
			}
			tfObject, d := types.ObjectValueFrom(ctx, MsTeamsAttributeTypes(), settingsStruct)
			if d.HasError() {
				diags.Append(d...)
				return
			}
			m.MsTeams = tfObject
		},
	},
	"opsgenie": {
		Get: func(m *notificationSettings, ctx context.Context, diags *diag.Diagnostics) any {
			var opsGenie notificationSettingsOpsGenie
			d := m.OpsGenie.As(ctx, &opsGenie, basetypes.ObjectAsOptions{})
			if d.HasError() {
				return nil
			}
			c := clientOpsGenie{
				HostName:   opsGenie.HostName.ValueString(),
				ApiKey:     opsGenie.ApiKey.ValueString(),
				Recipients: opsGenie.Recipients.ValueString(),
				Teams:      opsGenie.Teams.ValueString(),
				Tags:       opsGenie.Tags.ValueString(),
			}
			return c
		},
		Set: func(m *notificationSettings, settings any, ctx context.Context, diags *diag.Diagnostics) {
			settingsStruct, err := toSettingsStruct[clientOpsGenie](settings)
			if err != nil {
				diags.AddError("Marshal Error",
					fmt.Sprintf("Error marshalling 'opsgenie' settings: %s", err))
				return
			}

			if !m.OpsGenie.IsNull() {
				var opsGenie notificationSettingsOpsGenie
				d := m.OpsGenie.As(ctx, &opsGenie, basetypes.ObjectAsOptions{})
				if d.HasError() {
					return
				}

				opsGenie.HostName = types.StringValue(settingsStruct.HostName)
				opsGenie.Recipients = types.StringValue(settingsStruct.Recipients)
				opsGenie.Teams = types.StringValue(settingsStruct.Teams)
				opsGenie.Tags = types.StringValue(settingsStruct.Tags)
				tfObject, d := types.ObjectValueFrom(ctx, OpsGenieAttributeTypes(), opsGenie)
				if d.HasError() {
					diags.Append(d...)
					return
				}
				m.OpsGenie = tfObject
			} else {
				tfObject, d := types.ObjectValueFrom(ctx, OpsGenieAttributeTypes(), settingsStruct)
				if d.HasError() {
					diags.Append(d...)
					return
				}
				m.OpsGenie = tfObject
			}
		},
	},
	"pagerduty": {
		Get: func(m *notificationSettings, ctx context.Context, diags *diag.Diagnostics) any {
			var pagerDuty notificationSettingsPagerDuty
			d := m.PagerDuty.As(ctx, &pagerDuty, basetypes.ObjectAsOptions{})
			if d.HasError() {
				return nil
			}
			return clientPagerDuty{
				RoutingKey: pagerDuty.RoutingKey.ValueString(),
				Summary:    pagerDuty.Summary.ValueString(),
				DedupKey:   pagerDuty.DedupKey.ValueString(),
			}
		},
		Set: func(m *notificationSettings, settings any, ctx context.Context, diags *diag.Diagnostics) {
			settingsStruct, err := toSettingsStruct[clientPagerDuty](settings)
			if err != nil {
				diags.AddError("Marshal Error",
					fmt.Sprintf("Error marshalling 'pagerduty' settings: %s", err))
				return
			}

			if !m.PagerDuty.IsNull() {
				var pagerDuty notificationSettingsPagerDuty
				d := m.PagerDuty.As(ctx, &pagerDuty, basetypes.ObjectAsOptions{})
				if d.HasError() {
					return
				}

				pagerDuty.Summary = types.StringValue(settingsStruct.Summary)
				pagerDuty.DedupKey = types.StringValue(settingsStruct.DedupKey)
				tfObject, d := types.ObjectValueFrom(ctx, PagerDutyAttributeTypes(), pagerDuty)
				if d.HasError() {
					diags.Append(d...)
					return
				}
				m.PagerDuty = tfObject

			} else {
				tfObject, d := types.ObjectValueFrom(ctx, PagerDutyAttributeTypes(), settingsStruct)
				if d.HasError() {
					diags.Append(d...)
					return
				}
				m.PagerDuty = tfObject
			}
		},
	},
	"pushover": {
		Get: func(m *notificationSettings, ctx context.Context, diags *diag.Diagnostics) any {
			var pushover notificationSettingsPushover
			d := m.Pushover.As(ctx, &pushover, basetypes.ObjectAsOptions{})
			if d.HasError() {
				diags.Append(d...)
				return nil
			}
			return clientPushover{
				UserKey:  pushover.UserKey.ValueString(),
				AppToken: pushover.AppToken.ValueString(),
			}
		},
		Set: func(m *notificationSettings, settings any, ctx context.Context, diags *diag.Diagnostics) {
			settingsStruct, err := toSettingsStruct[clientPushover](settings)
			if err != nil {
				diags.AddError("Marshal Error",
					fmt.Sprintf("Error marshalling 'pushover' settings: %s", err))
				return
			}

			if !m.Pushover.IsNull() {
				var pushover notificationSettingsPushover
				d := m.Pushover.As(ctx, &pushover, basetypes.ObjectAsOptions{})

				pushover.UserKey = types.StringValue(settingsStruct.UserKey)

				tfObject, d := types.ObjectValueFrom(ctx, PushoverAttributeTypes(), pushover)
				if d.HasError() {
					diags.Append(d...)
					return
				}
				m.Pushover = tfObject
			} else {
				tfObject, d := types.ObjectValueFrom(ctx, PushoverAttributeTypes(), settingsStruct)
				if d.HasError() {
					diags.Append(d...)
					return
				}
				m.Pushover = tfObject
			}
		},
	},
	"servicenow": {
		Get: func(m *notificationSettings, ctx context.Context, diags *diag.Diagnostics) any {
			var serviceNow notificationSettingsServiceNow
			d := m.ServiceNow.As(ctx, &serviceNow, basetypes.ObjectAsOptions{})
			if d.HasError() {
				diags.Append(d...)
				return nil
			}
			return clientServiceNow{
				AppToken: serviceNow.AppToken.ValueString(),
				Instance: serviceNow.Instance.ValueString(),
			}
		},
		Set: func(m *notificationSettings, settings any, ctx context.Context, diags *diag.Diagnostics) {
			settingsStruct, err := toSettingsStruct[clientServiceNow](settings)
			if err != nil {
				diags.AddError("Marshal Error",
					fmt.Sprintf("Error marshalling 'servicenow' settings: %s", err))
				return
			}

			if !m.ServiceNow.IsNull() {
				var serviceNow notificationSettingsServiceNow
				d := m.ServiceNow.As(ctx, &serviceNow, basetypes.ObjectAsOptions{})
				if d.HasError() {
					diags.Append(d...)
					return
				}

				serviceNow.Instance = types.StringValue(settingsStruct.Instance)
				tfObject, d := types.ObjectValueFrom(ctx, ServiceNowAttributeTypes(), serviceNow)
				if d.HasError() {
					diags.Append(d...)
					return
				}
				m.ServiceNow = tfObject
			} else {
				tfObject, d := types.ObjectValueFrom(ctx, ServiceNowAttributeTypes(), settingsStruct)
				if d.HasError() {
					diags.Append(d...)
					return
				}
				m.ServiceNow = tfObject
			}
		},
	},
	"slack": {
		Get: func(m *notificationSettings, ctx context.Context, diags *diag.Diagnostics) any {
			var slack notificationSettingsSlack
			d := m.Slack.As(ctx, &slack, basetypes.ObjectAsOptions{})
			if d.HasError() {
				diags.Append(d...)
				return nil
			}
			return clientSlack{
				Url: slack.Url.ValueString(),
			}
		},
		Set: func(m *notificationSettings, settings any, ctx context.Context, diags *diag.Diagnostics) {
			settingsStruct, err := toSettingsStruct[clientSlack](settings)
			if err != nil {
				diags.AddError("Marshal Error",
					fmt.Sprintf("Error marshalling 'slack' settings: %s", err))
				return
			}
			tfObject, d := types.ObjectValueFrom(ctx, SlackAttributeTypes(), settingsStruct)
			if d.HasError() {
				diags.Append(d...)
				return
			}
			m.Slack = tfObject
		},
	},
	"swsd": {
		Get: func(m *notificationSettings, ctx context.Context, diags *diag.Diagnostics) any {
			var swoServiceDesk notificationSettingsSolarWindsServiceDesk
			d := m.SolarWindsServiceDesk.As(ctx, &swoServiceDesk, basetypes.ObjectAsOptions{})
			if d.HasError() {
				diags.Append(d...)
				return nil
			}
			return clientSolarWindsServiceDesk{
				AppToken: swoServiceDesk.AppToken.ValueString(),
				IsEU:     swoServiceDesk.IsEU.ValueBool(),
			}
		},
		Set: func(m *notificationSettings, settings any, ctx context.Context, diags *diag.Diagnostics) {
			settingsStruct, err := toSettingsStruct[clientSolarWindsServiceDesk](settings)
			if err != nil {
				diags.AddError("Marshal Error",
					fmt.Sprintf("Error marshalling 'swsd' settings: %s", err))
				return
			}

			if !m.SolarWindsServiceDesk.IsNull() {
				var swoServiceDesk notificationSettingsSolarWindsServiceDesk
				d := m.SolarWindsServiceDesk.As(ctx, &swoServiceDesk, basetypes.ObjectAsOptions{})
				if d.HasError() {
					diags.Append(d...)
					return
				}

				swoServiceDesk.IsEU = types.BoolValue(settingsStruct.IsEU)
				tfObject, d := types.ObjectValueFrom(ctx, SolarWindsServiceDeskAttributeTypes(), swoServiceDesk)
				if d.HasError() {
					diags.Append(d...)
					return
				}
				m.SolarWindsServiceDesk = tfObject
			} else {
				tfObject, d := types.ObjectValueFrom(ctx, SolarWindsServiceDeskAttributeTypes(), settingsStruct)
				if d.HasError() {
					diags.Append(d...)
					return
				}
				m.SolarWindsServiceDesk = tfObject
			}
		},
	},
	"webhook": {
		Get: func(m *notificationSettings, ctx context.Context, diags *diag.Diagnostics) any {
			var webhook notificationSettingsWebhook
			d := m.Webhook.As(ctx, &webhook, basetypes.ObjectAsOptions{})
			if d.HasError() {
				diags.Append(d...)
				return nil
			}
			return clientWebhook{
				Url:             webhook.Url.ValueString(),
				Method:          webhook.Method.ValueString(),
				AuthType:        webhook.AuthType.ValueString(),
				AuthUsername:    webhook.AuthUsername.ValueString(),
				AuthPassword:    webhook.AuthPassword.ValueString(),
				AuthHeaderName:  webhook.AuthHeaderName.ValueString(),
				AuthHeaderValue: webhook.AuthHeaderValue.ValueString(),
			}
		},
		Set: func(m *notificationSettings, settings any, ctx context.Context, diags *diag.Diagnostics) {
			settingsStruct, err := toSettingsStruct[clientWebhook](settings)
			if err != nil {
				diags.AddError("Marshal Error",
					fmt.Sprintf("Error marshalling 'webhook' settings: %s", err))
				return
			}
			settingsStruct.AuthPassword = "PASSWORD"
			settingsStruct.AuthHeaderValue = "VALUE"

			if !m.Webhook.IsNull() {
				var webhook notificationSettingsWebhook
				d := m.Webhook.As(ctx, &webhook, basetypes.ObjectAsOptions{})
				if d.HasError() {
					diags.Append(d...)
					return
				}

				webhook.Url = types.StringValue(settingsStruct.Url)
				webhook.Method = types.StringValue(settingsStruct.Method)
				webhook.AuthType = types.StringValue(settingsStruct.AuthType)
				webhook.AuthUsername = types.StringValue(settingsStruct.AuthUsername)
				webhook.AuthHeaderName = types.StringValue(settingsStruct.AuthHeaderName)

				tfObject, d := types.ObjectValueFrom(ctx, WebhookAttributeTypes(), webhook)
				if d.HasError() {
					diags.Append(d...)
					return
				}
				m.Webhook = tfObject

			} else {
				tfObject, d := types.ObjectValueFrom(ctx, WebhookAttributeTypes(), settingsStruct)
				if d.HasError() {
					diags.Append(d...)
					return
				}
				m.Webhook = tfObject
			}
		},
	},
	"zapier": {
		Get: func(m *notificationSettings, ctx context.Context, diags *diag.Diagnostics) any {
			var zapier notificationSettingsZapier
			d := m.Zapier.As(ctx, &zapier, basetypes.ObjectAsOptions{})
			if d.HasError() {
				diags.Append(d...)
				return nil
			}
			return clientZapier{
				Url: zapier.Url.ValueString(),
			}
		},
		Set: func(m *notificationSettings, settings any, ctx context.Context, diags *diag.Diagnostics) {
			settingsStruct, err := toSettingsStruct[clientZapier](settings)
			if err != nil {
				diags.AddError("Marshal Error",
					fmt.Sprintf("Error marshalling 'zapier' settings: %s", err))
				return
			}
			tfObject, d := types.ObjectValueFrom(ctx, ZapierAttributeTypes(), settingsStruct)
			if d.HasError() {
				diags.Append(d...)
			}
			m.Zapier = tfObject
		},
	},
}

// Utility function to marshal anonymous JSON to concrete models defined by T.
func toSettingsStruct[T any](settings any) (*T, error) {
	data, err := json.Marshal(settings)
	if err != nil {
		return nil, err
	}

	var concreteSettings T
	err = json.Unmarshal(data, &concreteSettings)
	if err != nil {
		return nil, err
	}

	return &concreteSettings, nil
}

func (m *notificationResourceModel) SetSettings(clientSettings *any, ctx context.Context, diags *diag.Diagnostics) {

	if accessor, found := settingsAccessors[m.Type.ValueString()]; found {
		if m.Settings.IsNull() {
			var model = notificationSettings{
				Email:                 types.ObjectNull(EmailAttributeTypes()),
				Slack:                 types.ObjectNull(SlackAttributeTypes()),
				PagerDuty:             types.ObjectNull(PagerDutyAttributeTypes()),
				MsTeams:               types.ObjectNull(MsTeamsAttributeTypes()),
				Webhook:               types.ObjectNull(WebhookAttributeTypes()),
				OpsGenie:              types.ObjectNull(OpsGenieAttributeTypes()),
				AmazonSNS:             types.ObjectNull(AmazonSNSAttributeTypes()),
				Zapier:                types.ObjectNull(ZapierAttributeTypes()),
				Pushover:              types.ObjectNull(PushoverAttributeTypes()),
				SolarWindsServiceDesk: types.ObjectNull(SolarWindsServiceDeskAttributeTypes()),
				ServiceNow:            types.ObjectNull(ServiceNowAttributeTypes()),
			}

			accessor.Set(&model, clientSettings, ctx, diags)
			if diags.HasError() {
				return
			}

			tfSettings, d := types.ObjectValueFrom(ctx, NotificationSettingsAttributeTypes(), model)
			if d.HasError() {
				diags.Append(d...)
				return
			}
			m.Settings = tfSettings
		} else {
			var notification notificationSettings
			d := m.Settings.As(ctx, &notification, basetypes.ObjectAsOptions{})
			if d.HasError() {
				diags.Append(d...)
				return
			}

			accessor.Set(&notification, clientSettings, ctx, diags)
			if diags.HasError() {
				return
			}
			tfSettings, d := types.ObjectValueFrom(ctx, NotificationSettingsAttributeTypes(), notification)
			if d.HasError() {
				diags.Append(d...)
				return
			}
			m.Settings = tfSettings
			return
		}
	} else {
		diags.AddError("Unsupported Notification Type Error",
			fmt.Sprintf("%s: %s", errUnsupportedNotificationType, m.Type.ValueString()))
	}
}

func (m *notificationResourceModel) GetSettings(ctx context.Context, diags *diag.Diagnostics) any {

	if accessor, found := settingsAccessors[m.Type.ValueString()]; found {
		if m.Settings.IsNull() {
			m.Settings = types.ObjectNull(NotificationSettingsAttributeTypes())
			return nil
		}

		var settings notificationSettings
		d := m.Settings.As(ctx, &settings, basetypes.ObjectAsOptions{})
		if d.HasError() {
			diags.Append(d...)
			return nil
		}
		clientSettings := accessor.Get(&settings, ctx, diags)
		if diags.HasError() {
			return nil
		}
		return clientSettings
	} else {
		diags.AddError("Unsupported Notification Type Error",
			fmt.Sprintf("%s: %s", errUnsupportedNotificationType, m.Type.ValueString()))
		return nil
	}
}
