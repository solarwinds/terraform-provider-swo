package provider

import (
	"encoding/json"
	"fmt"
	"log"
)

type NotificationSettings struct {
	Email                 *NotificationSettingsEmail                 `tfsdk:"email"`
	Slack                 *NotificationSettingsSlack                 `tfsdk:"slack"`
	PagerDuty             *NotificationSettingsPagerDuty             `tfsdk:"pagerduty"`
	MsTeams               *NotificationSettingsMsTeams               `tfsdk:"msteams"`
	Webhook               *NotificationSettingsWebhook               `tfsdk:"webhook"`
	VictorOps             *NotificationSettingsVictorOps             `tfsdk:"victorops"`
	OpsGenie              *NotificationSettingsOpsGenie              `tfsdk:"opsgenie"`
	AmazonSNS             *NotificationSettingsAmazonSNS             `tfsdk:"amazonsns"`
	Zapier                *NotificationSettingsZapier                `tfsdk:"zapier"`
	Pushover              *NotificationSettingsPushover              `tfsdk:"pushover"`
	Sms                   *NotificationSettingsSms                   `tfsdk:"sms"`
	SolarWindsServiceDesk *NotificationSettingsSolarWindsServiceDesk `tfsdk:"swsd"`
	ServiceNow            *NotificationSettingsServiceNow            `tfsdk:"servicenow"`
}

type NotificationSettingsEmail struct {
	Addresses []NotificationSettingsEmailAddress `tfsdk:"addresses" json:"addresses"`
}

type NotificationSettingsEmailAddress struct {
	Id    *string `tfsdk:"id" json:"id"`
	Email string  `tfsdk:"email" json:"email"`
}

type NotificationSettingsOpsGenie struct {
	HostName   string `tfsdk:"hostname" json:"hostname"`
	ApiKey     string `tfsdk:"api_key" json:"apiKey"`
	Recipients string `tfsdk:"recipients" json:"recipients"`
	Teams      string `tfsdk:"teams" json:"teams"`
	Tags       string `tfsdk:"tags" json:"tags"`
}

type NotificationSettingsSms struct {
	PhoneNumbers string `tfsdk:"phone_numbers" json:"phoneNumbers"`
}

type NotificationSettingsSlack struct {
	Url string `tfsdk:"url" json:"url"`
}

type NotificationSettingsMsTeams struct {
	Url string `tfsdk:"url" json:"url"`
}

type NotificationSettingsPagerDuty struct {
	RoutingKey string `tfsdk:"routing_key" json:"routingKey"`
	Summary    string `tfsdk:"summary" json:"summary"`
	DedupKey   string `tfsdk:"dedup_key" json:"dedupKey"`
}

type NotificationSettingsWebhook struct {
	Url             string `tfsdk:"url" json:"url"`
	Method          string `tfsdk:"method" json:"method"`
	AuthType        string `tfsdk:"auth_type" json:"authType"`
	AuthUsername    string `tfsdk:"auth_username" json:"authUsername"`
	AuthPassword    string `tfsdk:"auth_password" json:"authPassword"`
	AuthHeaderName  string `tfsdk:"auth_header_name" json:"authHeaderName"`
	AuthHeaderValue string `tfsdk:"auth_header_value" json:"authHeaderValue"`
}

type NotificationSettingsVictorOps struct {
	ApiKey     string `tfsdk:"api_key" json:"apiKey"`
	RoutingKey string `tfsdk:"routing_key" json:"routingKey"`
}

type NotificationSettingsAmazonSNS struct {
	TopicARN        string `tfsdk:"topic_arn" json:"topicARN"`
	AccessKeyID     string `tfsdk:"access_key_id" json:"accessKeyID"`
	SecretAccessKey string `tfsdk:"secret_access_key" json:"secretAccessKey"`
}

type NotificationSettingsZapier struct {
	Url string `tfsdk:"url" json:"url"`
}

type NotificationSettingsPushover struct {
	UserKey  string `tfsdk:"user_key" json:"userKey"`
	AppToken string `tfsdk:"app_token" json:"appToken"`
}

type NotificationSettingsSolarWindsServiceDesk struct {
	AppToken string `tfsdk:"app_token" json:"appToken"`
	IsEU     bool   `tfsdk:"is_eu" json:"isEu"`
}

type NotificationSettingsServiceNow struct {
	AppToken string `tfsdk:"app_token" json:"appToken"`
	Instance string `tfsdk:"instance" json:"instance"`
}

type NotificationSettingsAccessor struct {
	Get func(m *NotificationResourceModel) any
	Set func(m *NotificationResourceModel, settings any) error
}

