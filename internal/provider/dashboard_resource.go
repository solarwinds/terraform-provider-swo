package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	swoClient "github.com/solarwinds/swo-client-go/pkg/client"
	"golang.org/x/exp/slices"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &dashboardResource{}
	_ resource.ResourceWithConfigure   = &dashboardResource{}
	_ resource.ResourceWithImportState = &dashboardResource{}

	errLayoutMissing    = errors.New("layout missing for widget id")
	errWidgetProperties = errors.New("widget properties error")
)

func newLayoutError(id string) error {
	return fmt.Errorf("%w: %s", errLayoutMissing, id)
}

func newWidgetPropertiesError(msg string, id string) error {
	return fmt.Errorf("%w: %s id:%s", errWidgetProperties, msg, id)
}

func NewDashboardResource() resource.Resource {
	return &dashboardResource{}
}

// Defines the resource implementation.
type dashboardResource struct {
	client *swoClient.Client
}

func (r *dashboardResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "dashboard"
}

func (r *dashboardResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, _ := req.ProviderData.(providerClients)
	r.client = client.SwoClient
}

// Creates new WidgetInputs and LayoutInputs from plan widget data.
func widgetsFromPlan(ctx context.Context, plan dashboardResourceModel, diags *diag.Diagnostics) ([]swoClient.WidgetInput, []swoClient.LayoutInput) {

	var planWidgets []dashboardWidgetModel
	d := plan.Widgets.ElementsAs(ctx, &planWidgets, false)
	diags.Append(d...)
	if diags.HasError() {
		return nil, nil
	}

	widgets := make([]swoClient.WidgetInput, len(planWidgets))
	layouts := make([]swoClient.LayoutInput, len(planWidgets))

	for wIdx := range planWidgets {
		planW := &planWidgets[wIdx]
		id := uuid.NewString()

		// Marshal the json encoded properties string to an object.
		var props any
		err := json.Unmarshal([]byte(planW.Properties.ValueString()), &props)
		if err != nil {
			diags.AddError("swo provider error",
				fmt.Sprintf("convert plan to API error: %s", err))
			return nil, nil
		}

		widgets[wIdx] = swoClient.WidgetInput{
			Id:         id,
			Type:       planW.Type.ValueString(),
			Properties: &props,
		}

		layouts[wIdx] = swoClient.LayoutInput{
			Id:     id,
			X:      int(planW.X.ValueInt64()),
			Y:      int(planW.Y.ValueInt64()),
			Width:  int(planW.Width.ValueInt64()),
			Height: int(planW.Height.ValueInt64()),
		}
	}

	return widgets, layouts
}

// Sets the computed values of the dashboard models with the values returned from the Create request.
func setDashboardValuesFromCreate(ctx context.Context, dashboard *swoClient.CreateDashboardResult, plan *dashboardResourceModel, diags *diag.Diagnostics) {
	plan.Id = types.StringValue(dashboard.Id)

	// the client may modify 'version' value
	if dashboard.Version == nil {
		plan.Version = types.Int32PointerValue(nil)
	} else {
		dVersion := int32(*dashboard.Version)
		plan.Version = types.Int32PointerValue(&dVersion)
	}

	var planWidgets []dashboardWidgetModel
	d := plan.Widgets.ElementsAs(ctx, &planWidgets, false)
	diags.Append(d...)
	if diags.HasError() {
		return
	}

	for _, w := range dashboard.Widgets {
		lIdx := slices.IndexFunc(dashboard.Layout, func(l swoClient.CreateDashboardLayout) bool { return l.Id == w.Id })
		if lIdx <= -1 {
			diags.AddError("swo provider error",
				fmt.Sprintf("error setting computed values for dashboard: %s, id: %s", newLayoutError(w.Id), plan.Id))
			return
		}

		// The layout that will give us the widget coordinates for comparison to the plan.
		layout := &dashboard.Layout[lIdx]

		// We need to compare the properties of the plan widget with what is returned from the server
		// to reconcile the server data with the plan data. A widget's id in the plan is temporary, and
		// there isn't any single value we can use to make a match.
		for wIdx := range planWidgets {
			planW := &planWidgets[wIdx]
			if planW.Type.Equal(types.StringValue(w.Type)) &&
				planW.X.Equal(types.Int64Value(int64(layout.X))) &&
				planW.Y.Equal(types.Int64Value(int64(layout.Y))) &&
				planW.Width.Equal(types.Int64Value(int64(layout.Width))) &&
				planW.Height.Equal(types.Int64Value(int64(layout.Height))) {
				// Widget properties are equal, so we assume it must be the one we're looking for.
				planW.Id = types.StringValue(w.Id)
				break
			}
		}
	}

	updatedWidgets, d := types.SetValueFrom(ctx, types.ObjectType{AttrTypes: WidgetAttributeTypes()}, planWidgets)
	diags.Append(d...)
	if diags.HasError() {
		return
	}
	plan.Widgets = updatedWidgets
}

