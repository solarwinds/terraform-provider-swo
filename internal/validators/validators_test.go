package validators

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type validationTestInput struct {
	Input    string
	Expected bool
}

// Redeclaring to avoid creating a circular dependency.
var notificationActionTypes = []string{
	"email",
	"amazonsns",
	"msTeams",
	"newRelic",
	"opsgenie",
	"pagerduty",
	"pushover",
	"serviceNow",
	"slack",
	"webhook",
	"zapier",
	"swsd",
}

func TestValidateString(t *testing.T) {
	singleOptionValidator := SingleOption(notificationActionTypes...)

	testInputs := []validationTestInput{
		{"email", true},
		{"amazonsns", true},
		{"msTeams", true},
		{"newRelic", true},
		{"swsd", true},
		{"msteams", false},
		{"OPSgenie", false},
		{"", false},
		{"invalid", false},
		{"amazonsns", true},
	}

	validateTestString(t, testInputs, singleOptionValidator)
}

func TestCaseInsensitiveValidateString(t *testing.T) {
	singleOptionValidator := CaseInsensitiveSingleOption(notificationActionTypes...)

	testInputs := []validationTestInput{
		{"email", true},
		{"amazonSns", true},
		{"msteams", true},
		{"newRelic", true},
		{"swsd", true},
		{"msteam", false},
		{"MSteams", true},
		{"OPSgenie", true},
		{"", false},
		{"invalid", false},
		{"amazonsns", true},
	}

	validateTestString(t, testInputs, singleOptionValidator)
}

func validateTestString(t *testing.T, tests []validationTestInput, singleOptionValidator validator.String) {
	for _, test := range tests {

		req := validator.StringRequest{
			ConfigValue: types.StringValue(test.Input),
		}

		resp := &validator.StringResponse{}
		singleOptionValidator.ValidateString(context.Background(), req, resp)
		if (len(resp.Diagnostics) == 0) != test.Expected {
			t.Errorf("ValidateString(%q) = %v, expected %v", test.Input, len(resp.Diagnostics) == 0, test.Expected)
		}
	}
}
