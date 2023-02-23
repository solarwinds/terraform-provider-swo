resource "swo_alert" "https_response_time" {
  name        = "TF Test - High HTTPS Response Time"
  description = ""
  severity    = "INFO"
  type        = "metric"
  entity_type = "Website"
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
  notifications = ["123", "456"]
}
