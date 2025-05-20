package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccEmailNotificationResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		IsUnitTest:               true,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccEmailConfig("test-acc test one"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("swo_notification.test", "id"),
					resource.TestCheckResourceAttr("swo_notification.test", "title", "test-acc test one"),
					resource.TestCheckResourceAttr("swo_notification.test", "description", "testing..."),
					resource.TestCheckResourceAttr("swo_notification.test", "type", "email"),
					resource.TestCheckResourceAttr("swo_notification.test", "settings.email.addresses.0.email", "test1@host.com"),
					resource.TestCheckResourceAttr("swo_notification.test", "settings.email.addresses.1.email", "test2@host.com"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "swo_notification.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccEmailConfig("test-acc test two"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("swo_notification.test", "title", "test-acc test two"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
func testAccEmailConfig(title string) string {
	return providerConfig() + fmt.Sprintf(`
	resource "swo_notification" "test_email" {
		title        = %[1]q
		description = "testing..."
		type = "email"
		settings = {
			email = {
				addresses = [
					{
						email = "test1@host.com"
					},
					{
						email = "test2@host.com"
					},
				]
			}
		}
	}`, title)
}

func TestAccAmazonSnsNotificationResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		IsUnitTest:               true,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccAmazonSnsConfig("test-acc test one"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("swo_notification.test", "type", "amazonsns"),
					resource.TestCheckResourceAttr("swo_notification.test", "settings.amazonsn.access_key_id", "KEY_ID"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "swo_notification.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccAmazonSnsConfig("test-acc test two"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("swo_notification.test", "title", "test-acc test two"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
func testAccAmazonSnsConfig(title string) string {
	return providerConfig() + fmt.Sprintf(`
	resource "swo_notification" "test_amazon_sns" {
  		title       = %[1]q
  		description = "testing..."
  		type        = "amazonsns"
  		settings = {
    		amazonsn = {
				access_key_id = "KEY_ID"
				secret_access_key = "SECRET_KEY"
				topic_arn = "arn:aws:sns:us-east-1:123456789012:topic"
		}
	}`, title)
}

func TestAccMsTeamsNotificationResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		IsUnitTest:               true,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccMsTeamsConfig("test-acc test one"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("swo_notification.test", "type", "msTeams"),
					resource.TestCheckResourceAttr("swo_notification.test", "settings.msteams.url", "https://www.office.com/webhook"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "swo_notification.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccMsTeamsConfig("test-acc test two"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("swo_notification.test", "title", "test-acc test two"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
func testAccMsTeamsConfig(title string) string {
	return providerConfig() + fmt.Sprintf(`
	resource "swo_notification" "test_msteams" {
  		title       = %[1]q
  		description = "testing..."
  		type        = "msTeams"
  		settings = {
    		msteams = {
      			url = "https://www.office.com/webhook"
			}
		}
	}`, title)
}

func TestAccOpsGenieNotificationResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		IsUnitTest:               true,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccOpsGenieConfig("test-acc test one"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("swo_notification.test", "type", "opsgenie"),
					resource.TestCheckResourceAttr("swo_notification.test", "settings.opsgenie.hostname", "hostname"),
					resource.TestCheckResourceAttr("swo_notification.test", "settings.opsgenie.apikey", "API_KEY"),
					resource.TestCheckResourceAttr("swo_notification.test", "settings.opsgenie.recipients", "alice"),
					resource.TestCheckResourceAttr("swo_notification.test", "settings.opsgenie.teams", "team1, team2"),
					resource.TestCheckResourceAttr("swo_notification.test", "settings.opsgenie.tags", "tag1, tag2"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "swo_notification.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccOpsGenieConfig("test-acc test two"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("swo_notification.test", "title", "test-acc test two"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
func testAccOpsGenieConfig(title string) string {
	return providerConfig() + fmt.Sprintf(`
	resource "swo_notification" "test_opsgenie" {
  		title       = %[1]q
  		description = "testing..."
  		type        = "opsgenie"
  		settings = {
    		opsgenie = {
      			hostname   = "hostname"
      			apikey     = "API_KEY"
      			recipients = "alice"
      			teams      = "team1, team2"
      			tags       = "tag1, tag2"
			}
		}
	}`, title)
}

func TestAccSlackNotificationResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		IsUnitTest:               true,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccSlackConfig("test-acc test one"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("swo_notification.test", "type", "slack"),
					resource.TestCheckResourceAttr("swo_notification.test", "settings.slack.url", "https://hooks.slack.com/services/XXX/XXX/XXX"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "swo_notification.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccSlackConfig("test-acc test two"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("swo_notification.test", "title", "test-acc test two"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
func testAccSlackConfig(title string) string {
	return providerConfig() + fmt.Sprintf(`
	resource "swo_notification" "test_slack" {
  		title       = %[1]q
  		description = "testing..."
  		type        = "slack"
  		settings = {
    		slack = {
      			url = "https://hooks.slack.com/services/XXX/XXX/XXX"
    		}
		}
	}`, title)
}

func TestAccPagerDutyNotificationResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		IsUnitTest:               true,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccPagerdutyConfig("test-acc test one"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("swo_notification.test", "type", "pagerduty"),
					resource.TestCheckResourceAttr("swo_notification.test", "settings.pagerduty.routing_key", "99999999999999999999999999999999"),
					resource.TestCheckResourceAttr("swo_notification.test", "settings.pagerduty.summary", "some-summary"),
					resource.TestCheckResourceAttr("swo_notification.test", "settings.pagerduty.dedup_key", "DEDUP"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "swo_notification.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccPagerdutyConfig("test-acc test two"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("swo_notification.test", "title", "test-acc test two"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
func testAccPagerdutyConfig(title string) string {
	return providerConfig() + fmt.Sprintf(`
	resource "swo_notification" "test_pagerduty" {
  		title       = %[1]q
  		description = "testing..."
  		type        = "pagerduty"
  		settings = {
    		routing_key = "99999999999999999999999999999999"
      		summary     = "some-summary"
      		dedup_key   = "DEDUP"
		}
	}`, title)
}

func TestAccVictorOpsNotificationResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		IsUnitTest:               true,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccVictorOpsConfig("test-acc test one"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("swo_notification.test", "type", "victorops"),
					resource.TestCheckResourceAttr("swo_notification.test", "settings.victorops.api_key", "API_KEY"),
					resource.TestCheckResourceAttr("swo_notification.test", "settings.victorops.routing_key", "ROUTING_KEY"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "swo_notification.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccVictorOpsConfig("test-acc test two"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("swo_notification.test", "title", "test-acc test two"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
func testAccVictorOpsConfig(title string) string {
	return providerConfig() + fmt.Sprintf(`
	resource "swo_notification" "test_victorops" {
  		title       = %[1]q
  		description = "testing..."
  		type        = "victorops"
  		settings = {
    		victorops = {
      			api_key     = "API_KEY"
      			routing_key = "ROUTING_KEY"
    		}
		}
	}`, title)
}

func TestAccSmsNotificationResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		IsUnitTest:               true,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccSmsConfig("test-acc test one"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("swo_notification.test", "type", "sms"),
					resource.TestCheckResourceAttr("swo_notification.test", "settings.sms.phone_numbers", "+1 999 999 9999"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "swo_notification.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccSmsConfig("test-acc test two"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("swo_notification.test", "title", "test-acc test two"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
func testAccSmsConfig(title string) string {
	return providerConfig() + fmt.Sprintf(`
	resource "swo_notification" "test_sms" {
  		title       = %[1]q
  		description = "testing..."
  		type        = "sms"
  		settings = {
    		sms = {
      			phone_numbers = "+1 999 999 9999"
    		}
		}
	}`, title)
}

func TestAccServiceNowNotificationResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		IsUnitTest:               true,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccServicenowConfig("test-acc test one"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("swo_notification.test", "type", "servicenow"),
					resource.TestCheckResourceAttr("swo_notification.test", "settings.servicenow.app_token", "API_TOKEN"),
					resource.TestCheckResourceAttr("swo_notification.test", "settings.servicenow.instance", "US"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "swo_notification.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccServicenowConfig("test-acc test two"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("swo_notification.test", "title", "test-acc test two"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
func testAccServicenowConfig(title string) string {
	return providerConfig() + fmt.Sprintf(`
	resource "swo_notification" "test_servicenow" {
  		title       = %[1]q
  		description = "testing..."
  		type        = "servicenow"
  		settings = {
    		servicenow = {
      			app_token = "API_TOKEN"
      			instance  = "US"
    		}
		}
	}`, title)
}

func TestAccSwsdNotificationResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		IsUnitTest:               true,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccSwsdConfig("test-acc test one"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("swo_notification.test", "type", "swsd"),
					resource.TestCheckResourceAttr("swo_notification.test", "settings.swsd.app_token", "APP_TOKEN"),
					resource.TestCheckResourceAttr("swo_notification.test", "settings.swsd.is_eu", "false"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "swo_notification.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccSwsdConfig("test-acc test two"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("swo_notification.test", "title", "test-acc test two"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
func testAccSwsdConfig(title string) string {
	return providerConfig() + fmt.Sprintf(`
	resource "swo_notification" "test_swsd" {
  		title       = %[1]q
  		description = "testing..."
  		type        = "swsd"
  		settings = {
    		swsd = {
      			app_token = "APP_TOKEN"
      			is_eu     = false
    		}
		}
	}`, title)
}

func TestAccPushoverNotificationResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		IsUnitTest:               true,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccPushoverConfig("test-acc test one"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("swo_notification.test", "type", "pushover"),
					resource.TestCheckResourceAttr("swo_notification.test", "settings.pushover.app_token", "APP_TOKEN"),
					resource.TestCheckResourceAttr("swo_notification.test", "settings.pushover.user_key", "123xyz"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "swo_notification.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccPushoverConfig("test-acc test two"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("swo_notification.test", "title", "test-acc test two"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
func testAccPushoverConfig(title string) string {
	return providerConfig() + fmt.Sprintf(`
	resource "swo_notification" "test_pushover" {
  		title       = %[1]q
  		description = "testing..."
  		type        = "pushover"
  		settings = {
    		pushover = {
      			app_token = "APP_TOKEN"
				user_key  = "123xyz"
			}
		}
	}`, title)
}

func TestAccWebhookNotificationResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		IsUnitTest:               true,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccWebhookConfig("test-acc test one"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("swo_notification.test", "type", "webhook"),
					resource.TestCheckResourceAttr("swo_notification.test", "settings.webhook.method", "GET"),
					resource.TestCheckResourceAttr("swo_notification.test", "settings.webhook.url", "https://webhook.example.com/"),
					resource.TestCheckResourceAttr("swo_notification.test", "settings.webhook.auth_header_name", "X-Slack-Request-Id"),
					resource.TestCheckResourceAttr("swo_notification.test", "settings.webhook.auth_header_value", "VALUE"),
					resource.TestCheckResourceAttr("swo_notification.test", "settings.webhook.auth_password", "PASSWORD"),
					resource.TestCheckResourceAttr("swo_notification.test", "settings.webhook.auth_type", "basic"),
					resource.TestCheckResourceAttr("swo_notification.test", "settings.webhook.auth_username", "USERNAME"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "swo_notification.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccWebhookConfig("test-acc test two"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("swo_notification.test", "title", "test-acc test two"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
func testAccWebhookConfig(title string) string {
	return providerConfig() + fmt.Sprintf(`
	resource "swo_notification" "test_webhook" {
  		title       = %[1]q
  		description = "testing..."
  		type        = "webhook"
  		settings = {
    		webhook = {
      			method = "GET"
				url    = "https://webhook.example.com/"
				auth_header_name = "X-Slack-Request-Id"
				auth_header_value = "VALUE"
				auth_password = "PASSWORD"
				auth_type = "basic"
				auth_username = "USERNAME"
			}
		}
	}`, title)
}

func TestAccZapierNotificationResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		IsUnitTest:               true,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccZapierConfig("test-acc test one"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("swo_notification.test", "type", "zapier"),
					resource.TestCheckResourceAttr("swo_notification.test", "settings.zapier.url", "https://www.office.com/webhook"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "swo_notification.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccZapierConfig("test-acc test two"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("swo_notification.test", "title", "test-acc test two"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
func testAccZapierConfig(title string) string {
	return providerConfig() + fmt.Sprintf(`
	resource "swo_notification" "test_zapier" {
  		title       = %[1]q
  		description = "testing..."
  		type        = "zapier"
  		settings = {
    		zapier = {
      			url = "https://www.office.com/webhook"
			}
		}
	}`, title)
}
