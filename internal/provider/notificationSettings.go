package provider

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
)

var (
	unsupportedNotificationTypeError = errors.New("unsupported notification type")
)

func newUnsupportedNotificationTypeError(notificationType string) error {
	return fmt.Errorf("%w: %s", unsupportedNotificationTypeError, notificationType)
}

type notificationSettings struct {
	Email                 *notificationSettingsEmail                 `tfsdk:"email"`
	Slack                 *notificationSettingsSlack                 `tfsdk:"slack"`
	PagerDuty             *notificationSettingsPagerDuty             `tfsdk:"pagerduty"`
	MsTeams               *notificationSettingsMsTeams               `tfsdk:"msteams"`
	Webhook               *notificationSettingsWebhook               `tfsdk:"webhook"`
	VictorOps             *notificationSettingsVictorOps             `tfsdk:"victorops"`
	OpsGenie              *notificationSettingsOpsGenie              `tfsdk:"opsgenie"`
	AmazonSNS             *notificationSettingsAmazonSNS             `tfsdk:"amazonsns"`
	Zapier                *notificationSettingsZapier                `tfsdk:"zapier"`
	Pushover              *notificationSettingsPushover              `tfsdk:"pushover"`
	Sms                   *notificationSettingsSms                   `tfsdk:"sms"`
	SolarWindsServiceDesk *notificationSettingsSolarWindsServiceDesk `tfsdk:"swsd"`
	ServiceNow            *notificationSettingsServiceNow            `tfsdk:"servicenow"`
}

type notificationSettingsEmail struct {
	Addresses []notificationSettingsEmailAddress `tfsdk:"addresses" json:"addresses"`
}

type notificationSettingsEmailAddress struct {
	Id    *string `tfsdk:"id" json:"id"`
	Email string  `tfsdk:"email" json:"email"`
}

type notificationSettingsOpsGenie struct {
	HostName   string `tfsdk:"hostname" json:"hostname"`
	ApiKey     string `tfsdk:"api_key" json:"apiKey"`
	Recipients string `tfsdk:"recipients" json:"recipients"`
	Teams      string `tfsdk:"teams" json:"teams"`
	Tags       string `tfsdk:"tags" json:"tags"`
}

type notificationSettingsSms struct {
	PhoneNumbers string `tfsdk:"phone_numbers" json:"phoneNumbers"`
}

type notificationSettingsSlack struct {
	Url string `tfsdk:"url" json:"url"`
}

type notificationSettingsMsTeams struct {
	Url string `tfsdk:"url" json:"url"`
}

type notificationSettingsPagerDuty struct {
	RoutingKey string `tfsdk:"routing_key" json:"routingKey"`
	Summary    string `tfsdk:"summary" json:"summary"`
	DedupKey   string `tfsdk:"dedup_key" json:"dedupKey"`
}

type notificationSettingsWebhook struct {
	Url             string `tfsdk:"url" json:"url"`
	Method          string `tfsdk:"method" json:"method"`
	AuthType        string `tfsdk:"auth_type" json:"authType"`
	AuthUsername    string `tfsdk:"auth_username" json:"authUsername"`
	AuthPassword    string `tfsdk:"auth_password" json:"authPassword"`
	AuthHeaderName  string `tfsdk:"auth_header_name" json:"authHeaderName"`
	AuthHeaderValue string `tfsdk:"auth_header_value" json:"authHeaderValue"`
}

type notificationSettingsVictorOps struct {
	ApiKey     string `tfsdk:"api_key" json:"apiKey"`
	RoutingKey string `tfsdk:"routing_key" json:"routingKey"`
}

type notificationSettingsAmazonSNS struct {
	TopicARN        string `tfsdk:"topic_arn" json:"topicARN"`
	AccessKeyID     string `tfsdk:"access_key_id" json:"accessKeyID"`
	SecretAccessKey string `tfsdk:"secret_access_key" json:"secretAccessKey"`
}

type notificationSettingsZapier struct {
	Url string `tfsdk:"url" json:"url"`
}

type notificationSettingsPushover struct {
	UserKey  string `tfsdk:"user_key" json:"userKey"`
	AppToken string `tfsdk:"app_token" json:"appToken"`
}

type notificationSettingsSolarWindsServiceDesk struct {
	AppToken string `tfsdk:"app_token" json:"appToken"`
	IsEU     bool   `tfsdk:"is_eu" json:"isEu"`
}

type notificationSettingsServiceNow struct {
	AppToken string `tfsdk:"app_token" json:"appToken"`
	Instance string `tfsdk:"instance" json:"instance"`
}

