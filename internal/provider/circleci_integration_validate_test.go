package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestHasApiTokenWithoutName(t *testing.T) {
	tests := []struct {
		name      string
		apiToken  types.String
		tokenName types.String
		want      bool
	}{
		{
			name:      "token with name is fine",
			apiToken:  types.StringValue("CCI-token"),
			tokenName: types.StringValue("CI reader"),
			want:      false,
		},
		{
			name:      "token without name is invalid",
			apiToken:  types.StringValue("CCI-token"),
			tokenName: types.StringNull(),
			want:      true,
		},
		{
			name:      "token with empty name is invalid",
			apiToken:  types.StringValue("CCI-token"),
			tokenName: types.StringValue(""),
			want:      true,
		},
		{
			name:      "no token is fine regardless of name",
			apiToken:  types.StringNull(),
			tokenName: types.StringNull(),
			want:      false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := hasApiTokenWithoutName(circleCIIntegrationResourceModel{
				ApiToken:     tc.apiToken,
				ApiTokenName: tc.tokenName,
			})
			if got != tc.want {
				t.Errorf("hasApiTokenWithoutName() = %v, want %v", got, tc.want)
			}
		})
	}
}
