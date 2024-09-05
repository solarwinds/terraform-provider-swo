resource "swo_logfilter" "test" {
  name            = "terraform-provider-swo example"
  description     = "test log filter"
  token_signature = swo_apitoken.an_ingestion_token.id
  expressions = [
    {
      kind       = "STRING"
      expression = "test filter"
    }
  ]
}
