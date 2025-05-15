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
	VictorOps             types.Object `tfsdk:"victorops"`
	OpsGenie              types.Object `tfsdk:"opsgenie"`
	AmazonSNS             types.Object `tfsdk:"amazonsns"`
	Zapier                types.Object `tfsdk:"zapier"`
	Pushover              types.Object `tfsdk:"pushover"`
	Sms                   types.Object `tfsdk:"sms"`
	SolarWindsServiceDesk types.Object `tfsdk:"swsd"`
	ServiceNow            types.Object `tfsdk:"servicenow"`
}

func NotificationSettingsAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"port":             types.Int64Type,
		"string_to_expect": types.StringType,
		"string_to_send":   types.StringType,
	}
}

type notificationSettingsEmail struct {
	Addresses types.Set `tfsdk:"addresses" json:"addresses"`
}

type clientEmail struct {
	Addresses []clientEmailAddress `tfsdk:"addresses" json:"addresses"`
}

type notificationSettingsEmailAddress struct {
	Id    types.String `tfsdk:"id" json:"id"`
	Email types.String `tfsdk:"email" json:"email"`
}

type clientEmailAddress struct {
	Id    *string `tfsdk:"id" json:"id"`
	Email string  `tfsdk:"email" json:"email"`
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

type notificationSettingsSms struct {
	PhoneNumbers types.String `tfsdk:"phone_numbers" json:"phoneNumbers"`
}

type clientSms struct {
	PhoneNumbers string `tfsdk:"phone_numbers" json:"phoneNumbers"`
}

type notificationSettingsSlack struct {
	Url types.String `tfsdk:"url" json:"url"`
}

type clientSlack struct {
	Url string `tfsdk:"url" json:"url"`
}

type notificationSettingsMsTeams struct {
	Url types.String `tfsdk:"url" json:"url"`
}

type clientMsTeams struct {
	Url string `tfsdk:"url" json:"url"`
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

type notificationSettingsVictorOps struct {
	ApiKey     types.String `tfsdk:"api_key" json:"apiKey"`
	RoutingKey types.String `tfsdk:"routing_key" json:"routingKey"`
}

type clientVictorOps struct {
	ApiKey     string `tfsdk:"api_key" json:"apiKey"`
	RoutingKey string `tfsdk:"routing_key" json:"routingKey"`
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

type notificationSettingsZapier struct {
	Url types.String `tfsdk:"url" json:"url"`
}

type clientZapier struct {
	Url string `tfsdk:"url" json:"url"`
}

type notificationSettingsPushover struct {
	UserKey  types.String `tfsdk:"user_key" json:"userKey"`
	AppToken types.String `tfsdk:"app_token" json:"appToken"`
}

type clientPushover struct {
	UserKey  string `tfsdk:"user_key" json:"userKey"`
	AppToken string `tfsdk:"app_token" json:"appToken"`
}

type notificationSettingsSolarWindsServiceDesk struct {
	AppToken types.String `tfsdk:"app_token" json:"appToken"`
	IsEU     types.Bool   `tfsdk:"is_eu" json:"isEu"`
}

type clientSolarWindsServiceDesk struct {
	AppToken string `tfsdk:"app_token" json:"appToken"`
	IsEU     bool   `tfsdk:"is_eu" json:"isEu"`
}

type notificationSettingsServiceNow struct {
	AppToken types.String `tfsdk:"app_token" json:"appToken"`
	Instance types.String `tfsdk:"instance" json:"instance"`
}

type clientServiceNow struct {
	AppToken string `tfsdk:"app_token" json:"appToken"`
	Instance string `tfsdk:"instance" json:"instance"`
}

type notificationSettingsAccessor struct {
	Get func(m *notificationSettings, ctx context.Context) any
	Set func(m *notificationResourceModel, settings any) error
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
			var e = clientEmail{
				Addresses: c,
			}

			return e
		},
		Set: func(m *notificationResourceModel, settings any) error {
			_, err := toSettingsStruct[notificationSettingsEmail](settings)

			//tfTcpOptions, d := types.ObjectValueFrom(ctx, UriTcpOptionsAttributeTypes(), tcpElement)
			//m.Settings.Email = s
			return err
		},
	},
	"slack": {
		Get: func(m *notificationSettings, ctx context.Context) any {
			var slack notificationSettingsSlack
			m.Slack.As(ctx, &slack, basetypes.ObjectAsOptions{})
			c := clientSlack{
				Url: slack.Url.ValueString(),
			}
			return c
		},
		Set: func(m *notificationResourceModel, settings any) error {
			_, err := toSettingsStruct[notificationSettingsSlack](settings)
			//m.Settings.Slack = s
			return err
		},
	},
	"pagerduty": {
		Get: func(m *notificationSettings, ctx context.Context) any {
			var pagerDuty notificationSettingsPagerDuty
			m.PagerDuty.As(ctx, &pagerDuty, basetypes.ObjectAsOptions{})
			c := clientPagerDuty{
				RoutingKey: pagerDuty.RoutingKey.ValueString(),
				Summary:    pagerDuty.Summary.ValueString(),
				DedupKey:   pagerDuty.DedupKey.ValueString(),
			}
			return c
		},
		Set: func(m *notificationResourceModel, settings any) error {
			_, err := toSettingsStruct[notificationSettingsPagerDuty](settings)
			//m.Settings.PagerDuty = s
			return err
		},
	},
	"webhook": {
		Get: func(m *notificationSettings, ctx context.Context) any {
			var webhook notificationSettingsWebhook
			m.Webhook.As(ctx, &webhook, basetypes.ObjectAsOptions{})
			c := clientWebhook{
				Url:             webhook.Url.ValueString(),
				Method:          webhook.Method.ValueString(),
				AuthType:        webhook.AuthType.ValueString(),
				AuthUsername:    webhook.AuthUsername.ValueString(),
				AuthPassword:    webhook.AuthPassword.ValueString(),
				AuthHeaderName:  webhook.AuthHeaderName.ValueString(),
				AuthHeaderValue: webhook.AuthHeaderValue.ValueString(),
			}
			return c
		},
		Set: func(m *notificationResourceModel, settings any) error {
			_, err := toSettingsStruct[notificationSettingsWebhook](settings)
			//m.Settings.Webhook = s
			return err
		},
	},
	"victorops": {
		Get: func(m *notificationSettings, ctx context.Context) any {
			var victorOps notificationSettingsVictorOps
			m.VictorOps.As(ctx, &victorOps, basetypes.ObjectAsOptions{})
			c := clientVictorOps{
				ApiKey:     victorOps.ApiKey.ValueString(),
				RoutingKey: victorOps.RoutingKey.ValueString(),
			}
			return c
		},
		Set: func(m *notificationResourceModel, settings any) error {
			_, err := toSettingsStruct[notificationSettingsVictorOps](settings)
			//m.Settings.VictorOps = s
			return err
		},
	},
	"opsgenie": {
		Get: func(m *notificationSettings, ctx context.Context) any {
			var opsGenie notificationSettingsOpsGenie
			m.OpsGenie.As(ctx, &opsGenie, basetypes.ObjectAsOptions{})
			c := clientOpsGenie{
				HostName:   opsGenie.HostName.ValueString(),
				ApiKey:     opsGenie.ApiKey.ValueString(),
				Recipients: opsGenie.Recipients.ValueString(),
				Teams:      opsGenie.Teams.ValueString(),
				Tags:       opsGenie.Tags.ValueString(),
			}
			return c
		},
		Set: func(m *notificationResourceModel, settings any) error {
			_, err := toSettingsStruct[notificationSettingsOpsGenie](settings)

			//m.Settings.OpsGenie = s
			return err
		},
	},
	"amazonsns": {
		Get: func(m *notificationSettings, ctx context.Context) any {
			var amazonSns notificationSettingsAmazonSNS
			m.AmazonSNS.As(ctx, &amazonSns, basetypes.ObjectAsOptions{})
			c := clientAmazonSNS{
				TopicARN:        amazonSns.TopicARN.ValueString(),
				AccessKeyID:     amazonSns.AccessKeyID.ValueString(),
				SecretAccessKey: amazonSns.SecretAccessKey.ValueString(),
			}
			return c
		},
		Set: func(m *notificationResourceModel, settings any) error {
			_, err := toSettingsStruct[notificationSettingsAmazonSNS](settings)
			//m.Settings.AmazonSNS = s
			return err
		},
	},
	"zapier": {
		Get: func(m *notificationSettings, ctx context.Context) any {
			var zapier notificationSettingsZapier
			m.Zapier.As(ctx, &zapier, basetypes.ObjectAsOptions{})
			c := clientZapier{
				Url: zapier.Url.ValueString(),
			}
			return c
		},
		Set: func(m *notificationResourceModel, settings any) error {
			_, err := toSettingsStruct[notificationSettingsZapier](settings)
			//m.Settings.Zapier = s
			return err
		},
	},
	"msTeams": {
		Get: func(m *notificationSettings, ctx context.Context) any {
			var msTeams notificationSettingsMsTeams
			m.MsTeams.As(ctx, &msTeams, basetypes.ObjectAsOptions{})
			c := clientMsTeams{
				Url: msTeams.Url.ValueString(),
			}
			return c
		},
		Set: func(m *notificationResourceModel, settings any) error {
			_, err := toSettingsStruct[notificationSettingsMsTeams](settings)
			//m.Settings.MsTeams = s
			return err
		},
	},
	"pushover": {
		Get: func(m *notificationSettings, ctx context.Context) any {
			var pushover notificationSettingsPushover
			m.Pushover.As(ctx, &pushover, basetypes.ObjectAsOptions{})
			c := clientPushover{
				UserKey:  pushover.UserKey.ValueString(),
				AppToken: pushover.AppToken.ValueString(),
			}
			return c
		},
		Set: func(m *notificationResourceModel, settings any) error {
			_, err := toSettingsStruct[notificationSettingsPushover](settings)
			//m.Settings.Pushover = s
			return err
		},
	},
	"sms": {
		Get: func(m *notificationSettings, ctx context.Context) any {
			var sms notificationSettingsSms
			m.Sms.As(ctx, &sms, basetypes.ObjectAsOptions{})
			c := clientSms{
				PhoneNumbers: sms.PhoneNumbers.ValueString(),
			}
			return c
		},
		Set: func(m *notificationResourceModel, settings any) error {
			_, err := toSettingsStruct[notificationSettingsSms](settings)
			//m.Settings.Sms = s
			return err
		},
	},
	"swsd": {
		Get: func(m *notificationSettings, ctx context.Context) any {
			var swoServiceDesk notificationSettingsSolarWindsServiceDesk
			m.SolarWindsServiceDesk.As(ctx, &swoServiceDesk, basetypes.ObjectAsOptions{})
			c := clientSolarWindsServiceDesk{
				AppToken: swoServiceDesk.AppToken.ValueString(),
				IsEU:     swoServiceDesk.IsEU.ValueBool(),
			}
			return c
		},
		Set: func(m *notificationResourceModel, settings any) error {
			_, err := toSettingsStruct[notificationSettingsSolarWindsServiceDesk](settings)
			//m.Settings.SolarWindsServiceDesk = s
			return err
		},
	},
	"servicenow": {
		Get: func(m *notificationSettings, ctx context.Context) any {
			var serviceNow notificationSettingsServiceNow
			m.ServiceNow.As(ctx, &serviceNow, basetypes.ObjectAsOptions{})
			c := clientServiceNow{
				AppToken: serviceNow.AppToken.ValueString(),
				Instance: serviceNow.Instance.ValueString(),
			}
			return c
		},
		Set: func(m *notificationResourceModel, settings any) error {
			_, err := toSettingsStruct[notificationSettingsServiceNow](settings)
			//m.Settings.ServiceNow = s
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

func (m *notificationResourceModel) SetSettings(settings any) error {

	if accessor, found := settingsAccessors[m.Type.ValueString()]; found {
		if m.Settings.IsNull() {
			m.Settings = types.ObjectNull(NotificationSettingsAttributeTypes())
		}
		err := accessor.Set(m, settings)
		if err != nil {
			return err
		}
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
