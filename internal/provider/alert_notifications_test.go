package provider

import (
	"slices"
	"testing"

	swoClient "github.com/solarwinds/swo-client-go/pkg/client"
)

func TestCompareAlertActionDescriptions(t *testing.T) {
	intPtr := func(i int) *int { return &i }

	tests := []struct {
		name     string
		input    []alertActionDescription
		expected []alertActionDescription
	}{
		{
			name: "orders by type first",
			input: []alertActionDescription{
				&swoClient.AlertActionInput{Type: "webhook"},
				&swoClient.AlertActionInput{Type: "email"},
			},
			expected: []alertActionDescription{
				&swoClient.AlertActionInput{Type: "email"},
				&swoClient.AlertActionInput{Type: "webhook"},
			},
		},
		{
			name: "orders by resend interval within the same type",
			input: []alertActionDescription{
				&swoClient.AlertActionInput{Type: "email", ResendIntervalSeconds: intPtr(300)},
				&swoClient.AlertActionInput{Type: "email", ResendIntervalSeconds: intPtr(60)},
			},
			expected: []alertActionDescription{
				&swoClient.AlertActionInput{Type: "email", ResendIntervalSeconds: intPtr(60)},
				&swoClient.AlertActionInput{Type: "email", ResendIntervalSeconds: intPtr(300)},
			},
		},
		{
			name: "nil resend interval sorts first",
			input: []alertActionDescription{
				&swoClient.AlertActionInput{Type: "email", ResendIntervalSeconds: intPtr(60)},
				&swoClient.AlertActionInput{Type: "email"},
			},
			expected: []alertActionDescription{
				&swoClient.AlertActionInput{Type: "email"},
				&swoClient.AlertActionInput{Type: "email", ResendIntervalSeconds: intPtr(60)},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := slices.Clone(tt.input)
			slices.SortFunc(actual, compareAlertActionDescriptions)
			for i := range actual {
				if compareAlertActionDescriptions(actual[i], tt.expected[i]) != 0 {
					t.Fatalf("index %d: got type %q interval %v, want type %q interval %v", i,
						actual[i].GetType(), actual[i].GetResendIntervalSeconds(),
						tt.expected[i].GetType(), tt.expected[i].GetResendIntervalSeconds())
				}
			}
		})
	}
}
