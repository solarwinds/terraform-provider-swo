package validators

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
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

func TestValidateList(t *testing.T) {
	listOptionsValidator := ListOptions(notificationActionTypes...)
	testInputs := []string{"1234:email", "456:amazonSns", "6785:newRelic"}

	validateTestList(t, testInputs, true, listOptionsValidator)
}

func TestValidateListWithInvalidId(t *testing.T) {
	listOptionsValidator := ListOptions(notificationActionTypes...)
	testInputs := []string{"invalid:email", "456:amazonSns", "6785:newRelic"}

	validateTestList(t, testInputs, false, listOptionsValidator)
}

func TestValidateListWithInvalidType(t *testing.T) {
	listOptionsValidator := ListOptions(notificationActionTypes...)
	testInputs := []string{"1234:email", "456:amazonSns", "6785:invalid"}

	validateTestList(t, testInputs, false, listOptionsValidator)
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

func validateTestList(t *testing.T, tests []string, expected bool, listOptionsValidator validator.List) {
	var elements []attr.Value
	for _, test := range tests {
		elements = append(elements, types.StringValue(test))
	}
	listValue, diag := types.ListValue(types.StringType, elements)

	if diag.HasError() {
		t.Errorf("Diagnostic errors. %s", diag.Errors())
	}

	req := validator.ListRequest{
		ConfigValue: listValue,
	}

	resp := &validator.ListResponse{}
	listOptionsValidator.ValidateList(context.Background(), req, resp)

	if (len(resp.Diagnostics) == 0) != expected {
		t.Errorf("ValidateList(%q) = %v, expected %v", tests, len(resp.Diagnostics) == 0, expected)
	}
}
