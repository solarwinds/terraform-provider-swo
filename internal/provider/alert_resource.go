package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	swoClient "github.com/solarwinds/swo-client-go/pkg/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &AlertResource{}
var _ resource.ResourceWithImportState = &AlertResource{}

func NewAlertResource() resource.Resource {
	return &AlertResource{}
}

// ExampleResource defines the resource implementation.
type AlertResource struct {
	client *swoClient.Client
}

func (model *AlertResourceModel) ToAlertDefinitionInput() swoClient.AlertDefinitionInput {
	conditions := []swoClient.AlertConditionNodeInput{}

	for _, condition := range model.Conditions {
		conditions = condition.toAlertConditionInputs(conditions)
	}

	return swoClient.AlertDefinitionInput{
		Name:        model.Name.ValueString(),
		Description: model.Description.ValueStringPointer(),
		Enabled:     model.Enabled.ValueBool(),
		Severity:    swoClient.AlertSeverity(model.Severity.ValueString()),
		Actions:     model.toAlertActionInput(),
		Condition:   conditions,
	}
}

func (model *AlertResourceModel) toAlertActionInput() []swoClient.AlertActionInput {
	inputs := []swoClient.AlertActionInput{}
	actionTypes := map[string][]string{}

	for _, configId := range model.Notifications {
		actionId, actionType, found := strings.Cut(configId, ":")

		if found {
			if actionTypes[actionType] == nil {
				actionTypes[actionType] = []string{actionId}
			} else {
				actionTypes[actionType] = append(actionTypes[actionType], actionId)
			}
		}
	}

	for actionType, actionIds := range actionTypes {
		inputs = append(inputs, swoClient.AlertActionInput{
			Type:             actionType,
			ConfigurationIds: actionIds,
		})
	}

	return inputs
}

func (r *AlertResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_alert"
}

func (r *AlertResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	tflog.Trace(ctx, "AlertResource: Configure")

	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*swoClient.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Invalid Resource Client Type",
			fmt.Sprintf("Expected *swoClient.Client but received: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *AlertResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Trace(ctx, "AlertResource: Create")

	var tfModel *AlertResourceModel

	// Read the Terraform plan data into the model and log the results.
	resp.Diagnostics.Append(req.Plan.Get(ctx, &tfModel)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Create the alert from the provided Terraform model...
	input := tfModel.ToAlertDefinitionInput()
	newAlertDef, err := r.client.AlertsService().Create(ctx, input)

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Error creating alert definition '%s'. Error: %s",
			input.Name,
			err))
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("Alert definition '%s' created successfully. ID: %s", newAlertDef.Name, newAlertDef.Id))

	tfModel.ID = types.StringValue(newAlertDef.Id)

	resp.Diagnostics.Append(resp.State.Set(ctx, &tfModel)...)
}

func (r *AlertResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Trace(ctx, "AlertResource: Read")

	var tfModel *AlertResourceModel

	// Read any existing Terraform state into the model.
	resp.Diagnostics.Append(req.State.Get(ctx, &tfModel)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("Getting alert with ID: %s", tfModel.ID))

	alertDef, err := r.client.AlertsService().Read(ctx, tfModel.ID.String())

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Error getting alert %s. Error: %s",
			alertDef.Id,
			err))
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("Alert received: %s", alertDef.Name))

	resp.Diagnostics.Append(resp.State.Set(ctx, &tfModel)...)
}

func (r *AlertResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Trace(ctx, "AlertResource: Update")

	var tfModel, tfState *AlertResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &tfModel)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &tfState)...)

	if resp.Diagnostics.HasError() {
		return
	}

	alertId := tfState.ID.ValueString()
	input := tfModel.ToAlertDefinitionInput()

	tflog.Trace(ctx, fmt.Sprintf("Updating alert definition with ID: %s", alertId))

	// Update the alert definition...
	_, err := r.client.AlertsService().Update(ctx, alertId, input)

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Error updating alert definition %s. Error: %s", alertId, err))
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("Alert definition '%s' updated successfully.", input.Name))

	//Because Id is computed it needs to be set on the model before applying.
	tfModel.ID = types.StringValue(alertId)

	// Save and log the model into Terraform state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &tfModel)...)
}

func (r *AlertResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Trace(ctx, "AlertResource: Delete")
	var tfState *AlertResourceModel

	// Read from terraform state becasue Id is computed.
	resp.Diagnostics.Append(req.State.Get(ctx, &tfState)...)

	if resp.Diagnostics.HasError() {
		return
	}

	alertDefId := tfState.ID.ValueString()

	tflog.Trace(ctx, fmt.Sprintf("Deleting alert definition with ID: %s", alertDefId))

	// Delete the alert definition...
	err := r.client.AlertsService().Delete(ctx, alertDefId)

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Error deleting alert definition %s. Error: %s", alertDefId, err))
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("Alert definition deleted: %s", alertDefId))
}

func (r *AlertResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tflog.Trace(ctx, "AlertResource: ImportState")
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
