resource "swo_notification" "amazonsns" {
  title       = "Amazon SNS notification"
  description = "This is a description"
  type        = "amazonsns"
  settings = {
    amazonsns = {
      topic_arn         = "arn:aws:sns:us-east-1:123456789012:topic"
      access_key_id     = "KEY_ID"
      secret_access_key = "SECRET_KEY"
    }
  }
}

resource "swo_notification" "email" {
  title       = "Email notification"
  description = "This is a description"
  type        = "email"
  settings = {
    email = {
      addresses = [
        {
          email = "test1@host.com"
        },
        {
          email = "test2@host.com"
        },
      ]
    }
  }
}

resource "swo_notification" "msteams" {
  title       = "Microsoft Teams notification"
  description = "This is a description"
  type        = "msTeams"
  settings = {
    msteams = {
      url = "https://XXX.webhook.office.com/webhookb2/XXXXX"
    }
  }
}

resource "swo_notification" "opsgenie" {
  title       = "OpsGenie notification"
  description = "This is a description"
  type        = "opsgenie"
  settings = {
    opsgenie = {
      hostname   = "hostname"
      apikey     = "API_KEY"
      recipients = "on-call recipient"
      teams      = "team1, team2"
      tags       = "tag1, tag2"
    }
  }
}

resource "swo_notification" "pagerduty" {
  title       = "PagerDuty notification"
  description = "This is a description"
  type        = "pagerduty"
  settings = {
    pagerduty = {
      routing_key = "99999999999999999999999999999999"
      summary     = "some-summary"
      dedup_key   = "DEDUP_KEY"
    }
  }
}

resource "swo_notification" "pushover" {
  title       = "Pushover notification"
  description = "This is a description"
  type        = "pushover"
  settings = {
    pushover = {
      app_token = "APP_TOKEN"
      user_key  = "123xyz"
    }
  }
}

resource "swo_notification" "servicenow" {
  title       = "ServiceNow notification"
  description = "This is a description"
  type        = "servicenow"
  settings = {
    servicenow = {
      app_token = "APP_TOKEN"
      instance  = "US"
    }
  }
}

resource "swo_notification" "slack" {
  title       = "Slack notification"
  description = "This is a description"
  type        = "slack"
  settings = {
    slack = {
      url = "https://hooks.slack.com/services/XXX/XXX/XXX"
    }
  }
}

resource "swo_notification" "swsd" {
  title       = "SolarWinds Service Desk notification"
  description = "This is a description"
  type        = "swsd"
  settings = {
    swsd = {
      app_token = "APP_TOKEN"
      is_eu     = false
    }
  }
}

resource "swo_notification" "test_webhook" {
  title       = "Webhook notification"
  description = "This is a description"
  type        = "webhook"
  settings = {
    webhook = {
      method            = "GET"
      url               = "https://webhook.example.com/"
      auth_header_name  = "X-Request-Id"
      auth_header_value = "HEADER_VALUE"
      auth_password     = "AUTH_PASSWORD"
      auth_type         = "basic"
      auth_username     = "AUTH_USERNAME"
    }
  }
}

resource "swo_notification" "test_zapier" {
  title       = "Zapier notification"
  description = "This is a description"
  type        = "zapier"
  settings = {
    zapier = {
      url = "https://hooks.zapier.com/hooks/catch/XXX"
    }
  }
}