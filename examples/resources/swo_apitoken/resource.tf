resource "swo_apitoken" "test" {
  name         = "terraform-provider-swo example"
  access_level = "FULL"
  type         = "public-api"
  enabled      = true
  attributes = [
    {
      key   = "attribute-key"
      value = "attribute value"
    }
  ]
}
