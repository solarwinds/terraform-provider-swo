package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/solarwinds/swo-sdk-go/swov1"
	"github.com/solarwinds/swo-sdk-go/swov1/models/components"
	"github.com/solarwinds/swo-sdk-go/swov1/models/operations"
)

const (
	clientErrSummary = "SwoV1 Client Error"
	compositePrefix  = "composite."
)

var (
	_ resource.Resource                = &compositeMetricResource{}
	_ resource.ResourceWithConfigure   = &compositeMetricResource{}
	_ resource.ResourceWithImportState = &compositeMetricResource{}
)

type compositeMetricResource struct {
	client *swov1.Swo
}

func (r *compositeMetricResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "compositemetric"
}

func NewCompositeMetricResource() resource.Resource {
	return &compositeMetricResource{}
}

func (r *compositeMetricResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, _ := req.ProviderData.(providerClients)
	r.client = client.SwoV1Client
}

func (r *compositeMetricResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var tfPlan compositeMetricResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &tfPlan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hasPrefix := strings.HasPrefix(tfPlan.Name.ValueString(), compositePrefix)

	if !hasPrefix {
		resp.Diagnostics.AddError(clientErrSummary,
			fmt.Sprintf("error creating composite metric '%s' - error: metric name is invalid, missing 'composite.' prefix.", tfPlan.Name))
		return
	}

	input := &components.CompositeMetric{
		Name:        tfPlan.Name.ValueString(),
		DisplayName: tfPlan.DisplayName.ValueStringPointer(),
		Description: tfPlan.Description.ValueStringPointer(),
		Formula:     tfPlan.Formula.ValueString(),
		Units:       tfPlan.Unit.ValueStringPointer(),
	}

	res, err := r.client.Metrics.CreateCompositeMetric(ctx, *input)

	if err != nil {
		resp.Diagnostics.AddError(clientErrSummary,
			fmt.Sprintf("error creating composite metric '%s' - error: %s", tfPlan.Name, err))
		return
	}

	if res.CompositeMetric == nil {
		resp.Diagnostics.AddError("Empty Response",
			fmt.Sprintf("create composite metric response was empty '%s'", tfPlan.Name))
		return
	}

	tfPlan = r.updatePlanMetricInfo(tfPlan, res.CompositeMetric)
	resp.Diagnostics.Append(resp.State.Set(ctx, tfPlan)...)
}

func (r *compositeMetricResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var tfPlan compositeMetricResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &tfPlan)...)

	res, err := r.client.Metrics.GetMetricByName(ctx, operations.GetMetricByNameRequest{
		Name: tfPlan.Name.ValueString(),
	})

	compositeMetric := res.CommonMetricInfo

	if err != nil {
		resp.Diagnostics.AddError(clientErrSummary,
			fmt.Sprintf("error reading composite metric '%s' - error: %s", tfPlan.Name, err))
		return
	}

	if compositeMetric == nil {
		resp.Diagnostics.AddError("Empty Response",
			fmt.Sprintf("read composite metric response was empty '%s'", tfPlan.Name))
		return
	}

	tfPlan = r.updatePlanCommonMetricInfo(tfPlan, compositeMetric)
	resp.Diagnostics.Append(resp.State.Set(ctx, tfPlan)...)
}

func (r *compositeMetricResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var tfPlan compositeMetricResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &tfPlan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := operations.UpdateCompositeMetricRequest{
		Name: tfPlan.Name.ValueString(),
		UpdateCompositeMetric: components.UpdateCompositeMetric{
			DisplayName: tfPlan.DisplayName.ValueStringPointer(),
			Description: tfPlan.Description.ValueStringPointer(),
			Formula:     tfPlan.Formula.ValueString(),
			Units:       tfPlan.Unit.ValueStringPointer(),
		},
	}

	res, err := r.client.Metrics.UpdateCompositeMetric(ctx, input)

	if err != nil {
		resp.Diagnostics.AddError(clientErrSummary,
			fmt.Sprintf("error updating composite metric '%s' - error: %v", tfPlan.Name, err))
		return
	}

	if res == nil {
		resp.Diagnostics.AddError("Empty Response",
			fmt.Sprintf("update composite metric response was empty '%s'", tfPlan.Name))
		return
	}

	tfPlan = r.updatePlanMetricInfo(tfPlan, res.CompositeMetric)
	resp.Diagnostics.Append(resp.State.Set(ctx, &tfPlan)...)
}

func (r *compositeMetricResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var tfPlan compositeMetricResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &tfPlan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Metrics.DeleteCompositeMetric(ctx, operations.DeleteCompositeMetricRequest{
		Name: tfPlan.Name.ValueString(),
	})

	if err != nil {
		resp.Diagnostics.AddError(clientErrSummary,
			fmt.Sprintf("error deleting composite metric '%s' - error: %s", tfPlan.Name, err))
		return
	}
}

func (r *compositeMetricResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

func (r *compositeMetricResource) updatePlanMetricInfo(tfPlan compositeMetricResourceModel, compositeMetric *components.CompositeMetric) compositeMetricResourceModel {
	tfPlan.Name = types.StringValue(compositeMetric.Name)
	tfPlan.Id = tfPlan.Name
	tfPlan.DisplayName = types.StringValue(*compositeMetric.DisplayName)
	tfPlan.Description = types.StringValue(*compositeMetric.Description)
	tfPlan.Formula = types.StringValue(compositeMetric.Formula)
	tfPlan.Unit = types.StringValue(*compositeMetric.Units)

	return tfPlan
}

func (r *compositeMetricResource) updatePlanCommonMetricInfo(tfPlan compositeMetricResourceModel, compositeMetric *components.CommonMetricInfo) compositeMetricResourceModel {
	tfPlan.Name = types.StringValue(compositeMetric.Name)
	tfPlan.Id = tfPlan.Name
	tfPlan.DisplayName = types.StringValue(*compositeMetric.DisplayName)
	tfPlan.Description = types.StringValue(*compositeMetric.Description)
	tfPlan.Formula = types.StringValue(*compositeMetric.Formula)
	tfPlan.Unit = types.StringValue(*compositeMetric.Units)

	return tfPlan
}
