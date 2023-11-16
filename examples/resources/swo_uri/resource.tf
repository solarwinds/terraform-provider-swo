resource "swo_uri" "test" {
  name                = "terraform-provider-swo example"
  host                = "solarwinds.com"
  http_path_and_query = "/example?test=1"

  options = {
    is_ping_enabled = true
    is_http_enabled = true
    is_tcp_enabled  = false
  }

  http_options = {
    protocols = ["HTTP", "HTTPS"]

    check_for_string = {
      operator = "CONTAINS"
      value    = "example-string"
    }

    custom_headers = [
      {
        name  = "header1"
        value = "value1"
      },
      {
        name  = "header2"
        value = "value2"
      },
    ]
  }

  tcp_options = {
    port             = 80
    string_to_expect = "string to expect"
    string_to_send   = "string to send"
  }

  test_definitions = {
    test_from_location = "REGION"

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

    test_interval_in_seconds = 300

    platform_options = {
      test_from_all = false
      platforms     = ["AWS"]
    }
  }
}
