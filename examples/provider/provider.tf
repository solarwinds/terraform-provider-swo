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
  # The 'SWO_API_TOKEN' environment variable can be set as an alternative to using this field.
  # If 'api_token' is not provided, The provider attempt to use the 'SWO_API_TOKEN' environment variable.
  api_token = "[UPDATE WITH SWO FULL ACCESS TOKEN]"

  # Base URL for your SWO instance. Be sure to include your specific datacenter.
  # Datacenter options are one of [na-01, na-02, eu-01]. Overwrites base_url_env_name.
  # The 'SWO_BASE_URL' environment variable can be set as an alternative to using this field.
  # If 'base_url' is not provided, The provider will attempt to use the 'SWO_BASE_URL' environment variable.
  base_url = "https://api.na-01.cloud.solarwinds.com/v1/tfproxy"
}
