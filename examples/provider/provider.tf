terraform {
  required_providers {
    swo = {
      version = ">= 0.0.7"
      source  = "solarwinds/swo"
      # Uncomment the following line to use the latest version of the provider from GitHub.
      # source = "github.com/solarwinds/swo"
    }
  }
}

provider "swo" {
  # API token. Tokens can be created in your SWO account settings under API tokens.
  # The token type should be Full Access.
  api_token = "[UPDATE WITH SWO FULL ACCESS TOKEN]"

  # Base URL for your SWO instance. Be sure to include your specific datacenter.
  # Datacenter options are one of [na-01, na-02, eu-01, apj-01].
  base_url = "https://api.na-01.cloud.solarwinds.com/v1/tfproxy"
}
