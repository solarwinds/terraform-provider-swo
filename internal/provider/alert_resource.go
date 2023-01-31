package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	swoClient "github.com/solarwindscloud/terraform-provider-swo/internal/client"
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

func (model *AlertResourceModel) ToCreateAlertDefinition() swoClient.CreateAlertDefinitionAlertMutationsCreateAlertDefinition {
	return swoClient.CreateAlertDefinitionAlertMutationsCreateAlertDefinition{
		Id:                  model.ID.String(),
		Name:                model.Name.String(),
		Description:         model.Description.String(),
		Enabled:             model.Enabled.ValueBool(),
		Severity:            swoClient.AlertSeverity(model.Severity.String()),
		Actions:             []swoClient.CreateAlertDefinitionAlertMutationsCreateAlertDefinitionActionsAlertAction{},
		TriggerResetActions: model.TriggerResetActions.ValueBool(),
		FlatCondition:       []swoClient.CreateAlertDefinitionAlertMutationsCreateAlertDefinitionFlatConditionFlatAlertConditionExpression{},
	}
}

func (model *AlertResourceModel) ToUpdateAlertDefinition() swoClient.UpdateAlertDefinitionAlertMutationsUpdateAlertDefinition {
	return swoClient.UpdateAlertDefinitionAlertMutationsUpdateAlertDefinition{
		Id:                  model.ID.String(),
		Name:                model.Name.String(),
		Description:         model.Description.String(),
		Enabled:             model.Enabled.ValueBool(),
		Severity:            swoClient.AlertSeverity(model.Severity.String()),
		Actions:             []swoClient.UpdateAlertDefinitionAlertMutationsUpdateAlertDefinitionActionsAlertAction{},
		TriggerResetActions: model.TriggerResetActions.ValueBool(),
		//FlatCondition:       []swoClient.UpdateAlertDefinitionAlertMutationsUpdateAlertDefinitionFlatConditionFlatAlertConditionExpression{},
	}
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
	alertDef := tfModel.ToCreateAlertDefinition()
	newAlertDef, err := r.client.AlertsService().Create(alertDef)

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Error creating alert definition '%s'. Error: %s",
			alertDef.Name,
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

	alertDef, err := r.client.AlertsService().Read(tfModel.ID.String())

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Error getting alert %s. Error: %s",
			alertDef.Id,
			err))
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("Alert received: %s", alertDef.Name))

	tfModel.Name = types.StringValue("Mock Alert Name")
	tfModel.Description = types.StringValue("Mock alert description.")
	tfModel.Severity = types.StringValue("CRITICAL")
	tfModel.Type = types.StringValue("ENTITY_METRIC")
	tfModel.TargetEntityTypes = []string{"Website"}
	tfModel.Enabled = types.BoolValue(true)
	tfModel.Conditions = []AlertConditionModel{
		{
			MetricName:      types.StringValue("synthetics.https.response.time"),
			Threshold:       types.StringValue(">=3000ms"),
			Duration:        types.StringValue("5m"),
			AggregationType: types.StringValue("AVG"),
			EntityIds: []string{
				"e-1521946194448543744",
				"e-1521947552186691584",
			},
			IncludeTags: &[]AlertTagsModel{
				{
					Name:   types.StringValue("probe.city"),
					Values: []string{"Tokyo", "Sao Paulo"},
				},
			},
		},
	}
	tfModel.Notifications = []int{
		123,
		456,
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &tfModel)...)
}

func (r *AlertResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Trace(ctx, "AlertResource: Update")

	var tfModel *AlertResourceModel

	// Read the Terraform plan data into the model.
	resp.Diagnostics.Append(req.Plan.Get(ctx, &tfModel)...)

	if resp.Diagnostics.HasError() {
		return
	}

	alertDef := tfModel.ToUpdateAlertDefinition()

	tflog.Trace(ctx, fmt.Sprintf("Updating alert definition with ID: %s", alertDef.Id))

	// Update the alert definition...
	err := r.client.AlertsService().Update(alertDef)

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Error updating alert definition %s. Error: %s", alertDef.Id, err))
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("Alert definition '%s' updated successfully.", alertDef.Name))

	tfModel.Name = types.StringValue("Mock Alert Name")
	tfModel.Description = types.StringValue("Mock alert description.")
	tfModel.Severity = types.StringValue("CRITICAL")
	tfModel.Type = types.StringValue("METRIC")
	tfModel.TargetEntityTypes = []string{"Website"}
	tfModel.Enabled = types.BoolValue(true)
	tfModel.Conditions = []AlertConditionModel{
		{
			MetricName:      types.StringValue("synthetics.https.response.time"),
			Threshold:       types.StringValue(">=3000ms"),
			Duration:        types.StringValue("5m"),
			AggregationType: types.StringValue("AVG"),
			EntityIds: []string{
				"e-1521946194448543744",
				"e-1521947552186691584",
			},
			IncludeTags: &[]AlertTagsModel{
				{
					Name:   types.StringValue("probe.city"),
					Values: []string{"Tokyo", "Sao Paulo"},
				},
			},
		},
	}
	tfModel.Notifications = []int{
		123,
		456,
	}

	// Save and log the model into Terraform state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &tfModel)...)
}

func (r *AlertResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Trace(ctx, "AlertResource: Delete")
	var tfModel *AlertResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &tfModel)...)

	if resp.Diagnostics.HasError() {
		return
	}

	alertDefId := tfModel.ID.String()

	tflog.Trace(ctx, fmt.Sprintf("Deleting alert definition with ID: %s", alertDefId))

	// Delete the alert definition...
	err := r.client.AlertsService().Delete(alertDefId)

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
