package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	swoClient "github.com/solarwinds/swo-client-go/pkg/client"
)

func TestDashboardRefreshWidgets(t *testing.T) {
	var keepProperties any = map[string]any{"title": "Keep"}
	var otherProperties any = map[string]any{"title": "Other"}
	var editedProperties any = map[string]any{"title": "Edited"}
	keep := swoClient.ReadDashboardWidget{Id: "keep", Type: "TimeSeries", Properties: &keepProperties}
	deleted := swoClient.ReadDashboardWidget{Id: "delete", Type: "TimeSeries", Properties: &otherProperties}
	added := swoClient.ReadDashboardWidget{Id: "new", Type: "TimeSeries", Properties: &otherProperties}
	edited := swoClient.ReadDashboardWidget{Id: "keep", Type: "TimeSeries", Properties: &editedProperties}

	keepLayout := swoClient.ReadDashboardLayout{Id: "keep", Width: 6, Height: 4}
	deleteLayout := swoClient.ReadDashboardLayout{Id: "delete", X: 6, Width: 6, Height: 4}
	addLayout := swoClient.ReadDashboardLayout{Id: "new", X: 6, Width: 6, Height: 4}
	movedLayout := keepLayout
	movedLayout.X = 1

	keepState := dashboardWidgetModel{
		Id: types.StringValue("keep"), Type: types.StringValue("TimeSeries"),
		Properties: types.StringValue(`{ "title": "Keep" }`),
		X:          types.Int64Value(0), Y: types.Int64Value(0),
		Width: types.Int64Value(6), Height: types.Int64Value(4),
	}
	deleteState := keepState
	deleteState.Id = types.StringValue("delete")
	deleteState.X = types.Int64Value(6)
	deleteState.Properties = types.StringValue(`{"title":"Other"}`)
	addState := deleteState
	addState.Id = types.StringValue("new")
	movedState := keepState
	movedState.X = types.Int64Value(1)
	editedState := keepState
	editedState.Properties = types.StringValue(`{"title":"Edited"}`)

	original := []dashboardWidgetModel{keepState, deleteState}
	for _, tc := range []struct {
		name    string
		widgets []swoClient.ReadDashboardWidget
		layout  []swoClient.ReadDashboardLayout
		want    []dashboardWidgetModel
	}{
		{"unchanged", []swoClient.ReadDashboardWidget{keep, deleted}, []swoClient.ReadDashboardLayout{keepLayout, deleteLayout}, original},
		{"reordered", []swoClient.ReadDashboardWidget{deleted, keep}, []swoClient.ReadDashboardLayout{deleteLayout, keepLayout}, original},
		{"one deleted", []swoClient.ReadDashboardWidget{keep}, []swoClient.ReadDashboardLayout{keepLayout}, []dashboardWidgetModel{keepState}},
		{"all deleted", []swoClient.ReadDashboardWidget{}, []swoClient.ReadDashboardLayout{}, []dashboardWidgetModel{}},
		{"deleted, added and moved", []swoClient.ReadDashboardWidget{added, keep}, []swoClient.ReadDashboardLayout{movedLayout, addLayout}, []dashboardWidgetModel{movedState, addState}},
		{"properties changed", []swoClient.ReadDashboardWidget{edited, deleted}, []swoClient.ReadDashboardLayout{keepLayout, deleteLayout}, []dashboardWidgetModel{editedState, deleteState}},
		{"missing layout", []swoClient.ReadDashboardWidget{keep}, nil, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := t.Context()
			widgetType := types.ObjectType{AttrTypes: WidgetAttributeTypes()}
			before, d := types.SetValueFrom(ctx, widgetType, original)
			if d.HasError() {
				t.Fatal(d)
			}
			state := dashboardResourceModel{Widgets: before}
			remote := swoClient.ReadDashboardResult{Widgets: tc.widgets, Layout: tc.layout}

			var diagnostics diag.Diagnostics
			setDashboardValuesFromRead(ctx, &remote, &state, &diagnostics)
			if tc.want == nil {
				if !diagnostics.HasError() || !state.Widgets.Equal(before) {
					t.Fatalf("expected a layout error and unchanged widgets, got %v and %v", diagnostics, state.Widgets)
				}
				return
			}
			if diagnostics.HasError() {
				t.Fatal(diagnostics)
			}
			want := types.SetNull(widgetType)
			if len(tc.want) > 0 {
				want, d = types.SetValueFrom(ctx, widgetType, tc.want)
				if d.HasError() {
					t.Fatal(d)
				}
			}
			if !state.Widgets.Equal(want) {
				t.Errorf("refreshed widgets = %v, want %v", state.Widgets, want)
			}
		})
	}
}
