package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

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
					resource.TestCheckResourceAttr("swo_notification.test_amazonsns", "type", "amazonsns"),
					resource.TestCheckResourceAttr("swo_notification.test_amazonsns", "settings.amazonsns.access_key_id", "KEY_ID"),
					resource.TestCheckResourceAttr("swo_notification.test_amazonsns", "settings.amazonsns.secret_access_key", "SECRET_KEY"),
					resource.TestCheckResourceAttr("swo_notification.test_amazonsns", "settings.amazonsns.topic_arn", "arn:aws:sns:us-east-1:123456789012:topic"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "swo_notification.test_amazonsns",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"settings.amazonsns.secret_access_key"},
			},
			// Update and Read testing
			{
				Config: testAccAmazonSnsConfig("test-acc test two"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("swo_notification.test_amazonsns", "title", "test-acc test two"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
func testAccAmazonSnsConfig(title string) string {
	return providerConfig() + fmt.Sprintf(`
	resource "swo_notification" "test_amazonsns" {
  		title       = %[1]q
  		description = "testing..."
  		type        = "amazonsns"
  		settings = {
    		amazonsns = {
				topic_arn = "arn:aws:sns:us-east-1:123456789012:topic"
				access_key_id = "KEY_ID"
				secret_access_key = "SECRET_KEY"
			}
		}
	}`, title)
}

func TestAccEmailResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		IsUnitTest:               true,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccEmailConfig("test-acc test one"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("swo_notification.test_email", "id"),
					resource.TestCheckResourceAttr("swo_notification.test_email", "title", "test-acc test one"),
					resource.TestCheckResourceAttr("swo_notification.test_email", "description", "testing..."),
					resource.TestCheckResourceAttr("swo_notification.test_email", "type", "email"),
					resource.TestCheckResourceAttr("swo_notification.test_email", "settings.email.addresses.0.email", "test1@host.com"),
					resource.TestCheckResourceAttr("swo_notification.test_email", "settings.email.addresses.1.email", "test2@host.com"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "swo_notification.test_email",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccEmailConfig("test-acc test two"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("swo_notification.test_email", "title", "test-acc test two"),
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
					resource.TestCheckResourceAttr("swo_notification.test_msteams", "type", "msTeams"),
					resource.TestCheckResourceAttr("swo_notification.test_msteams", "settings.msteams.url", "https://XXX.webhook.office.com/webhookb2/XXXXX"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "swo_notification.test_msteams",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccMsTeamsConfig("test-acc test two"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("swo_notification.test_msteams", "title", "test-acc test two"),
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
      			url = "https://XXX.webhook.office.com/webhookb2/XXXXX"
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
					resource.TestCheckResourceAttr("swo_notification.test_opsgenie", "type", "opsgenie"),
					resource.TestCheckResourceAttr("swo_notification.test_opsgenie", "settings.opsgenie.hostname", "hostname"),
					resource.TestCheckResourceAttr("swo_notification.test_opsgenie", "settings.opsgenie.api_key", "API_KEY"),
					resource.TestCheckResourceAttr("swo_notification.test_opsgenie", "settings.opsgenie.recipients", "on-call recipient"),
					resource.TestCheckResourceAttr("swo_notification.test_opsgenie", "settings.opsgenie.teams", "team1, team2"),
					resource.TestCheckResourceAttr("swo_notification.test_opsgenie", "settings.opsgenie.tags", "tag1, tag2"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "swo_notification.test_opsgenie",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"settings.opsgenie.api_key"},
			},
			// Update and Read testing
			{
				Config: testAccOpsGenieConfig("test-acc test two"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("swo_notification.test_opsgenie", "title", "test-acc test two"),
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
      			api_key     = "API_KEY"
      			recipients = "on-call recipient"
      			teams      = "team1, team2"
      			tags       = "tag1, tag2"
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
					resource.TestCheckResourceAttr("swo_notification.test_pagerduty", "type", "pagerduty"),
					resource.TestCheckResourceAttr("swo_notification.test_pagerduty", "settings.pagerduty.routing_key", "99999999999999999999999999999999"),
					resource.TestCheckResourceAttr("swo_notification.test_pagerduty", "settings.pagerduty.summary", "some-summary"),
					resource.TestCheckResourceAttr("swo_notification.test_pagerduty", "settings.pagerduty.dedup_key", "DEDUP_KEY"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "swo_notification.test_pagerduty",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"settings.pagerduty.routing_key"},
			},
			// Update and Read testing
			{
				Config: testAccPagerdutyConfig("test-acc test two"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("swo_notification.test_pagerduty", "title", "test-acc test two"),
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
			pagerduty = {
				routing_key = "99999999999999999999999999999999"
      			summary     = "some-summary"
      			dedup_key   = "DEDUP_KEY"
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
					resource.TestCheckResourceAttr("swo_notification.test_pushover", "type", "pushover"),
					resource.TestCheckResourceAttr("swo_notification.test_pushover", "settings.pushover.app_token", "APP_TOKEN"),
					resource.TestCheckResourceAttr("swo_notification.test_pushover", "settings.pushover.user_key", "123xyz"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "swo_notification.test_pushover",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"settings.pushover.app_token"},
			},
			// Update and Read testing
			{
				Config: testAccPushoverConfig("test-acc test two"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("swo_notification.test_pushover", "title", "test-acc test two"),
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
					resource.TestCheckResourceAttr("swo_notification.test_servicenow", "type", "servicenow"),
					resource.TestCheckResourceAttr("swo_notification.test_servicenow", "settings.servicenow.app_token", "APP_TOKEN"),
					resource.TestCheckResourceAttr("swo_notification.test_servicenow", "settings.servicenow.instance", "US"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "swo_notification.test_servicenow",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"settings.servicenow.app_token"},
			},
			// Update and Read testing
			{
				Config: testAccServicenowConfig("test-acc test two"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("swo_notification.test_servicenow", "title", "test-acc test two"),
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
      			app_token = "APP_TOKEN"
      			instance  = "US"
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
					resource.TestCheckResourceAttr("swo_notification.test_slack", "type", "slack"),
					resource.TestCheckResourceAttr("swo_notification.test_slack", "settings.slack.url", "https://hooks.slack.com/services/XXX/XXX/XXX"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "swo_notification.test_slack",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccSlackConfig("test-acc test two"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("swo_notification.test_slack", "title", "test-acc test two"),
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
					resource.TestCheckResourceAttr("swo_notification.test_swsd", "type", "swsd"),
					resource.TestCheckResourceAttr("swo_notification.test_swsd", "settings.swsd.app_token", "APP_TOKEN"),
					resource.TestCheckResourceAttr("swo_notification.test_swsd", "settings.swsd.is_eu", "false"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "swo_notification.test_swsd",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"settings.swsd.app_token"},
			},
			// Update and Read testing
			{
				Config: testAccSwsdConfig("test-acc test two"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("swo_notification.test_swsd", "title", "test-acc test two"),
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
					resource.TestCheckResourceAttr("swo_notification.test_webhook", "type", "webhook"),
					resource.TestCheckResourceAttr("swo_notification.test_webhook", "settings.webhook.method", "GET"),
					resource.TestCheckResourceAttr("swo_notification.test_webhook", "settings.webhook.url", "https://webhook.example.com/"),
					resource.TestCheckResourceAttr("swo_notification.test_webhook", "settings.webhook.auth_header_name", "X-Request-Id"),
					resource.TestCheckResourceAttr("swo_notification.test_webhook", "settings.webhook.auth_header_value", "HEADER_VALUE"),
					resource.TestCheckResourceAttr("swo_notification.test_webhook", "settings.webhook.auth_password", "AUTH_PASSWORD"),
					resource.TestCheckResourceAttr("swo_notification.test_webhook", "settings.webhook.auth_type", "basic"),
					resource.TestCheckResourceAttr("swo_notification.test_webhook", "settings.webhook.auth_username", "AUTH_USERNAME"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "swo_notification.test_webhook",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"settings.webhook.auth_password", "settings.webhook.auth_header_value"},
			},
			// Update and Read testing
			{
				Config: testAccWebhookConfig("test-acc test two"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("swo_notification.test_webhook", "title", "test-acc test two"),
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
				auth_header_name = "X-Request-Id"
				auth_header_value = "HEADER_VALUE"
				auth_password = "AUTH_PASSWORD"
				auth_type = "basic"
				auth_username = "AUTH_USERNAME"
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
					resource.TestCheckResourceAttr("swo_notification.test_zapier", "type", "zapier"),
					resource.TestCheckResourceAttr("swo_notification.test_zapier", "settings.zapier.url", "https://hooks.zapier.com/hooks/catch/XXX"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "swo_notification.test_zapier",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccZapierConfig("test-acc test two"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("swo_notification.test_zapier", "title", "test-acc test two"),
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
      			url = "https://hooks.zapier.com/hooks/catch/XXX"
			}
		}
	}`, title)
}
