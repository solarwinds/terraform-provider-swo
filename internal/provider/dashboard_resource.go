package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	swoClient "github.com/solarwinds/swo-client-go/pkg/client"
	"golang.org/x/exp/slices"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &DashboardResource{}
var _ resource.ResourceWithConfigure = &DashboardResource{}
var _ resource.ResourceWithImportState = &DashboardResource{}

func NewDashboardResource() resource.Resource {
	return &DashboardResource{}
}

// Defines the resource implementation.
type DashboardResource struct {
	client *swoClient.Client
}

func (r *DashboardResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dashboard"
}

func (r *DashboardResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*swoClient.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"swo provider error",
			fmt.Sprintf("expected *swoClient.Client but received: %T", req.ProviderData),
		)
		return
	}

	r.client = client
}

// Creates new WidgetInputs and LayoutInputs from plan widget data.
func widgetsFromPlan(plan DashboardResourceModel) ([]swoClient.WidgetInput, []swoClient.LayoutInput, error) {
	widgets := make([]swoClient.WidgetInput, len(plan.Widgets))
	layouts := make([]swoClient.LayoutInput, len(plan.Widgets))

	for wIdx := range plan.Widgets {
		planW := &plan.Widgets[wIdx]
		id := uuid.NewString()

		// Marshal the json encoded properties string to an object.
		var props any
		err := json.Unmarshal([]byte(planW.Properties.ValueString()), &props)
		if err != nil {
			return nil, nil, err
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

	return widgets, layouts, nil
}

// Sets the computed values of the dashboard models with the values returned from the Create request.
func setDashboardValuesFromCreate(dashboard *swoClient.CreateDashboardResult, plan *DashboardResourceModel) error {
	plan.Id = types.StringValue(dashboard.Id)
	plan.CreatedAt = types.StringValue(dashboard.CreatedAt.Format(time.RFC3339))
	plan.UpdatedAt = types.StringValue(dashboard.UpdatedAt.Format(time.RFC3339))

	for _, w := range dashboard.Widgets {
		lIdx := slices.IndexFunc(dashboard.Layout, func(l swoClient.CreateDashboardLayout) bool { return l.Id == w.Id })
		if lIdx <= -1 {
			return fmt.Errorf("layout missing for widget. this may indicate a data intigrity problem on the server. looking for id: %s", w.Id)
		}

		// The layout that will give us the widget coordinates for comparison to the plan.
		layout := &dashboard.Layout[lIdx]

		// We need to compare the properties of the plan widget with what is returned from the server
		// to reconcile the server data with the plan data. Widgets ids in the plan are temporary and
		// there isn't any single value we can use to make a match.
		for wIdx := range plan.Widgets {
			planW := &plan.Widgets[wIdx]
			if planW.Type.Equal(types.StringValue(w.Type)) &&
				planW.X.Equal(types.Int64Value(int64(layout.X))) &&
				planW.Y.Equal(types.Int64Value(int64(layout.Y))) &&
				planW.Width.Equal(types.Int64Value(int64(layout.Width))) &&
				planW.Height.Equal(types.Int64Value(int64(layout.Height))) {
				// Widget properties are equal so we assume it must be the one we're looking for.
				planW.Id = types.StringValue(w.Id)
				break
			}
		}
	}

	return nil
}

// Sets the values of the terraform state with the values returned from the Read request.
func setDashboardValuesFromRead(dashboard *swoClient.ReadDashboardResult, state *DashboardResourceModel) error {
	state.Id = types.StringValue(dashboard.Id)
	state.Name = types.StringValue(dashboard.Name)
	state.CreatedAt = types.StringValue(dashboard.CreatedAt.Format(time.RFC3339))
	state.UpdatedAt = types.StringValue(dashboard.UpdatedAt.Format(time.RFC3339))
	if dashboard.Category != nil {
		state.CategoryId = types.StringValue(dashboard.Category.Id)
	}
	if dashboard.IsPrivate != nil {
		state.IsPrivate = types.BoolValue(*dashboard.IsPrivate)
	}

	for _, w := range dashboard.Widgets {
		lIdx := slices.IndexFunc(dashboard.Layout, func(l swoClient.ReadDashboardLayout) bool { return l.Id == w.Id })
		if lIdx <= -1 {
			return fmt.Errorf("layout missing for widget. this may indicate a data intigrity problem on the server. looking for id: %s", w.Id)
		}

		// We found the layout that will give us the widget coordinates for comparison to the plan.
		layout := &dashboard.Layout[lIdx]
		isInState := false
		props, err := json.Marshal(w.Properties)
		if err != nil {
			return fmt.Errorf("widget properties error: %s, id: %s",
				err, w.Id)
		}

		for wIdx := range state.Widgets {
			stateW := &state.Widgets[wIdx]
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
					return fmt.Errorf("widget properties error: %s, id: %s", err, w.Id)
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
		// state with the server. This can happen if a dashboard is modified outside of terraform (e.g. in the UI).
		if !isInState {
			state.Widgets = append(state.Widgets, DashboardWidgetModel{
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

	return nil
}

func (r *DashboardResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan DashboardResourceModel

	// Read the Terraform plan.
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	widgets, layouts, err := widgetsFromPlan(plan)
	if err != nil {
		resp.Diagnostics.AddError("swo provider error", fmt.Sprintf("convert plan to api error: %s, name: %s",
			err, plan.Name))
		return
	}

	// Create the dashboard...
	dashboard, err := r.client.
		DashboardsService().
		Create(ctx, swoClient.CreateDashboardInput{
			Name:       plan.Name.ValueString(),
			CategoryId: stringPtr(plan.CategoryId),
			IsPrivate:  boolPtr(plan.IsPrivate),
			Widgets:    widgets,
			Layout:     layouts,
		})

	if err != nil {
		resp.Diagnostics.AddError("swo provider error", fmt.Sprintf("create dashboard error: %s, name: %s",
			err, plan.Name))
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("dashboard create success: %s", dashboard.Id))

	err = setDashboardValuesFromCreate(dashboard, &plan)
	if err != nil {
		resp.Diagnostics.AddError("swo provider error", fmt.Sprintf("set dashboard computed values error: %s, id: %s",
			err, dashboard.Id))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *DashboardResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state DashboardResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dId := state.Id.ValueString()

	// Read the dashboard...
	tflog.Trace(ctx, fmt.Sprintf("read dashboard: %s", dId))
	dashboard, err := r.client.
		DashboardsService().
		Read(ctx, dId)

	if err != nil {
		req.State.RemoveResource(ctx)
		resp.Diagnostics.AddError("swo provider error", fmt.Sprintf("read dashboard error: %s, id: %s",
			err, dId))
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("read dashboard success: %s", dashboard.Id))
	err = setDashboardValuesFromRead(dashboard, &state)
	if err != nil {
		resp.Diagnostics.AddError("swo provider error", fmt.Sprintf("error updating local state for dashboard: %s, id: %s",
			err, state.Id))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *DashboardResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state DashboardResourceModel

	// Read the Terraform plan and state data.
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Computed values like Id need to be read from terraform state.
	id := state.Id.ValueString()

	widgets, layouts, err := widgetsFromPlan(plan)
	if err != nil {
		resp.Diagnostics.AddError("swo provider error", fmt.Sprintf("error converting plan to api: %s, name: %s",
			err, plan.Name))
		return
	}

	// Update the dashboard...
	tflog.Trace(ctx, fmt.Sprintf("update dashboard: %s", id))
	dashboard, err := r.client.
		DashboardsService().
		Update(ctx, swoClient.UpdateDashboardInput{
			Id:         id,
			Name:       plan.Name.ValueString(),
			CategoryId: stringPtr(plan.CategoryId),
			IsPrivate:  boolPtr(plan.IsPrivate),
			Widgets:    widgets,
			Layout:     layouts,
		})

	if err != nil {
		resp.Diagnostics.AddError("swo provider error", fmt.Sprintf("update dashboard error: %s, id: %s",
			err, id))
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("update dashboard success: %s", id))

	// The create and update response objects are identical so we convert so we don't have to have 2 separate
	// methods for 'setDashboardValuesFromCreate()'.
	d, err := convertObject[swoClient.CreateDashboardResult](dashboard)
	if err != nil {
		resp.Diagnostics.AddError("swo provider error", fmt.Sprintf("error setting computed values for dashboard: %s, id: %s",
			err, state.Id))
		return
	}

	err = setDashboardValuesFromCreate(d, &plan)
	if err != nil {
		resp.Diagnostics.AddError("swo provider error", fmt.Sprintf("error setting computed values for dashboard: %s, id: %s",
			err, state.Id))
		return
	}

	// Save to Terraform state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *DashboardResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state DashboardResourceModel

	// Read existing Terraform state.
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.Id.ValueString()

	// Delete the dashboard...
	tflog.Trace(ctx, fmt.Sprintf("delete dashboard: %s", id))
	err := r.client.
		DashboardsService().
		Delete(ctx, id)

	if err != nil {
		resp.State.RemoveResource(ctx)
		resp.Diagnostics.AddError("swo provider error", fmt.Sprintf("delete dashboard error: %s, id: %s",
			err, id))
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("delete dashboard success: %s", id))
}

func (r *DashboardResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
