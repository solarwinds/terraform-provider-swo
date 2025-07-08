resource "swo_alert" "alert_with_metric_condition" {
  name        = "Alert with Metric Condition"
  description = "A high response time has been identified."
  severity    = "INFO"
  enabled     = true
  notification_actions = [
    {
      configuration_ids       = ["4661:email", "8112:webhook", "2456:newrelic"]
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
      query_search = "healthScore.categoryV2:good"
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
  trigger_reset_actions = true
  runbook_link          = "https://www.runbook.com/highresponsetime"
  trigger_delay_seconds = 300
}

resource "swo_alert" "alert_with_attribute_conditions" {
  name        = "Alert with Attribute Conditions"
  description = "Alert on conditions below."
  severity    = "INFO"
  enabled     = true
  notification_actions = [
    {
      configuration_ids       = ["4661:email", "8112:webhook", "2456:newrelic"]
      resend_interval_seconds = 600
    },
  ]
  conditions = [
    {
      attribute_name      = "inMaintenance"
      attribute_value     = "true"
      attribute_operator  = "="
      target_entity_types = ["Website"]
    },
    {
      attribute_name      = "healthScore.scoreV2"
      attribute_values    = "0"
      attribute_operator  = "="
      target_entity_types = ["Website"]
    }
  ]
  trigger_reset_actions = true
  runbook_link          = "https://www.runbook.com/highresponsetime"
  trigger_delay_seconds = 300
}
