package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	swoClient "github.com/solarwinds/swo-client-go/pkg/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &LogFilterResource{}
var _ resource.ResourceWithConfigure = &LogFilterResource{}
var _ resource.ResourceWithImportState = &LogFilterResource{}

func NewLogFilterResource() resource.Resource {
	return &LogFilterResource{}
}

// Defines the resource implementation.
type LogFilterResource struct {
	client *swoClient.Client
}

func (r *LogFilterResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_logfilter"
}

func (r *LogFilterResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	tflog.Trace(ctx, "LogFilterResource: Configure")

	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*swoClient.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Invalid Resource Client Type",
			fmt.Sprintf("expected *swoClient.Client but received: %T.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *LogFilterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Trace(ctx, "LogFilterResource: Create")

	var tfPlan LogFilterResourceModel

	// Read the Terraform plan.
	resp.Diagnostics.Append(req.Plan.Get(ctx, &tfPlan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create our input request.
	createInput := swoClient.CreateExclusionFilterInput{
		Name:           tfPlan.Name.ValueString(),
		Description:    tfPlan.Description.ValueString(),
		TokenSignature: tfPlan.TokenSignature,
		Expressions: convertArray(tfPlan.Expressions, func(e LogFilterExpression) swoClient.CreateExclusionFilterExpressionInput {
			return swoClient.CreateExclusionFilterExpressionInput{
				Kind:       swoClient.ExclusionFilterExpressionKind(e.Kind),
				Expression: e.Expression,
			}
		}),
	}

	// Create the LogFilter...
	result, err := r.client.LogFilterService().Create(ctx, createInput)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("error creating logFilter '%s' - error: %s",
			tfPlan.Name,
			err))
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("logFilter %s created successfully - id=%s", tfPlan.Name, result.Id))
	tfPlan.Id = types.StringValue(result.Id)
	resp.Diagnostics.Append(resp.State.Set(ctx, tfPlan)...)
}

func (r *LogFilterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Trace(ctx, "LogFilterResource: Read")

	var tfState LogFilterResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &tfState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read the LogFilter...
	tflog.Trace(ctx, fmt.Sprintf("read logFilter with id: %s", tfState.Id))
	logFilter, err := r.client.LogFilterService().Read(ctx, tfState.Id.ValueString())

	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error reading logFilter %s. error: %s", tfState.Id, err))
		return
	} else if logFilter == nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("logFilter not found. id=%s", tfState.Id))
		return
	}

	// Update the Terraform state with latest values from the server.
	tfState.Name = types.StringValue(logFilter.Name)
	tfState.Description = types.StringValue(*logFilter.Description)
	tfState.TokenSignature = logFilter.TokenSignature

	var lfe []LogFilterExpression
	for _, p := range logFilter.Expressions {
		lfe = append(lfe, LogFilterExpression{
			Kind:       swoClient.ExclusionFilterExpressionKind(p.Kind),
			Expression: p.Expression,
		})
	}
	tfState.Expressions = lfe

	// Save to Terraform state.
	tflog.Trace(ctx, fmt.Sprintf("read logFilter success: %s", logFilter.Name))
	resp.Diagnostics.Append(resp.State.Set(ctx, tfState)...)
}

func (r *LogFilterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var tfPlan, tfState *LogFilterResourceModel

	// Read the Terraform plan and state data.
	resp.Diagnostics.Append(req.Plan.Get(ctx, &tfPlan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &tfState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Update the LogFilter...
	tflog.Trace(ctx, fmt.Sprintf("updating logFilter with id: %s", tfState.Id))
	err := r.client.LogFilterService().Update(ctx, swoClient.UpdateExclusionFilterInput{
		Id:          tfState.Id.ValueString(),
		Name:        tfPlan.Name.ValueString(),
		Description: tfPlan.Description.ValueString(),
		Expressions: convertArray(tfPlan.Expressions, func(e LogFilterExpression) swoClient.UpdateExclusionFilterExpressionInput {
			return swoClient.UpdateExclusionFilterExpressionInput{
				Kind:       swoClient.ExclusionFilterExpressionKind(e.Kind),
				Expression: e.Expression,
			}
		}),
	})

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("error updating logFilter %s. err: %s", tfState.Id, err))
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("logFilter '%s' updated successfully", tfState.Id))

	// Save to Terraform state.
	tfPlan.Id = tfState.Id
	resp.Diagnostics.Append(resp.State.Set(ctx, &tfPlan)...)
}

func (r *LogFilterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var tfState LogFilterResourceModel

	// Read existing Terraform state.
	resp.Diagnostics.Append(req.State.Get(ctx, &tfState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// // Delete the LogFilter...
	tflog.Trace(ctx, fmt.Sprintf("deleting logFilter: id=%s", tfState.Id))
	if err := r.client.LogFilterService().
		Delete(ctx, tfState.Id.ValueString()); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("error deleting logFilter %s - %s", tfState.Id, err))
		return
	}
	tflog.Trace(ctx, fmt.Sprintf("logFilter deleted: id=%s", tfState.Id))
}

func (r *LogFilterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
