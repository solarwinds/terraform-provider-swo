# 1. Create the SWO CircleCI integration
resource "swo_circleci_integration" "main" {
  name          = "My Organization"
  api_token     = "CIRCLECI_API_TOKEN_VALUE"                      # Optional: CircleCI API token for log aggregation
  receiver_base = "https://webhook.swo.XX.solarwinds.com/webhook" # Optional: Prefix for receiver_url. SWO endpoint where your organization is monitored.
}

# 2. For each CircleCI project, add or remove webhook blocks as needed
resource "circleci_webhook" "service_a" {
  name           = "SolarWinds Observability"
  url            = swo_circleci_integration.main.receiver_url
  signing_secret = swo_circleci_integration.main.secret_token
  scope_id       = "PROJECT_ID" # Replace with your CircleCI project UUID
  scope_type     = "project"
  events         = ["workflow-completed", "job-completed"] # SWO requires both
  verify_tls     = true                                    # SWO requires TLS certificate verification
  is_active      = true
}
