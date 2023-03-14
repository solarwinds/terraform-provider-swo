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
  # base_url = "https://my.na-01.cloud.solarwinds.com/common/graphql"
}

# resource "swo_dashboard" "metrics_dashboard" {
#   name = "terraform-provider-swo [TEST]"
#   is_private = true
#   # category_id = 
#   widgets = [
#     {
#       type = "Kpi"
#       x = 0
#       y = 0
#       width = 3
#       height = 2
#       properties = <<EOF
#       {
#         "unit": "ms",
#         "title": "Widget Title",
#         "linkUrl": "https://www.solarwinds.com",
#         "subtitle": "Widget Subtitle",
#         "linkLabel": "Linky",
#         "dataSource": {
#           "type": "kpi",
#           "properties": {
#             "series": [
#               {
#                 "type": "metric",
#                 "limit": {
#                   "value": 50,
#                   "isAscending": false
#                 },
#                 "metric": "synthetics.https.response.time",
#                 "groupBy": [],
#                 "formatOptions": {
#                   "unit": "ms",
#                   "precision": 3,
#                   "minUnitSize": -2
#                 },
#                 "bucketGrouping": [],
#                 "aggregationFunction": "AVG"
#               }
#             ],
#             "isHigherBetter": false,
#             "includePercentageChange": true
#           }
#         }
#       }
#       EOF
#     },
#     {
#       type = "TimeSeries"
#       x = 3
#       y = 0
#       width = 9
#       height = 2
#       properties = <<EOF
#       {
#         "title": "Widget",
#         "subtitle": "",
#         "chart": {
#           "type": "LineChart",
#           "max": "auto",
#           "yAxisLabel": "",
#           "showLegend": true,
#           "yAxisFormatOverrides": {
#             "conversionFactor": 1,
#             "precision": 3
#           },
#           "formatOptions": {
#             "unit": "ms",
#             "minUnitSize": -2,
#             "precision": 3
#           }
#         },
#         "dataSource": {
#           "type": "timeSeries",
#           "properties": {
#             "series": [
#               {
#                 "type": "metric",
#                 "metric": "synthetics.https.response.time",
#                 "aggregationFunction": "AVG",
#                 "bucketGrouping": [],
#                 "groupBy": [
#                   "probe.region"
#                 ],
#                 "limit": {
#                   "value": 50,
#                   "isAscending": false
#                 },
#                 "formatOptions": {
#                   "unit": "ms",
#                   "minUnitSize": -2,
#                   "precision": 3
#                 }
#               },
#               {
#                 "type": "metric",
#                 "metric": "synthetics.error_rate",
#                 "aggregationFunction": "AVG",
#                 "bucketGrouping": [],
#                 "groupBy": [
#                   "probe.region"
#                 ],
#                 "limit": {
#                   "value": 50,
#                   "isAscending": false
#                 },
#                 "formatOptions": {
#                   "unit": "%",
#                   "precision": 3
#                 }
#               }
#             ]
#           }
#         }
#       }
#       EOF
#     },
#     {
#       type = "Proportional"
#       x = 0
#       y = 2
#       width = 12
#       height = 2
#       properties = <<EOF
#       {
#           "title": "Widget",
#           "subtitle": "",
#           "type": "HorizontalBar",
#           "showLegend": false,
#           "formatOptions": {
#               "unit": "ms"
#           },
#           "dataSource": {
#               "type": "proportional",
#               "properties": {
#                   "series": [
#                       {
#                           "type": "metric",
#                           "metric": "synthetics.http.response.time",
#                           "aggregationFunction": "AVG",
#                           "bucketGrouping": [],
#                           "groupBy": [
#                               "synthetics.target"
#                           ],
#                           "limit": {
#                               "value": 10,
#                               "isAscending": true
#                           },
#                           "formatOptions": {
#                               "unit": "ms",
#                               "minUnitSize": -2,
#                               "precision": 3
#                           }
#                       }
#                   ]
#               }
#           }
#       }
#       EOF
#     }
#   ]
# }

# resource "swo_notification" "test_email" {
#   title = "terraform-provider-swo [TEST]"
#   description = "testing..."
#   type = "email"
#   settings = {
#     email = {
#       addresses = [
#         {
#           email = "user1@host.com"
#         },
#         {
#           email = "user2@host.com"
#         },
#       ]
#     }
#   }
# }

# resource "swo_notification" "test_msteams" {
#   title = "terraform-provider-swo [TEST]"
#   description = "testing..."
#   type = "msTeams"
#   settings = {
#     msteams = {
#       url = "https://www.office.com/webhook"
#     }
#   }
# }

# resource "swo_notification" "test_opsgenie" {
#   title = "terraform-provider-swo [TEST]"
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
#   title = "terraform-provider-swo [TEST]"
#   description = "testing..."
#   type = "slack"
#   settings = {
#     slack = {
#       url = "https://hooks.slack.com/services/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"
#     }
#   }
# }

# resource "swo_notification" "test_pagerduty" {
#   title = "terraform-provider-swo [TEST]"
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
#   title = "terraform-provider-swo [TEST]"
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
#   title = "terraform-provider-swo [TEST]"
#   description = "testing..."
#   type = "sms"
#   settings = {
#     sms = {
#       phone_numbers = "+1 999 999 9999"
#     }
#   }
# }

# resource "swo_notification" "test_servicenow" {
#   title = "terraform-provider-swo [TEST]"
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
#   title = "terraform-provider-swo [TEST]"
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
