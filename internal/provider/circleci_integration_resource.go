package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	swoClient "github.com/solarwinds/swo-client-go/pkg/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &circleCIIntegrationResource{}
	_ resource.ResourceWithConfigure   = &circleCIIntegrationResource{}
	_ resource.ResourceWithImportState = &circleCIIntegrationResource{}
)

func NewCircleCIIntegrationResource() resource.Resource {
	return &circleCIIntegrationResource{}
}

// Defines the resource implementation.
type circleCIIntegrationResource struct {
	client *swoClient.Client
}

func (r *circleCIIntegrationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "circleciintegration"
}

func (r *circleCIIntegrationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, _ := req.ProviderData.(providerClients)
	r.client = client.SwoClient
}

func (r *circleCIIntegrationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var tfPlan circleCIIntegrationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &tfPlan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.CircleCIIntegrationService().Create(
		ctx,
		tfPlan.Name.ValueString(),
		tfPlan.ApiToken.ValueStringPointer(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error creating CircleCI integration '%s' - error: %s", tfPlan.Name.ValueString(), err))
		return
	}

	tfPlan.Id = types.StringValue(result.Id)
	tfPlan.SecretToken = types.StringValue(result.SecretToken)
	tfPlan.ReceiverUrl = types.StringValue(fmt.Sprintf("%s?state=%s", tfPlan.ReceiverBase.ValueString(), result.Id))

	resp.Diagnostics.Append(resp.State.Set(ctx, tfPlan)...)
}

func (r *circleCIIntegrationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var tfState circleCIIntegrationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &tfState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.CircleCIIntegrationService().Read(ctx, tfState.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error reading CircleCI integration %s. error: %s", tfState.Id.ValueString(), err))
		return
	}

	tfState.Name = types.StringValue(result.Name)
	tfState.SecretToken = types.StringValue(result.SecretToken)
	tfState.ReceiverUrl = types.StringValue(fmt.Sprintf("%s?state=%s", tfState.ReceiverBase.ValueString(), result.Id))

	resp.Diagnostics.Append(resp.State.Set(ctx, tfState)...)
}

func (r *circleCIIntegrationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var tfPlan, tfState circleCIIntegrationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &tfPlan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &tfState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.CircleCIIntegrationService().Update(
		ctx,
		tfState.Id.ValueString(),
		tfPlan.Name.ValueStringPointer(),
		tfPlan.ApiToken.ValueStringPointer(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error updating CircleCI integration %s. err: %s", tfState.Id.ValueString(), err))
		return
	}

	tfPlan.Id = types.StringValue(result.Id)
	tfPlan.SecretToken = types.StringValue(result.SecretToken)
	tfPlan.ReceiverUrl = types.StringValue(fmt.Sprintf("%s?state=%s", tfPlan.ReceiverBase.ValueString(), result.Id))

	resp.Diagnostics.Append(resp.State.Set(ctx, &tfPlan)...)
}

func (r *circleCIIntegrationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var tfState circleCIIntegrationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &tfState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.CircleCIIntegrationService().Delete(ctx, tfState.Id.ValueString()); err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error deleting CircleCI integration %s - %s", tfState.Id.ValueString(), err))
	}
}

func (r *circleCIIntegrationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
