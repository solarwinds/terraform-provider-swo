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
var _ resource.Resource = &ApiTokenResource{}
var _ resource.ResourceWithConfigure = &ApiTokenResource{}
var _ resource.ResourceWithImportState = &ApiTokenResource{}

func NewApiTokenResource() resource.Resource {
	return &ApiTokenResource{}
}

// Defines the resource implementation.
type ApiTokenResource struct {
	client *swoClient.Client
}

func (r *ApiTokenResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_apitoken"
}

func (r *ApiTokenResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	tflog.Trace(ctx, "ApiTokenResource: Configure")

	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	if client, ok := req.ProviderData.(*swoClient.Client); !ok {
		resp.Diagnostics.AddError(
			"Invalid Resource Client Type",
			fmt.Sprintf("expected *swoClient.Client but received: %T.", req.ProviderData),
		)
		return
	} else {
		r.client = client
	}
}

func (r *ApiTokenResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Trace(ctx, "ApiTokenResource: Create")

	var tfPlan ApiTokenResourceModel

	// Read the Terraform plan.
	resp.Diagnostics.Append(req.Plan.Get(ctx, &tfPlan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create our input request.
	createInput := swoClient.CreateTokenInput{
		Name:        tfPlan.Name.ValueString(),
		AccessLevel: *tfPlan.AccessLevel,
		Type:        tfPlan.Type.ValueStringPointer(),
		Attributes: convertArray(tfPlan.Attributes, func(v ApiTokenAttribute) swoClient.TokenAttributeInput {
			return swoClient.TokenAttributeInput{
				Key:   v.Key.ValueString(),
				Value: v.Value.ValueString(),
			}
		}),
	}

	// Create the ApiToken...
	newApiToken, err := r.client.ApiTokenService().Create(ctx, createInput)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("error creating apiToken '%s' - error: %s",
			tfPlan.Name,
			err))
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("apiToken %s created successfully: id=%s", tfPlan.Name, newApiToken.Id))
	tfPlan.Id = types.StringValue(newApiToken.Id)
	tfPlan.Token = types.StringPointerValue(newApiToken.Token)
	tfPlan.Secure = types.BoolValue(true)

	resp.Diagnostics.Append(resp.State.Set(ctx, tfPlan)...)
}

func (r *ApiTokenResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Trace(ctx, "ApiTokenResource: Read")

	var tfState ApiTokenResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &tfState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read the ApiToken...
	tflog.Trace(ctx, fmt.Sprintf("read apiToken: id=%s", tfState.Id))
	apiToken, err := r.client.ApiTokenService().Read(ctx, tfState.Id.ValueString())

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("error reading apiToken %s. error: %s",
			tfState.Id, err))
		return
	}

	// Update the Terraform state.
	tfState.Id = types.StringValue(apiToken.Id)
	tfState.Name = types.StringPointerValue(apiToken.Name)
	tfState.Enabled = types.BoolPointerValue(apiToken.Enabled)
	tfState.Type = types.StringPointerValue(apiToken.Type)
	tfState.Secure = types.BoolPointerValue(apiToken.Secure)
	tfState.AccessLevel = apiToken.AccessLevel

	// Attributes
	var attrs = []ApiTokenAttribute{}
	for _, attr := range apiToken.Attributes {
		attrs = append(attrs, ApiTokenAttribute{
			Key:   types.StringValue(attr.Key),
			Value: types.StringValue(attr.Value),
		})
	}
	tfState.Attributes = attrs

	tflog.Trace(ctx, fmt.Sprintf("read apiToken success: id=%s", apiToken.Id))
	resp.Diagnostics.Append(resp.State.Set(ctx, tfState)...)
}

func (r *ApiTokenResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var tfPlan, tfState *ApiTokenResourceModel

	// Read the Terraform plan and state data.
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
		Attributes: convertArray(tfPlan.Attributes, func(v ApiTokenAttribute) swoClient.TokenAttributeInput {
			return swoClient.TokenAttributeInput{
				Key:   v.Key.ValueString(),
				Value: v.Value.ValueString(),
			}
		}),
	}

	// Update the ApiToken...
	tflog.Trace(ctx, fmt.Sprintf("updating apiToken with id: %s", tfState.Id))
	err := r.client.ApiTokenService().Update(ctx, updateInput)

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("error updating apiToken %s. err: %s", tfState.Id, err))
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("apiToken '%s' updated successfully", tfState.Id))

	// Save to Terraform state.
	tfPlan.Id = tfState.Id
	tfPlan.Token = tfState.Token
	tfPlan.Secure = tfState.Secure

	resp.Diagnostics.Append(resp.State.Set(ctx, &tfPlan)...)
}

func (r *ApiTokenResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ApiTokenResourceModel

	// Read existing Terraform state.
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the ApiToken...
	tflog.Trace(ctx, fmt.Sprintf("deleting apiToken: id=%s", state.Id))
	if err := r.client.ApiTokenService().Delete(ctx, state.Id.ValueString()); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("error deleting apiToken %s - %s", state.Id, err))
		return
	}
	tflog.Trace(ctx, fmt.Sprintf("apiToken deleted: id=%s", state.Id))
}

func (r *ApiTokenResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