// Sets the values of the terraform state with the values returned from the Read request.
func setDashboardValuesFromRead(ctx context.Context, dashboard *swoClient.ReadDashboardResult, state *dashboardResourceModel, diags *diag.Diagnostics) {
	state.Id = types.StringValue(dashboard.Id)
	state.Name = types.StringValue(dashboard.Name)
	if dashboard.Category != nil {
		state.CategoryId = types.StringValue(dashboard.Category.Id)
	}
	if dashboard.IsPrivate != nil {
		state.IsPrivate = types.BoolValue(*dashboard.IsPrivate)
	}

	if dashboard.Version == nil {
		state.Version = types.Int32PointerValue(nil)
	} else {
		dVersion := int32(*dashboard.Version)
		state.Version = types.Int32PointerValue(&dVersion)
	}

	var stateWidgets []dashboardWidgetModel
	d := state.Widgets.ElementsAs(ctx, &stateWidgets, false)
	diags.Append(d...)
	if diags.HasError() {
		return
	}

	for _, w := range dashboard.Widgets {
		lIdx := slices.IndexFunc(dashboard.Layout, func(l swoClient.ReadDashboardLayout) bool { return l.Id == w.Id })
		if lIdx <= -1 {
			diags.AddError("swo provider error",
				fmt.Sprintf("error updating local state for dashboard: %s, id: %s", newLayoutError(w.Id), state.Id))
			return
		}

		// We found the layout that will give us the widget coordinates for comparison to the plan.
		layout := &dashboard.Layout[lIdx]
		isInState := false
		props, err := json.Marshal(w.Properties)
		if err != nil {
			diags.AddError("swo provider error",
				fmt.Sprintf("error updating local state for dashboard: %s, id: %s", newWidgetPropertiesError(err.Error(), w.Id), state.Id))
			return
		}

		for wIdx := range stateWidgets {
			stateW := &stateWidgets[wIdx]
			if stateW.Id.Equal(types.StringValue(w.Id)) {
				// Ensure the local widget state is aligned with what was returned by the server.
				isInState = true
				stateW.Type = types.StringValue(w.Type)
				stateW.X = types.Int64Value(int64(layout.X))
				stateW.Y = types.Int64Value(int64(layout.Y))
				stateW.Width = types.Int64Value(int64(layout.Width))
				stateW.Height = types.Int64Value(int64(layout.Height))

				var stateProps any
				err = json.Unmarshal([]byte(stateW.Properties.ValueString()), &stateProps)
				if err != nil {
					diags.AddError("swo provider error",
						fmt.Sprintf("error updating local state for dashboard: %s, id: %s", newWidgetPropertiesError(err.Error(), w.Id), state.Id))
					return
				}

				// The json string can be marshalled differently than what is specified in the terraform
				// file so we need to compare the marshalled values instead of the raw json string.
				if !cmp.Equal(&stateProps, w.Properties) {
					fmt.Println(cmp.Diff(&stateProps, w.Properties))
					stateW.Properties = types.StringValue(string(props))
				}

				break
			}
		}

		// If the terraform state doesn't have a widget returned from a Read we need to add it to align the
		// state with the server. This can happen if a dashboard is modified outside terraform (e.g., in the UI).
		if !isInState {
			stateWidgets = append(stateWidgets, dashboardWidgetModel{
				Id:         types.StringValue(w.Id),
				Type:       types.StringValue(w.Type),
				X:          types.Int64Value(int64(layout.X)),
				Y:          types.Int64Value(int64(layout.Y)),
				Width:      types.Int64Value(int64(layout.Width)),
				Height:     types.Int64Value(int64(layout.Height)),
				Properties: types.StringValue(string(props)),
			})
		}
	}

	updatedWidgets, d := types.SetValueFrom(ctx, types.ObjectType{AttrTypes: WidgetAttributeTypes()}, stateWidgets)
	diags.Append(d...)
	if diags.HasError() {
		return
	}
	state.Widgets = updatedWidgets
}

