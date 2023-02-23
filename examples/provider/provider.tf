terraform {
  required_providers {
    swo = {
      version = "0.0.1"
      source  = "github.com/solarwindscloud/swo"
    }
  }
}

provider "swo" {
  api_token             = "[UPDATE WITH SWO TOKEN]"
  request_retry_timeout = 10
  debug_mode = true
}

resource "swo_alert" "https_response_time" {
  name        = "High HTTPS Response Time"
  description = "A high response time has been identified."
  severity    = "CRITICAL"
  type        = "ENTITY_METRIC"
  enabled     = true
  target_entity_types = [
    "Website"
  ]
  conditions = [
    {
      metric_name      = "synthetics.https.response.time"
      threshold        = ">=3000ms"
      duration         = "5m"
      aggregation_type = "AVG"
      entity_ids = [
        "e-1521946194448543744",
        "e-1521947552186691584"
      ]
      include_tags = [
        {
          name = "probe.city"
          values : [
            "Tokyo",
            "Sao Paulo"
          ]
        }
      ],
      exclude_tags = []
    },
  ]
  notifications = [123, 456]
}


resource "swo_alert" "https_response_time" {
  name        = "High HTTPS Response Time"
  description = "A high response time has been identified."
  severity    = "CRITICAL"
  type        = "ENTITY_METRIC"
  enabled     = true
  target_entity_types = [
    "Website"
  ]
  conditions = [
    {
      metric_name      = "synthetics.https.response.time"
      threshold        = ">=3000ms"
      duration         = "5m"
      aggregation_type = "AVG"
      entity_ids = [
        "e-1521946194448543744",
        "e-1521947552186691584"
      ]
      include_tags = [
        {
          name = "probe.city"
          values : [
            "Tokyo",
            "Sao Paulo"
          ]
        }
      ],
      exclude_tags = []
    },
  ]
  notifications = [123, 456]
}
