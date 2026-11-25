package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestStandardizeJSONPlanModifier(t *testing.T) {
	tests := []struct {
		name      string
		value     types.String
		want      types.String
		wantError bool
	}{
		{name: "unknown", value: types.StringUnknown(), want: types.StringUnknown()},
		{name: "null", value: types.StringNull(), want: types.StringNull()},
		{
			name:  "valid JSON is normalized",
			value: types.StringValue(`{ "title": "CPU", "legend": true }`),
			want:  types.StringValue(`{"legend":true,"title":"CPU"}`),
		},
		{
			name:  "normalized JSON is unchanged",
			value: types.StringValue(`{"title":"CPU"}`),
			want:  types.StringValue(`{"title":"CPU"}`),
		},
		{name: "invalid JSON", value: types.StringValue(`{"title":`), wantError: true},
		{name: "empty string", value: types.StringValue(""), wantError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := planmodifier.StringRequest{PlanValue: tt.value, ConfigValue: tt.value}
			resp := planmodifier.StringResponse{PlanValue: tt.value}
			useStandarizedJson().PlanModifyString(context.Background(), req, &resp)

			if resp.Diagnostics.HasError() != tt.wantError {
				t.Fatalf("unexpected diagnostics: %v", resp.Diagnostics)
			}
			if !tt.wantError && !resp.PlanValue.Equal(tt.want) {
				t.Errorf("plan value = %s, want %s", resp.PlanValue, tt.want)
			}
		})
	}
}

func TestWidgetsFromPlanRejectsInvalidJSON(t *testing.T) {
	ctx := context.Background()
	widgets, d := types.SetValueFrom(ctx, types.ObjectType{AttrTypes: WidgetAttributeTypes()}, []dashboardWidgetModel{
		{Properties: types.StringValue(`{"title":`)},
	})
	if d.HasError() {
		t.Fatal(d)
	}

	var diagnostics diag.Diagnostics
	inputs, layouts := widgetsFromPlan(ctx, dashboardResourceModel{Widgets: widgets}, &diagnostics)
	if !diagnostics.HasError() {
		t.Fatal("expected invalid JSON to fail before creating API inputs")
	}
	if inputs != nil || layouts != nil {
		t.Fatal("invalid JSON must not produce widget or layout inputs")
	}
}
