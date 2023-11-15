package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDashboardResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		IsUnitTest:               true,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config:             testAccDashboardResourceConfig("swo-terraform-provider [CREATE_TEST]"),
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("swo_dashboard.test", "id"),
					resource.TestCheckResourceAttr("swo_dashboard.test", "name", "swo-terraform-provider [CREATE_TEST]"),
					resource.TestCheckResourceAttrSet("swo_dashboard.test", "updated_at"),
					resource.TestCheckResourceAttrSet("swo_dashboard.test", "created_at"),
					resource.TestCheckResourceAttr("swo_dashboard.test", "widgets.#", "2"),
					resource.TestCheckResourceAttr("swo_dashboard.test", "widgets.0.type", "TimeSeries"),
					resource.TestCheckResourceAttr("swo_dashboard.test", "widgets.1.width", "4"),
					resource.TestCheckResourceAttrSet("swo_dashboard.test", "updated_at"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "swo_dashboard.test",
				ImportState:       true,
				ImportStateVerify: false, // False because the server sends widget properties back in a different format.
			},
			// Update and Read testing
			{
				Config:             testAccDashboardResourceConfig("swo-terraform-provider [UPDATE_TEST]"),
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("swo_dashboard.test", "name", "swo-terraform-provider [UPDATE_TEST]"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccDashboardResourceConfig(name string) string {
	return providerConfig + fmt.Sprintf(`
	resource "swo_dashboard" "test" {
		name = %[1]q
		is_private = true
		widgets = [
			{
				type = "Kpi"
				x = 0
				y = 0
				width = 4
				height = 2
				properties = <<EOF
				{
					"unit": "ms",
					"title": "Kpi Widget",
					"linkUrl": "https://www.solarwinds.com",
					"subtitle": "Widget with a Kpi display.",
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
				type = "TimeSeries"
				x = 4
				y = 0
				width = 4
				height = 2
				properties = <<EOF
				{
					"title": "TimeSeries Widget",
					"subtitle": "Widget with a TimeSeries chart.",
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
										"unit": "%%",
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
	}`, name)
}
