terraform {
  required_providers {
    swo = {
      version = ">= 0.0.13"
      source  = "solarwinds/swo"
    }
  }
  required_version = ">=1.0.7"
}

provider "swo" {
  # API token. Tokens can be created in your SWO account settings under API tokens.
  # The token type should be Full Access. Overwrites api_token_env_name.
  api_token = "[UPDATE WITH SWO FULL ACCESS TOKEN]"

  # Base URL for your SWO instance. Be sure to include your specific datacenter.
  # Datacenter options are one of [na-01, na-02, eu-01]. Overwrites base_url_env_name.
  base_url = "https://api.na-01.cloud.solarwinds.com/v1/tfproxy"

  # API token environment variable name. This variable should reference a Full Access SWO API token.  
  api_token_env_name = "SWO_API_TOKEN"

  # Base URL environment variable name. This variable should reference a base URL for your SWO instance.  
  base_url_env_name = "SWO_BASE_URL"
}
