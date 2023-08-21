resource "swo_notification" "msteams" {
  title       = "Microsoft teams notification"
  description = "testing..."
  type        = "msTeams"
  settings = {
    msteams = {
      url = "https://www.office.com/webhook"
    }
  }
}
resource "swo_notification" "opsgenie" {
  title       = "OpsGenie notification"
  description = "testing..."
  type        = "opsgenie"
  settings = {
    opsgenie = {
      hostname   = "hostname"
      apikey     = "123xyz"
      recipients = "alice"
      teams      = "team1, team2"
      tags       = "tag1, tag2"
    }
  }
}
resource "swo_notification" "slack" {
  title       = "Slack notification"
  description = "testing..."
  type        = "slack"
  settings = {
    slack = {
      url = "https://hooks.slack.com/services/XXX/XXX/XXX"
    }
  }
}
resource "swo_notification" "pagerduty" {
  title       = "PagerDuty notification"
  description = "testing..."
  type        = "pagerduty"
  settings = {
    pagerduty = {
      routing_key = "99999999999999999999999999999999"
      summary     = "summary"
      dedup_key   = "dedup"
    }
  }
}
resource "swo_notification" "victorops" {
  title       = "VictorOps notification"
  description = "testing..."
  type        = "victorops"
  settings = {
    victorops = {
      api_key     = "xyz"
      routing_key = "123"
    }
  }
}
resource "swo_notification" "sms" {
  title       = "SMS notification"
  description = "testing..."
  type        = "sms"
  settings = {
    sms = {
      phone_numbers = "+1 999 999 9999"
    }
  }
}
resource "swo_notification" "servicenow" {
  title       = "ServiceNow notification"
  description = "testing..."
  type        = "servicenow"
  settings = {
    servicenow = {
      app_token = "xyz"
      instance  = "US"
    }
  }
}
resource "swo_notification" "swsd" {
  title       = "SolarWinds Service Desk notification"
  description = "testing..."
  type        = "swsd"
  settings = {
    swsd = {
      app_token = "xyz"
      is_eu     = false
    }
  }
}

resource "swo_notification" "email" {
  title       = "Email notification"
  description = "testing..."
  type        = "email"
  settings = {
    email = {
      addresses = [
        {
          email = "bob@xyz.com"
        },
      ]
    }
  }
}