func (r *dashboardResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var tfPlan dashboardResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &tfPlan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tfVersion := tfPlan.Version.ValueInt32Pointer()
	var convertedTfVersion *int = nil
	if tfVersion != nil {
		temp := int(*tfVersion)
		convertedTfVersion = &temp
	}
	widgets, layouts := widgetsFromPlan(ctx, tfPlan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create the dashboard...
	dashboard, err := r.client.
		DashboardsService().
		Create(ctx, swoClient.CreateDashboardInput{
			Name:       tfPlan.Name.ValueString(),
			CategoryId: tfPlan.CategoryId.ValueStringPointer(),
			IsPrivate:  tfPlan.IsPrivate.ValueBoolPointer(),
			Widgets:    widgets,
			Layout:     layouts,
			Version:    convertedTfVersion,
		})

	if err != nil {
		resp.Diagnostics.AddError("swo provider error",
			fmt.Sprintf("create dashboard error: %s, name: %s", err, tfPlan.Name))
		return
	}

	setDashboardValuesFromCreate(ctx, dashboard, &tfPlan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &tfPlan)...)
}

func (r *dashboardResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var tfState dashboardResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &tfState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dId := tfState.Id.ValueString()

	// Read the dashboard...
	dashboard, err := r.client.DashboardsService().Read(ctx, dId)

	if err != nil {
		req.State.RemoveResource(ctx)
		resp.Diagnostics.AddError("swo provider error",
			fmt.Sprintf("read dashboard error: %s, id: %s", err, dId))
		return
	}

	setDashboardValuesFromRead(ctx, dashboard, &tfState, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &tfState)...)
}

func (r *dashboardResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state dashboardResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Computed value Id needs to be read from terraform state.
	id := state.Id.ValueString()

	tfVersion := plan.Version.ValueInt32Pointer()
	var convertedTfVersion *int = nil
	if tfVersion != nil {
		temp := int(*tfVersion)
		convertedTfVersion = &temp
	}
	widgets, layouts := widgetsFromPlan(ctx, plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update the dashboard...
	dashboard, err := r.client.DashboardsService().Update(ctx,
		swoClient.UpdateDashboardInput{
			Id:         id,
			Name:       plan.Name.ValueString(),
			CategoryId: plan.CategoryId.ValueStringPointer(),
			IsPrivate:  plan.IsPrivate.ValueBoolPointer(),
			Widgets:    widgets,
			Layout:     layouts,
			Version:    convertedTfVersion,
		})

	if err != nil {
		resp.Diagnostics.AddError("swo provider error",
			fmt.Sprintf("update dashboard error: %s, id: %s", err, id))
		return
	}

	// The create and update response objects are identical so we convert, so we don't have to have 2 separate
	// methods for 'setDashboardValuesFromCreate()'.
	d, err := convertObject[swoClient.CreateDashboardResult](dashboard)
	if err != nil {
		resp.Diagnostics.AddError("swo provider error",
			fmt.Sprintf("error setting computed values for dashboard: %s, id: %s", err, state.Id))
		return
	}

	setDashboardValuesFromCreate(ctx, d, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save to Terraform state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *dashboardResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state dashboardResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.Id.ValueString()

	// Delete the dashboard...
	err := r.client.DashboardsService().Delete(ctx, id)

	if err != nil {
		resp.State.RemoveResource(ctx)
		resp.Diagnostics.AddError("swo provider error",
			fmt.Sprintf("delete dashboard error: %s, id: %s", err, id))
	}
}

func (r *dashboardResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
