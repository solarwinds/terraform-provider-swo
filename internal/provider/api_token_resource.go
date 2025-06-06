package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"

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
	client, _ := req.ProviderData.(providerClients)
	r.client = client.SwoClient
}

func (r *apiTokenResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var tfPlan apiTokenResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &tfPlan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var attributes []apiTokenAttribute
	d := tfPlan.Attributes.ElementsAs(ctx, &attributes, false)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create our input request.
	createInput := swoClient.CreateTokenInput{
		Name:        tfPlan.Name.ValueString(),
		AccessLevel: swoClient.TokenAccessLevel(tfPlan.AccessLevel.ValueString()),
		Type:        tfPlan.Type.ValueStringPointer(),
		Attributes: convertArray(attributes, func(v apiTokenAttribute) swoClient.TokenAttributeInput {
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

	r.updateState(ctx, &tfState, apiToken, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, tfState)...)
}

func (r *apiTokenResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var tfPlan, tfState *apiTokenResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &tfPlan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &tfState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var attributes []apiTokenAttribute
	d := tfPlan.Attributes.ElementsAs(ctx, &attributes, false)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create our input request.
	updateInput := swoClient.UpdateTokenInput{
		Id:          tfState.Id.ValueString(),
		Name:        tfPlan.Name.ValueStringPointer(),
		Enabled:     tfPlan.Enabled.ValueBoolPointer(),
		Type:        tfPlan.Type.ValueStringPointer(),
		AccessLevel: (*swoClient.TokenAccessLevel)(tfPlan.AccessLevel.ValueStringPointer()),
		Attributes: convertArray(attributes, func(v apiTokenAttribute) swoClient.TokenAttributeInput {
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

func (r *apiTokenResource) updateState(ctx context.Context, state *apiTokenResourceModel, result *swoClient.ReadApiTokenResult, diags *diag.Diagnostics) {
	state.Id = types.StringValue(result.Id)
	state.Name = types.StringPointerValue(result.Name)
	state.Enabled = types.BoolPointerValue(result.Enabled)
	state.Type = types.StringPointerValue(result.Type)
	state.AccessLevel = types.StringValue(string(*result.AccessLevel))

	var elements []attr.Value
	var attributeTypes = TokenAttributeTypes()
	for _, attribute := range result.Attributes {
		objectValue, d := types.ObjectValueFrom(
			ctx,
			attributeTypes,
			apiTokenAttribute{
				Key:   types.StringValue(attribute.Key),
				Value: types.StringValue(attribute.Value),
			},
		)

		diags.Append(d...)
		if diags.HasError() {
			return
		}
		elements = append(elements, objectValue)
	}

	attributes, d := types.SetValueFrom(ctx, types.ObjectType{AttrTypes: attributeTypes}, elements)
	diags.Append(d...)
	if diags.HasError() {
		return
	}

	state.Attributes = attributes
}
