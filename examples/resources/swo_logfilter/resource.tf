resource "swo_logfilter" "test" {
  name = "terraform-provider-swo example"
  description = "test log filter"
  token_signature = null
  expressions = [
    {
      kind = "STRING"
      expression = "test filter"
    }
  ]
}
