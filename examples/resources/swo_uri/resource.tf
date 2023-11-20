resource "swo_uri" "test" {
  name                = "terraform-provider-swo example"
  host                = "solarwinds.com"

  options = {
    is_ping_enabled = true
    is_tcp_enabled  = false
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
