package provider

import (
	"fmt"
	"regexp"
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
				Config:             testAccDashboardResourceConfig("test-acc swo-terraform-provider [CREATE_TEST]"),
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("swo_dashboard.test", "id"),
					resource.TestCheckResourceAttr("swo_dashboard.test", "name", "test-acc swo-terraform-provider [CREATE_TEST]"),
					resource.TestCheckResourceAttr("swo_dashboard.test", "is_private", "false"),
					resource.TestCheckNoResourceAttr("swo_dashboard.test", "version"),

					resource.TestCheckResourceAttr("swo_dashboard.test", "widgets.#", "2"),

					resource.TestCheckResourceAttr("swo_dashboard.test", "widgets.0.type", "TimeSeries"),
					resource.TestCheckResourceAttr("swo_dashboard.test", "widgets.0.x", "4"),
					resource.TestCheckResourceAttr("swo_dashboard.test", "widgets.0.y", "0"),
					resource.TestCheckResourceAttr("swo_dashboard.test", "widgets.0.width", "4"),
					resource.TestCheckResourceAttr("swo_dashboard.test", "widgets.0.height", "2"),

					resource.TestCheckResourceAttr("swo_dashboard.test", "widgets.1.type", "Kpi"),
					resource.TestCheckResourceAttr("swo_dashboard.test", "widgets.1.x", "0"),
					resource.TestCheckResourceAttr("swo_dashboard.test", "widgets.1.y", "0"),
					resource.TestCheckResourceAttr("swo_dashboard.test", "widgets.1.width", "4"),
					resource.TestCheckResourceAttr("swo_dashboard.test", "widgets.1.height", "2"),
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
				Config:             testAccDashboardResourceConfig("test-acc swo-terraform-provider [UPDATE_TEST]"),
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("swo_dashboard.test", "name", "test-acc swo-terraform-provider [UPDATE_TEST]"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccDashboardVersionNilResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		IsUnitTest:               true,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config:             testAccDashboardVersionNilResourceConfig("test-acc version=null [CREATE_TEST]"),
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("swo_dashboard.test", "id"),
					resource.TestCheckResourceAttr("swo_dashboard.test", "name", "test-acc version=null [CREATE_TEST]"),
					resource.TestCheckResourceAttr("swo_dashboard.test", "is_private", "false"),
					resource.TestCheckNoResourceAttr("swo_dashboard.test", "version"),

					resource.TestCheckResourceAttr("swo_dashboard.test", "widgets.#", "1"),

					resource.TestCheckResourceAttr("swo_dashboard.test", "widgets.0.type", "Kpi"),
					resource.TestCheckResourceAttr("swo_dashboard.test", "widgets.0.x", "0"),
					resource.TestCheckResourceAttr("swo_dashboard.test", "widgets.0.y", "0"),
					resource.TestCheckResourceAttr("swo_dashboard.test", "widgets.0.width", "4"),
					resource.TestCheckResourceAttr("swo_dashboard.test", "widgets.0.height", "2"),
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
				Config:             testAccDashboardVersionNilResourceConfig("test-acc version=null [UPDATE_TEST]"),
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("swo_dashboard.test", "name", "test-acc version=null [UPDATE_TEST]"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccDashboardVersion2Resource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		IsUnitTest:               true,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config:             testAccDashboardVersion2ResourceConfig("test-acc version=2 [CREATE_TEST]"),
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("swo_dashboard.test", "id"),
					resource.TestCheckResourceAttr("swo_dashboard.test", "name", "test-acc version=2 [CREATE_TEST]"),
					resource.TestCheckResourceAttr("swo_dashboard.test", "is_private", "false"),
					resource.TestCheckResourceAttr("swo_dashboard.test", "version", "2"),

					resource.TestCheckResourceAttr("swo_dashboard.test", "widgets.#", "1"),

					resource.TestCheckResourceAttr("swo_dashboard.test", "widgets.0.type", "Kpi"),
					resource.TestCheckResourceAttr("swo_dashboard.test", "widgets.0.x", "0"),
					resource.TestCheckResourceAttr("swo_dashboard.test", "widgets.0.y", "0"),
					resource.TestCheckResourceAttr("swo_dashboard.test", "widgets.0.width", "4"),
					resource.TestCheckResourceAttr("swo_dashboard.test", "widgets.0.height", "6"),
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
				Config:             testAccDashboardVersion2ResourceConfig("test-acc version=2 [UPDATE_TEST]"),
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("swo_dashboard.test", "name", "test-acc version=2 [UPDATE_TEST]"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccDashboardValidationResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		IsUnitTest:               true,
		Steps: []resource.TestStep{
			// Create and Read testing with widget validation enabled.
			{
				Config:             testAccDashboardValidationResourceConfig("test-acc validation [CREATE_TEST]"),
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("swo_dashboard.test", "id"),
					resource.TestCheckResourceAttr("swo_dashboard.test", "name", "test-acc validation [CREATE_TEST]"),
					resource.TestCheckResourceAttr("swo_dashboard.test", "enable_validation", "true"),
					resource.TestCheckResourceAttr("swo_dashboard.test", "validation_version", "1"),
					resource.TestCheckResourceAttr("swo_dashboard.test", "widgets.#", "1"),
					resource.TestCheckResourceAttr("swo_dashboard.test", "widgets.0.type", "Kpi"),
				),
			},
			// Update and Read testing (validation stays enabled).
			{
				Config:             testAccDashboardValidationResourceConfig("test-acc validation [UPDATE_TEST]"),
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("swo_dashboard.test", "name", "test-acc validation [UPDATE_TEST]"),
					resource.TestCheckResourceAttr("swo_dashboard.test", "enable_validation", "true"),
					resource.TestCheckResourceAttr("swo_dashboard.test", "validation_version", "1"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccDashboardValidationErrorResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		IsUnitTest:               true,
		Steps: []resource.TestStep{
			// Creating a dashboard with an invalid widget while validation is enabled must fail
			// with a server-side validation error surfaced by the client.
			// NOTE: the exact invalid payload / matched message may need adjusting to the server's
			// validation rules on first run against the dev environment.
			{
				Config:      testAccDashboardValidationInvalidResourceConfig("test-acc validation-error [CREATE_TEST]"),
				ExpectError: regexp.MustCompile(`(?s)(validationErrors|create dashboard failed)`),
			},
		},
	})
}

// func testAccDashboardValidationResourceConfig(name string) string {
// 	return providerConfig() + fmt.Sprintf(`
// 	resource "swo_dashboard" "test" {
// 		name = %[1]q
// 		is_private = false
// 		enable_validation = true
// 		validation_version = 1
// 		widgets = [
// 			{
// 				type = "Kpi"
// 				x = 0
// 				y = 0
// 				width = 4
// 				height = 2
// 				properties = <<EOF
// 				{
// 					"unit": "ms",
// 					"title": "Kpi Widget",
// 					"linkUrl": "https://www.solarwinds.com",
// 					"subtitle": "Widget with a Kpi display.",
// 					"linkLabel": "Linky",
// 					"dataSource": {
// 						"type": "kpi",
// 						"properties": {
// 							"series": [
// 								{
// 									"type": "metric",
// 									"limit": {
// 										"value": 50,
// 										"isAscending": false
// 									},
// 									"metric": "synthetics.https.response.time",
// 									"groupBy": [],
// 									"formatOptions": {
// 										"unit": "ms",
// 										"precision": 3,
// 										"minUnitSize": -2
// 									},
// 									"bucketGrouping": [],
// 									"aggregationFunction": "AVG"
// 								}
// 							],
// 							"isHigherBetter": false,
// 							"includePercentageChange": true
// 						}
// 					}
// 				}
// 				EOF
// 			}
// 		]
// 	}`, name)
// }

// testAccDashboardValidationInvalidResourceConfig defines a Kpi widget with empty properties.
// With enable_validation = true the server is expected to reject it with a validation error.
func testAccDashboardValidationInvalidResourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
	resource "swo_dashboard" "test" {
		name = %[1]q
		is_private = false
		enable_validation = true
		validation_version = 1
		widgets = [
			{
				type = "Kpi"
				x = 0
				y = 0
				width = 4
				height = 2
				properties = "{}"
			}
		]
	}`, name)
}

func testAccDashboardResourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
	resource "swo_dashboard" "test" {
		name = %[1]q
		is_private = false
		widgets = [
			{
				type = "Kpi"
				x = 0
				y = 0
				width = 4
				height = 2
				properties = <<EOF
				{
					"mode": "Standard",
					"description": "Kpi Widget",
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

func testAccDashboardVersionNilResourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
	resource "swo_dashboard" "test" {
		name = %[1]q
		is_private = false
		version = null
		widgets = [
			{
				type = "Kpi"
				x = 0
				y = 0
				width = 4
				height = 2
				properties = <<EOF
				{
					"mode": "Standard",
					"description": "Kpi Widget",
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
			}
		]
	}`, name)
}

func testAccDashboardVersion2ResourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
	resource "swo_dashboard" "test" {
		name = %[1]q
		is_private = false
		version = 2
		widgets = [
			{
				type = "Kpi"
				x = 0
				y = 0
				width = 4
				height = 6
				properties = <<EOF
				{
					"mode": "Standard",
					"description": "Kpi Widget",
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
			}
		]
	}`, name)
}