var settingsAccessors = map[string]NotificationSettingsAccessor{
	"email": {
		Get: func(m *NotificationResourceModel) any {
			return m.Settings.Email
		},
		Set: func(m *NotificationResourceModel, settings any) error {
			s, err := toSettingsStruct[NotificationSettingsEmail](settings)
			m.Settings.Email = s
			return err
		},
	},
	"slack": {
		Get: func(m *NotificationResourceModel) any {
			return m.Settings.Slack
		},
		Set: func(m *NotificationResourceModel, settings any) error {
			s, err := toSettingsStruct[NotificationSettingsSlack](settings)
			m.Settings.Slack = s
			return err
		},
	},
	"pagerduty": {
		Get: func(m *NotificationResourceModel) any {
			return m.Settings.PagerDuty
		},
		Set: func(m *NotificationResourceModel, settings any) error {
			s, err := toSettingsStruct[NotificationSettingsPagerDuty](settings)
			m.Settings.PagerDuty = s
			return err
		},
	},
	"webhook": {
		Get: func(m *NotificationResourceModel) any {
			return m.Settings.Webhook
		},
		Set: func(m *NotificationResourceModel, settings any) error {
			s, err := toSettingsStruct[NotificationSettingsWebhook](settings)
			m.Settings.Webhook = s
			return err
		},
	},
	"victorops": {
		Get: func(m *NotificationResourceModel) any {
			return m.Settings.VictorOps
		},
		Set: func(m *NotificationResourceModel, settings any) error {
			s, err := toSettingsStruct[NotificationSettingsVictorOps](settings)
			m.Settings.VictorOps = s
			return err
		},
	},
	"opsgenie": {
		Get: func(m *NotificationResourceModel) any {
			return m.Settings.OpsGenie
		},
		Set: func(m *NotificationResourceModel, settings any) error {
			s, err := toSettingsStruct[NotificationSettingsOpsGenie](settings)
			m.Settings.OpsGenie = s
			return err
		},
	},
	"amazonsns": {
		Get: func(m *NotificationResourceModel) any {
			return m.Settings.AmazonSNS
		},
		Set: func(m *NotificationResourceModel, settings any) error {
			s, err := toSettingsStruct[NotificationSettingsAmazonSNS](settings)
			m.Settings.AmazonSNS = s
			return err
		},
	},
	"zapier": {
		Get: func(m *NotificationResourceModel) any {
			return m.Settings.Zapier
		},
		Set: func(m *NotificationResourceModel, settings any) error {
			s, err := toSettingsStruct[NotificationSettingsZapier](settings)
			m.Settings.Zapier = s
			return err
		},
	},
	"msTeams": {
		Get: func(m *NotificationResourceModel) any {
			return m.Settings.MsTeams
		},
		Set: func(m *NotificationResourceModel, settings any) error {
			s, err := toSettingsStruct[NotificationSettingsMsTeams](settings)
			m.Settings.MsTeams = s
			return err
		},
	},
	"pushover": {
		Get: func(m *NotificationResourceModel) any {
			return m.Settings.Pushover
		},
		Set: func(m *NotificationResourceModel, settings any) error {
			s, err := toSettingsStruct[NotificationSettingsPushover](settings)
			m.Settings.Pushover = s
			return err
		},
	},
	"sms": {
		Get: func(m *NotificationResourceModel) any {
			return m.Settings.Sms
		},
		Set: func(m *NotificationResourceModel, settings any) error {
			s, err := toSettingsStruct[NotificationSettingsSms](settings)
			m.Settings.Sms = s
			return err
		},
	},
	"swsd": {
		Get: func(m *NotificationResourceModel) any {
			return m.Settings.SolarWindsServiceDesk
		},
		Set: func(m *NotificationResourceModel, settings any) error {
			s, err := toSettingsStruct[NotificationSettingsSolarWindsServiceDesk](settings)
			m.Settings.SolarWindsServiceDesk = s
			return err
		},
	},
	"servicenow": {
		Get: func(m *NotificationResourceModel) any {
			return m.Settings.ServiceNow
		},
		Set: func(m *NotificationResourceModel, settings any) error {
			s, err := toSettingsStruct[NotificationSettingsServiceNow](settings)
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

func (m *NotificationResourceModel) SetSettings(settings any) error {
	if accessor, found := settingsAccessors[m.Type.ValueString()]; found {
		if m.Settings == nil {
			m.Settings = &NotificationSettings{}
		}
		err := accessor.Set(m, settings)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("unsupported notification type: %s", m.Type.ValueString())
	}

	return nil
}

func (m *NotificationResourceModel) GetSettings() any {
	if accessor, found := settingsAccessors[m.Type.ValueString()]; found {
		if m.Settings == nil {
			m.Settings = &NotificationSettings{}
		}
		return accessor.Get(m)
	}

	log.Printf("unsupported notification type. got %s", m.Type.ValueString())
	return nil
}
