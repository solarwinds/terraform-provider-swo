package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"log"
)

var (
	errUnsupportedNotificationType = errors.New("unsupported notification type")
)

func newUnsupportedNotificationTypeError(notificationType string) error {
	return fmt.Errorf("%w: %s", errUnsupportedNotificationType, notificationType)
}

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
	Addresses types.Set `tfsdk:"addresses" json:"addresses"`
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
	Id    types.String `tfsdk:"id" json:"id"`
	Email types.String `tfsdk:"email" json:"email"`
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
	HostName   types.String `tfsdk:"hostname" json:"hostname"`
	ApiKey     types.String `tfsdk:"api_key" json:"apiKey"`
	Recipients types.String `tfsdk:"recipients" json:"recipients"`
	Teams      types.String `tfsdk:"teams" json:"teams"`
	Tags       types.String `tfsdk:"tags" json:"tags"`
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
	Url types.String `tfsdk:"url" json:"url"`
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
	Url types.String `tfsdk:"url" json:"url"`
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
	RoutingKey types.String `tfsdk:"routing_key" json:"routingKey"`
	Summary    types.String `tfsdk:"summary" json:"summary"`
	DedupKey   types.String `tfsdk:"dedup_key" json:"dedupKey"`
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
	Url             types.String `tfsdk:"url" json:"url"`
	Method          types.String `tfsdk:"method" json:"method"`
	AuthType        types.String `tfsdk:"auth_type" json:"authType"`
	AuthUsername    types.String `tfsdk:"auth_username" json:"authUsername"`
	AuthPassword    types.String `tfsdk:"auth_password" json:"authPassword"`
	AuthHeaderName  types.String `tfsdk:"auth_header_name" json:"authHeaderName"`
	AuthHeaderValue types.String `tfsdk:"auth_header_value" json:"authHeaderValue"`
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
	TopicARN        types.String `tfsdk:"topic_arn" json:"topicARN"`
	AccessKeyID     types.String `tfsdk:"access_key_id" json:"accessKeyID"`
	SecretAccessKey types.String `tfsdk:"secret_access_key" json:"secretAccessKey"`
}

type clientAmazonSNS struct {
	TopicARN        string `tfsdk:"topic_arn" json:"topicARN"`
	AccessKeyID     string `tfsdk:"access_key_id" json:"accessKeyID"`
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
	Url types.String `tfsdk:"url" json:"url"`
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
	UserKey  types.String `tfsdk:"user_key" json:"userKey"`
	AppToken types.String `tfsdk:"app_token" json:"appToken"`
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
	AppToken types.String `tfsdk:"app_token" json:"appToken"`
	IsEU     types.Bool   `tfsdk:"is_eu" json:"isEu"`
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
	AppToken types.String `tfsdk:"app_token" json:"appToken"`
	Instance types.String `tfsdk:"instance" json:"instance"`
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
	Get func(m *notificationSettings, ctx context.Context) any
	Set func(m *notificationSettings, settings any, ctx context.Context) error
}