type notificationSettingsAccessor struct {
	Get func(m *notificationResourceModel) any
	Set func(m *notificationResourceModel, settings any) error
}

var settingsAccessors = map[string]notificationSettingsAccessor{
	"email": {
		Get: func(m *notificationResourceModel) any {
			return m.Settings.Email
		},
		Set: func(m *notificationResourceModel, settings any) error {
			s, err := toSettingsStruct[notificationSettingsEmail](settings)
			m.Settings.Email = s
			return err
		},
	},
	"slack": {
		Get: func(m *notificationResourceModel) any {
			return m.Settings.Slack
		},
		Set: func(m *notificationResourceModel, settings any) error {
			s, err := toSettingsStruct[notificationSettingsSlack](settings)
			m.Settings.Slack = s
			return err
		},
	},
	"pagerduty": {
		Get: func(m *notificationResourceModel) any {
			return m.Settings.PagerDuty
		},
		Set: func(m *notificationResourceModel, settings any) error {
			s, err := toSettingsStruct[notificationSettingsPagerDuty](settings)
			m.Settings.PagerDuty = s
			return err
		},
	},
	"webhook": {
		Get: func(m *notificationResourceModel) any {
			return m.Settings.Webhook
		},
		Set: func(m *notificationResourceModel, settings any) error {
			s, err := toSettingsStruct[notificationSettingsWebhook](settings)
			m.Settings.Webhook = s
			return err
		},
	},
	"victorops": {
		Get: func(m *notificationResourceModel) any {
			return m.Settings.VictorOps
		},
		Set: func(m *notificationResourceModel, settings any) error {
			s, err := toSettingsStruct[notificationSettingsVictorOps](settings)
			m.Settings.VictorOps = s
			return err
		},
	},
	"opsgenie": {
		Get: func(m *notificationResourceModel) any {
			return m.Settings.OpsGenie
		},
		Set: func(m *notificationResourceModel, settings any) error {
			s, err := toSettingsStruct[notificationSettingsOpsGenie](settings)
			m.Settings.OpsGenie = s
			return err
		},
	},
	"amazonsns": {
		Get: func(m *notificationResourceModel) any {
			return m.Settings.AmazonSNS
		},
		Set: func(m *notificationResourceModel, settings any) error {
			s, err := toSettingsStruct[notificationSettingsAmazonSNS](settings)
			m.Settings.AmazonSNS = s
			return err
		},
	},
	"zapier": {
		Get: func(m *notificationResourceModel) any {
			return m.Settings.Zapier
		},
		Set: func(m *notificationResourceModel, settings any) error {
			s, err := toSettingsStruct[notificationSettingsZapier](settings)
			m.Settings.Zapier = s
			return err
		},
	},
	"msTeams": {
		Get: func(m *notificationResourceModel) any {
			return m.Settings.MsTeams
		},
		Set: func(m *notificationResourceModel, settings any) error {
			s, err := toSettingsStruct[notificationSettingsMsTeams](settings)
			m.Settings.MsTeams = s
			return err
		},
	},
	"pushover": {
		Get: func(m *notificationResourceModel) any {
			return m.Settings.Pushover
		},
		Set: func(m *notificationResourceModel, settings any) error {
			s, err := toSettingsStruct[notificationSettingsPushover](settings)
			m.Settings.Pushover = s
			return err
		},
	},
	"sms": {
		Get: func(m *notificationResourceModel) any {
			return m.Settings.Sms
		},
		Set: func(m *notificationResourceModel, settings any) error {
			s, err := toSettingsStruct[notificationSettingsSms](settings)
			m.Settings.Sms = s
			return err
		},
	},
	"swsd": {
		Get: func(m *notificationResourceModel) any {
			return m.Settings.SolarWindsServiceDesk
		},
		Set: func(m *notificationResourceModel, settings any) error {
			s, err := toSettingsStruct[notificationSettingsSolarWindsServiceDesk](settings)
			m.Settings.SolarWindsServiceDesk = s
			return err
		},
	},
	"servicenow": {
		Get: func(m *notificationResourceModel) any {
			return m.Settings.ServiceNow
		},
		Set: func(m *notificationResourceModel, settings any) error {
			s, err := toSettingsStruct[notificationSettingsServiceNow](settings)
			m.Settings.ServiceNow = s
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
		if m.Settings == nil {
			m.Settings = &notificationSettings{}
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

func (m *notificationResourceModel) GetSettings() any {
	if accessor, found := settingsAccessors[m.Type.ValueString()]; found {
		if m.Settings == nil {
			m.Settings = &notificationSettings{}
		}
		return accessor.Get(m)
	}

	log.Printf("unsupported notification type. got %s", m.Type.ValueString())
	return nil
}
