resource "swo_alert" "https_response_time" {
  name        = "TF Test - High HTTPS Response Time"
  description = "A high response time has been identified."
  severity    = "INFO"
  type        = "ENTITY_METRIC"
  enabled     = false
  conditions = [
    {
      metric_name      = "synthetics.https.response.time"
      threshold        = ">=3000ms" 
      duration         = "5m"
      aggregation_type = "AVG"
      target_entity_types = [
        "Website"
      ]
      entity_ids = [
        "e-1521946194448543744",
        "e-1521947552186691584"
      ]
      include_tags = [
        {
          name = "probe.city"
          values : [
            "Tokyo",
            "New York"
          ]
        }
      ],
      exclude_tags = []
    },
  ]
  notification_type = "email"
  notifications = [123, 456]
}
