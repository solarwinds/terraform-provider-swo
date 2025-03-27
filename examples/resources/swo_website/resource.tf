resource "swo_website" "test_website" {
  name: "example-website",
  url: "https://example.com",
  availability_check_settings = {
    check_for_string = {
       operator = "CONTAINS"
       value    = "example-string"
    },
    test_interval_in_seconds = 300,
    protocols = [
      "HTTP",
      "HTTPS"
    ],
    platform_options = {
      probe_platforms = [
        "AWS"
      ],
      test_from_all = false
    },
    test_from = {
      type = "REGION",
      values = [
        "NA",
        "AS",
        "SA",
        "OC"
      ]
    },
    ssl = {
      enabled = true,
      days_prior_to_expiration = 30,
      ignore_intermediate_certificates = true
    },
    custom_headers = [
      {
        name  = "Custom-Header-1"
        value = "Custom-Value-1"
      },
      {
        name  = "Custom-Header-2"
        value = "Custom-Value-2"
      }
    ],
    allow_insecure_renegotiation = true,
    post_data = "{\"example\= \"value\"}"
  },
  tags = [
    {
      key = "string",
      value = "string"
    }
  ],
  rum = {
    apdex_time_in_seconds = 4,
    spa = true
  }
}
