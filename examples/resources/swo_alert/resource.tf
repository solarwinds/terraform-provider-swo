resource "swo_alert" "https_response_time" {
  name        = "High HTTPS Response Time"
  description = "A high response time has been identified."
  severity    = "INFO"
  enabled     = true
  conditions = [
    {
      metric_name      = "synthetics.https.response.time"
      threshold        = ">=3000"
      duration         = "5m"
      aggregation_type = "AVG"
      type             = "ENTITY_METRIC"
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
  notifications = [swo_notification.msteams.id, swo_notification.opsgenie.id]
}