var settingsAccessors = map[string]notificationSettingsAccessor{
	"email": {
		Get: func(m *notificationSettings, ctx context.Context) any {
			var email notificationSettingsEmail
			m.Email.As(ctx, &email, basetypes.ObjectAsOptions{})
			var addresses []notificationSettingsEmailAddress
			email.Addresses.ElementsAs(ctx, &addresses, false)

			var c []clientEmailAddress
			c = convertArray(addresses, func(h notificationSettingsEmailAddress) clientEmailAddress {
				return clientEmailAddress{
					Id:    h.Id.ValueStringPointer(),
					Email: h.Email.ValueString(),
				}
			})
			return clientEmail{
				Addresses: c,
			}
		},
		Set: func(m *notificationSettings, settings any, ctx context.Context) error {
			d, err := toSettingsStruct[clientEmail](settings)

			var elements []attr.Value
			for _, a := range d.Addresses {
				objectValue, _ := types.ObjectValueFrom(
					ctx,
					EmailAddressAttributeTypes(),
					notificationSettingsEmailAddress{
						Id:    types.StringPointerValue(a.Id),
						Email: types.StringValue(a.Email),
					},
				)
				elements = append(elements, objectValue)
			}

			setValue, d2 := types.SetValueFrom(ctx, types.ObjectType{AttrTypes: EmailAddressAttributeTypes()}, elements)
			if d2.HasError() {
				return nil
			}
			var email notificationSettingsEmail
			m.Email.As(ctx, &email, basetypes.ObjectAsOptions{})
			email.Addresses = setValue
			o, d3 := types.ObjectValueFrom(ctx, EmailAttributeTypes(), email)
			if d3.HasError() {
				return nil
			}
			m.Email = o
			return err
		},
	},
	"slack": {
		Get: func(m *notificationSettings, ctx context.Context) any {
			var slack notificationSettingsSlack
			m.Slack.As(ctx, &slack, basetypes.ObjectAsOptions{})
			return clientSlack{
				Url: slack.Url.ValueString(),
			}
		},
		Set: func(m *notificationSettings, settings any, ctx context.Context) error {
			s, err := toSettingsStruct[clientSlack](settings)
			o, _ := types.ObjectValueFrom(ctx, SlackAttributeTypes(), s)
			m.Slack = o
			return err
		},
	},
	"pagerduty": {
		Get: func(m *notificationSettings, ctx context.Context) any {
			var pagerDuty notificationSettingsPagerDuty
			m.PagerDuty.As(ctx, &pagerDuty, basetypes.ObjectAsOptions{})
			return clientPagerDuty{
				RoutingKey: pagerDuty.RoutingKey.ValueString(),
				Summary:    pagerDuty.Summary.ValueString(),
				DedupKey:   pagerDuty.DedupKey.ValueString(),
			}
		},
		Set: func(m *notificationSettings, settings any, ctx context.Context) error {
			s, err := toSettingsStruct[clientPagerDuty](settings)
			o, _ := types.ObjectValueFrom(ctx, PagerDutyAttributeTypes(), s)
			m.PagerDuty = o
			return err
		},
	},
	"webhook": {
		Get: func(m *notificationSettings, ctx context.Context) any {
			var webhook notificationSettingsWebhook
			m.Webhook.As(ctx, &webhook, basetypes.ObjectAsOptions{})
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
		Set: func(m *notificationSettings, settings any, ctx context.Context) error {
			s, err := toSettingsStruct[clientWebhook](settings)
			o, _ := types.ObjectValueFrom(ctx, WebhookAttributeTypes(), s)
			m.Webhook = o
			return err
		},
	},
	"opsgenie": {
		Get: func(m *notificationSettings, ctx context.Context) any {
			var opsGenie notificationSettingsOpsGenie
			m.OpsGenie.As(ctx, &opsGenie, basetypes.ObjectAsOptions{})
			return clientOpsGenie{
				HostName:   opsGenie.HostName.ValueString(),
				ApiKey:     opsGenie.ApiKey.ValueString(),
				Recipients: opsGenie.Recipients.ValueString(),
				Teams:      opsGenie.Teams.ValueString(),
				Tags:       opsGenie.Tags.ValueString(),
			}
		},
		Set: func(m *notificationSettings, settings any, ctx context.Context) error {
			s, err := toSettingsStruct[clientOpsGenie](settings)
			o, _ := types.ObjectValueFrom(ctx, OpsGenieAttributeTypes(), s)
			m.OpsGenie = o
			return err
		},
	},
	"amazonsns": {
		Get: func(m *notificationSettings, ctx context.Context) any {
			var amazonSns notificationSettingsAmazonSNS
			m.AmazonSNS.As(ctx, &amazonSns, basetypes.ObjectAsOptions{})
			return clientAmazonSNS{
				TopicARN:        amazonSns.TopicARN.ValueString(),
				AccessKeyID:     amazonSns.AccessKeyID.ValueString(),
				SecretAccessKey: amazonSns.SecretAccessKey.ValueString(),
			}
		},
		Set: func(m *notificationSettings, settings any, ctx context.Context) error {
			s, err := toSettingsStruct[clientAmazonSNS](settings)
			o, _ := types.ObjectValueFrom(ctx, AmazonSNSAttributeTypes(), s)
			m.AmazonSNS = o
			return err
		},
	},
	"zapier": {
		Get: func(m *notificationSettings, ctx context.Context) any {
			var zapier notificationSettingsZapier
			m.Zapier.As(ctx, &zapier, basetypes.ObjectAsOptions{})
			return clientZapier{
				Url: zapier.Url.ValueString(),
			}
		},
		Set: func(m *notificationSettings, settings any, ctx context.Context) error {
			s, err := toSettingsStruct[clientZapier](settings)
			o, _ := types.ObjectValueFrom(ctx, ZapierAttributeTypes(), s)
			m.Zapier = o
			return err
		},
	},
	"msTeams": {
		Get: func(m *notificationSettings, ctx context.Context) any {
			var msTeams notificationSettingsMsTeams
			m.MsTeams.As(ctx, &msTeams, basetypes.ObjectAsOptions{})
			return clientMsTeams{
				Url: msTeams.Url.ValueString(),
			}
		},
		Set: func(m *notificationSettings, settings any, ctx context.Context) error {
			s, err := toSettingsStruct[clientMsTeams](settings)
			o, _ := types.ObjectValueFrom(ctx, MsTeamsAttributeTypes(), s)
			m.MsTeams = o
			return err
		},
	},
	"pushover": {
		Get: func(m *notificationSettings, ctx context.Context) any {
			var pushover notificationSettingsPushover
			m.Pushover.As(ctx, &pushover, basetypes.ObjectAsOptions{})
			return clientPushover{
				UserKey:  pushover.UserKey.ValueString(),
				AppToken: pushover.AppToken.ValueString(),
			}
		},
		Set: func(m *notificationSettings, settings any, ctx context.Context) error {
			s, err := toSettingsStruct[clientPushover](settings)
			o, _ := types.ObjectValueFrom(ctx, PushoverAttributeTypes(), s)
			m.Pushover = o
			return err
		},
	},
	"swsd": {
		Get: func(m *notificationSettings, ctx context.Context) any {
			var swoServiceDesk notificationSettingsSolarWindsServiceDesk
			m.SolarWindsServiceDesk.As(ctx, &swoServiceDesk, basetypes.ObjectAsOptions{})
			return clientSolarWindsServiceDesk{
				AppToken: swoServiceDesk.AppToken.ValueString(),
				IsEU:     swoServiceDesk.IsEU.ValueBool(),
			}
		},
		Set: func(m *notificationSettings, settings any, ctx context.Context) error {
			s, err := toSettingsStruct[clientSolarWindsServiceDesk](settings)
			o, _ := types.ObjectValueFrom(ctx, SolarWindsServiceDeskAttributeTypes(), s)
			m.SolarWindsServiceDesk = o
			return err
		},
	},
	"servicenow": {
		Get: func(m *notificationSettings, ctx context.Context) any {
			var serviceNow notificationSettingsServiceNow
			m.ServiceNow.As(ctx, &serviceNow, basetypes.ObjectAsOptions{})
			return clientServiceNow{
				AppToken: serviceNow.AppToken.ValueString(),
				Instance: serviceNow.Instance.ValueString(),
			}
		},
		Set: func(m *notificationSettings, settings any, ctx context.Context) error {
			s, err := toSettingsStruct[clientServiceNow](settings)
			o, _ := types.ObjectValueFrom(ctx, ServiceNowAttributeTypes(), s)
			m.ServiceNow = o
			return err
		},
	},
}

// Utility function to marshal anonymous json to concrete models defined by T.
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

func (m *notificationResourceModel) SetSettings(settings *any, ctx context.Context) error {

	if accessor, found := settingsAccessors[m.Type.ValueString()]; found {
		if m.Settings.IsNull() {
			m.Settings = types.ObjectNull(NotificationSettingsAttributeTypes())
		}

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

		err := accessor.Set(&model, settings, ctx)
		if err != nil {
			return err
		}

		o, _ := types.ObjectValueFrom(ctx, NotificationSettingsAttributeTypes(), model)
		m.Settings = o

	} else {
		return newUnsupportedNotificationTypeError(m.Type.ValueString())
	}

	return nil
}

func (m *notificationResourceModel) GetSettings(ctx context.Context) any {

	if accessor, found := settingsAccessors[m.Type.ValueString()]; found {
		if m.Settings.IsNull() {
			m.Settings = types.ObjectNull(NotificationSettingsAttributeTypes())
			return nil
		}

		var settings notificationSettings
		m.Settings.As(ctx, &settings, basetypes.ObjectAsOptions{})
		return accessor.Get(&settings, ctx)
	}

	log.Printf("unsupported notification type. got %s", m.Type.ValueString())
	return nil
}
