resource "swo_alert" "https_response_time" {
  name        = "High HTTPS Response Time"
  description = "A high response time has been identified."
  severity    = "INFO"
  enabled     = true
  notification_actions = [
    {
      type                    = "msteams"
      configuration_ids       = [swo_notification.msteams.id, swo_notification.opsgenie.id]
      resend_interval_seconds = 600
    },
  ]
  conditions = [
    {
      metric_name      = "synthetics.https.response.time"
      not_reporting    = false
      threshold        = ">=3000"
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
  notifications         = [swo_notification.msteams.id, swo_notification.opsgenie.id]
  trigger_reset_actions = true
  runbookLink           = "https://www.runbook.com/highresponsetime"
}
