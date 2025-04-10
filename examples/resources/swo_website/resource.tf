resource "swo_website" "test_website" {
  name = "example-website"
  url  = "https://example.com"

  monitoring = {

    availability = {
      check_for_string = {
        operator = "CONTAINS"
        value    = "example-string"
      }

      ssl = {
        days_prior_to_expiration         = 30
        enabled                          = true
        ignore_intermediate_certificates = true
      }

      protocols                = ["HTTP", "HTTPS"]
      test_interval_in_seconds = 300
      test_from_location       = "REGION"

      location_options = [
        {
          type  = "REGION"
          value = "NA"
        },
        {
          type  = "REGION"
          value = "AS"
        },
        {
          type  = "REGION"
          value = "SA"
        },
        {
          type  = "REGION"
          value = "OC"
        }
      ]

      platform_options = {
        test_from_all = false
        platforms     = ["AWS"]
      }

      custom_headers = [
        {
          name  = "Custom-Header-1"
          value = "Custom-Value-1"
        },
        {
          name  = "Custom-Header-2"
          value = "Custom-Value-2"
        }
      ]
    }

    rum = {
      apdex_time_in_seconds = 4
      spa                   = true
    }
  }
}
