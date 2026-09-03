resource "swo_dashboard" "metrics_dashboard" {
  name               = "My metrics dashboard"
  version            = 2
  is_private         = true
  mode               = "Standard"
  description        = "My metrics dashboard"
  validation_version = 1
  enable_validation  = true
  widgets = [
    {
      type       = "Kpi"
      x          = 0
      y          = 0
      width      = 3
      height     = 2
      properties = <<EOF
      {
        "unit": "ms",
        "title": "Widget Title",
        "linkUrl": "https://www.solarwinds.com",
        "subtitle": "Widget Subtitle",
        "linkLabel": "Linky",
        "dataSource": {
          "type": "kpi",
          "properties": {
            "series": [
              {
                "type": "metric",
                "limit": {
                  "value": 50,
                  "isAscending": false
                },
                "metric": "synthetics.https.response.time",
                "entityId": "",
                "entityType": "All",
                "groupBy": [],
                "formatOptions": {
                  "unit": "ms",
                  "precision": 3,
                  "minUnitSize": -2
                },
                "bucketGrouping": [],
                "aggregationFunction": "AVG"
              }
            ],
            "isHigherBetter": false,
            "includePercentageChange": true
          }
        }
      }
      EOF
    },
    {
      type       = "TimeSeries"
      x          = 3
      y          = 0
      width      = 9
      height     = 2
      properties = <<EOF
      {
        "title": "Widget",
        "subtitle": "",
        "chart": {
          "type": "LineChart",
          "max": "auto",
          "yAxisLabel": "",
          "showLegend": true,
          "yAxisFormatOverrides": {
            "conversionFactor": 1,
            "precision": 3
          },
          "formatOptions": {
            "unit": "ms",
            "minUnitSize": -2,
            "precision": 3
          }
        },
        "dataSource": {
          "type": "timeSeries",
          "properties": {
            "series": [
              {
                "type": "metric",
                "metric": "synthetics.https.response.time",
                "entityId": "",
                "entityType": "All",
                "aggregationFunction": "AVG",
                "bucketGrouping": [],
                "groupBy": [
                  "probe.region"
                ],
                "limit": {
                  "value": 50,
                  "isAscending": false
                },
                "formatOptions": {
                  "unit": "ms",
                  "minUnitSize": -2,
                  "precision": 3
                }
              },
              {
                "type": "metric",
                "metric": "synthetics.error_rate",
                "entityId": "",
                "entityType": "All",
                "aggregationFunction": "AVG",
                "bucketGrouping": [],
                "groupBy": [
                  "probe.region"
                ],
                "limit": {
                  "value": 50,
                  "isAscending": false
                },
                "formatOptions": {
                  "unit": "%",
                  "precision": 3
                }
              }
            ]
          }
        }
      }
      EOF
    },
    {
      type       = "Proportional"
      x          = 0
      y          = 2
      width      = 12
      height     = 2
      properties = <<EOF
      {
          "title": "Widget",
          "subtitle": "",
          "type": "HorizontalBar",
          "sortDirection": "ascending",
          "showLegend": false,
          "formatOptions": {
              "unit": "ms"
          },
          "dataSource": {
              "type": "proportional",
              "properties": {
                  "series": [
                      {
                          "type": "metric",
                          "metric": "synthetics.http.response.time",
                          "entityId": "",
                          "entityType": "All",
                          "aggregationFunction": "AVG",
                          "bucketGrouping": [],
                          "groupBy": [
                              "synthetics.target"
                          ],
                          "limit": {
                              "value": 10,
                              "isAscending": true
                          },
                          "formatOptions": {
                              "unit": "ms",
                              "minUnitSize": -2,
                              "precision": 3
                          }
                      }
                  ]
              }
          }
      }
      EOF
    }
  ]
}
