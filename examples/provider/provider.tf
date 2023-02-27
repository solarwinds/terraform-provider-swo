terraform {
  required_providers {
    swo = {
      version = "0.1.0"
      source  = "github.com/solarwindscloud/swo"
    }
  }
}

provider "swo" {
  api_token = "[UPDATE WITH SWO TOKEN]"
  request_timeout = 30
  debug_mode = true
}

# resource "swo_notification" "test_email" {
#   title = "RobS Email Test"
#   description = "testing..."
#   type = "email"
#   settings = {
#     email = {
#       addresses = [
#         {
#           email = "r@t.gov"
#         },
#         {
#           email = "rob.stovenour@noop.com"
#         },
#       ]
#     }
#   }
# }

# resource "swo_notification" "test_msteams" {
#   title = "RobS MS Teams Test"
#   description = "testing..."
#   type = "msTeams"
#   settings = {
#     msteams = {
#       url = "https://www.office.com/webhook"
#     }
#   }
# }

# resource "swo_notification" "test_opsgenie" {
#   title = "RobS OpsGenie Test"
#   description = "testing..."
#   type = "opsgenie"
#   settings = {
#     opsgenie = {
#       hostname = "hostname"
#       apikey = "123xyz"
#       recipients = "robstove"
#       teams = "team1, team2"
#       tags = "tag1, tag2"
#     }
#   }
# }

# resource "swo_notification" "test_slack" {
#   title = "RobS Slack Test"
#   description = "testing..."
#   type = "slack"
#   settings = {
#     slack = {
#       url = "https://hooks.slack.com/services/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"
#     }
#   }
# }

# resource "swo_notification" "test_pagerduty" {
#   title = "RobS PagerDuty Test"
#   description = "testing..."
#   type = "pagerduty"
#   settings = {
#     pagerduty = {
#       routing_key = "99999999999999999999999999999999"
#       summary = "summary"
#       dedup_key = "dedup"
#     }
#   }
# }

# resource "swo_notification" "test_victorops" {
#   title = "RobS VictorOps Test"
#   description = "testing..."
#   type = "victorops"
#   settings = {
#     victorops = {
#       api_key = "xyz"
#       routing_key = "123"
#     }
#   }
# }

# resource "swo_notification" "test_sms" {
#   title = "RobS SMS Test"
#   description = "testing..."
#   type = "sms"
#   settings = {
#     sms = {
#       phone_numbers = "+1 999 999 9999"
#     }
#   }
# }

# resource "swo_notification" "test_servicenow" {
#   title = "RobS ServiceNow Test"
#   description = "testing..."
#   type = "servicenow"
#   settings = {
#     servicenow = {
#       app_token = "xyz"
#       instance = "US"
#     }
#   }
# }

# resource "swo_notification" "test_swsd" {
#   title = "RobS ServiceDesk Test"
#   description = "testing..."
#   type = "swsd"
#   settings = {
#     swsd = {
#       app_token = "xyz"
#       is_eu = false
#     }
#   }
# }

# resource "swo_alert" "https_response_time" {
#   name        = "High HTTPS Response Time"
#   description = "A high response time has been identified."
#   severity    = "CRITICAL"
#   type        = "ENTITY_METRIC"
#   enabled     = true
#   target_entity_types = [
#     "Website"
#   ]
#   conditions = [
#     {
#       metric_name      = "synthetics.https.response.time"
#       threshold        = ">=3000ms"
#       duration         = "5m"
#       aggregation_type = "AVG"
#       entity_ids = [
#         "e-1521946194448543744",
#         "e-1521947552186691584"
#       ]
#       include_tags = [
#         {
#           name = "probe.city"
#           values = [
#             "Tokyo",
#             "Sao Paulo"
#           ]
#         }
#       ],
#       exclude_tags = []
#     },
#   ]
#   notifications = [123, 456]
# }
>>>>>>> 4e685e6 (Nh 32994 notifications client and terraform resource (#6))
