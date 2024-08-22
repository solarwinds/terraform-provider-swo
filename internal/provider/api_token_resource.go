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
	_ resource.Resource                = &apiTokenResource{}
	_ resource.ResourceWithConfigure   = &apiTokenResource{}
	_ resource.ResourceWithImportState = &apiTokenResource{}
)

func NewApiTokenResource() resource.Resource {
	return &apiTokenResource{}
}

// Defines the resource implementation.
type apiTokenResource struct {
	client *swoClient.Client
}

func (r *apiTokenResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "apitoken"
}

func (r *apiTokenResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, _ := req.ProviderData.(*swoClient.Client)
	r.client = client
}

func (r *apiTokenResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var tfPlan apiTokenResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &tfPlan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create our input request.
	createInput := swoClient.CreateTokenInput{
		Name:        tfPlan.Name.ValueString(),
		AccessLevel: *tfPlan.AccessLevel,
		Type:        tfPlan.Type.ValueStringPointer(),
		Attributes: convertArray(tfPlan.Attributes, func(v apiTokenAttribute) swoClient.TokenAttributeInput {
			return swoClient.TokenAttributeInput{
				Key:   v.Key.ValueString(),
				Value: v.Value.ValueString(),
			}
		}),
	}

	// Create the ApiToken...
	newApiToken, err := r.client.ApiTokenService().Create(ctx, createInput)
	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error creating apiToken '%s' - error: %s", tfPlan.Name, err))
		return
	}

	tfPlan.Id = types.StringValue(newApiToken.Id)
	tfPlan.Token = types.StringPointerValue(newApiToken.Token)

	resp.Diagnostics.Append(resp.State.Set(ctx, tfPlan)...)
}

func (r *apiTokenResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var tfState apiTokenResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &tfState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read the ApiToken...
	apiToken, err := r.client.ApiTokenService().Read(ctx, tfState.Id.ValueString())

	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error reading API token %s. error: %s", tfState.Id, err))
		return
	}

	r.updateState(&tfState, apiToken)
	resp.Diagnostics.Append(resp.State.Set(ctx, tfState)...)
}

func (r *apiTokenResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var tfPlan, tfState *apiTokenResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &tfPlan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &tfState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create our input request.
	updateInput := swoClient.UpdateTokenInput{
		Id:          tfState.Id.ValueString(),
		Name:        tfPlan.Name.ValueStringPointer(),
		Enabled:     tfPlan.Enabled.ValueBoolPointer(),
		Type:        tfPlan.Type.ValueStringPointer(),
		AccessLevel: tfPlan.AccessLevel,
		Attributes: convertArray(tfPlan.Attributes, func(v apiTokenAttribute) swoClient.TokenAttributeInput {
			return swoClient.TokenAttributeInput{
				Key:   v.Key.ValueString(),
				Value: v.Value.ValueString(),
			}
		}),
	}

	// Update the ApiToken...
	err := r.client.ApiTokenService().Update(ctx, updateInput)
	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error updating API token %s. err: %s", tfState.Id, err))
		return
	}

	// Save to Terraform state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &tfPlan)...)
}

func (r *apiTokenResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var tfState apiTokenResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &tfState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the ApiToken...
	if err := r.client.ApiTokenService().Delete(ctx, tfState.Id.ValueString()); err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error deleting API token %s - %s", tfState.Id, err))
	}
}

func (r *apiTokenResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *apiTokenResource) updateState(state *apiTokenResourceModel, result *swoClient.ReadApiTokenResult) {
	state.Id = types.StringValue(result.Id)
	state.Name = types.StringPointerValue(result.Name)
	state.Enabled = types.BoolPointerValue(result.Enabled)
	state.Type = types.StringPointerValue(result.Type)
	state.AccessLevel = result.AccessLevel

	var attrs = []apiTokenAttribute{}
	for _, attr := range result.Attributes {
		attrs = append(attrs, apiTokenAttribute{
			Key:   types.StringValue(attr.Key),
			Value: types.StringValue(attr.Value),
		})
	}
	state.Attributes = attrs
}
