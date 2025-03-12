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
	_ resource.Resource              = &alertResource{}
	_ resource.ResourceWithConfigure = &alertResource{}
	// TODO: Uncomment this once ImportState is implemented.
	// _ resource.ResourceWithImportState = &alertResource{}
)

func NewAlertResource() resource.Resource {
	return &alertResource{}
}

type alertResource struct {
	client *swoClient.Client
}

func (r *alertResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "alert"
}

func (r *alertResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, _ := req.ProviderData.(*swoClient.Client)
	r.client = client
}

func (r *alertResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data alertResourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if len(data.Conditions) > 1 {
		resp.Diagnostics.AddAttributeError(
			path.Root("conditions"),
			"More than one condition.",
			"Cannot support more than one condition at this time.",
		)
	}

	for _, condition := range data.Conditions {
		// Validation if not_reporting = true
		notReporting := condition.NotReporting.ValueBool()
		if notReporting {
			// Can't use threshold in the same condition
			threshold := condition.Threshold.ValueString()
			if threshold != "" {
				resp.Diagnostics.AddAttributeError(
					path.Root("threshold"),
					"Cannot set threshold when not_reporting is set to true.",
					"Cannot set threshold when not_reporting is set to true.",
				)
			}

			// Aggregation must be count
			operator := condition.AggregationType.ValueString()
			operatorType, _ := swoClient.GetAlertConditionType(operator)
			if operatorType == string(swoClient.AlertAggregationOperatorType) && operator != string(swoClient.AlertOperatorCount) {
				resp.Diagnostics.AddAttributeError(
					path.Root("aggregationType"),
					"Aggregation type must be COUNT when not_reporting is set to true.",
					"Aggregation type must be COUNT when not_reporting is set to true.",
				)
			}
		}
	}
}

func (r *alertResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var tfPlan *alertResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &tfPlan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create the alert from the provided Terraform model...
	input := tfPlan.toAlertDefinitionInput()
	newAlertDef, err := r.client.AlertsService().Create(ctx, input)

	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error creating alert definition '%s'. error: %s", input.Name, err))
		return
	}

	tfPlan.Id = types.StringValue(newAlertDef.Id)

	resp.Diagnostics.Append(resp.State.Set(ctx, &tfPlan)...)
}

func (r *alertResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var tfState *alertResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &tfState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	alertId := tfState.Id.ValueString()
	_, err := r.client.AlertsService().Read(ctx, alertId)

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("error getting alert %s. error: %s",
			alertId,
			err))
		return
	}

	// r.updateState(tfState, alertDef)

	resp.Diagnostics.Append(resp.State.Set(ctx, &tfState)...)
}

func (r *alertResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var tfPlan, tfState *alertResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &tfPlan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &tfState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	alertId := tfState.Id.ValueString()
	input := tfPlan.toAlertDefinitionInput()

	// Update the alert definition...
	_, err := r.client.AlertsService().Update(ctx, alertId, input)

	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error updating alert definition %s. error: %s", alertId, err))
		return
	}

	// Save and log the model into Terraform state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &tfPlan)...)
}

func (r *alertResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var tfState *alertResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &tfState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	alertDefId := tfState.Id.ValueString()

	// Delete the alert definition...
	err := r.client.AlertsService().Delete(ctx, alertDefId)

	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error deleting alert definition %s. error: %s", alertDefId, err))
	}
}

// TODO: Implement ImportState by handling the Read request with latest data from the server.
// func (r *alertResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
// 	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
// }

// TODO: Update the state model with the latest data from the server.
// func (r *alertResource) updateState(state *alertResourceModel, data *swoClient.ReadAlertDefinitionResult) {
// }

func (model *alertResourceModel) toAlertDefinitionInput() swoClient.AlertDefinitionInput {
	conditions := []swoClient.AlertConditionNodeInput{}

	for _, condition := range model.Conditions {
		conditions = condition.toAlertConditionInputs(conditions)
	}

	return swoClient.AlertDefinitionInput{
		Name:                model.Name.ValueString(),
		Description:         model.Description.ValueStringPointer(),
		Enabled:             model.Enabled.ValueBool(),
		Severity:            swoClient.AlertSeverity(model.Severity.ValueString()),
		Actions:             model.toAlertActionInput(),
		TriggerResetActions: model.TriggerResetActions.ValueBoolPointer(),
		Condition:           conditions,
		RunbookLink:         model.RunbookLink.ValueStringPointer(),
	}
}

func (model *alertResourceModel) toAlertActionInput() []swoClient.AlertActionInput {
	inputs := []swoClient.AlertActionInput{}

	//Notifications is deprecated. NotificationActions should be used instead.
	// This if/else maintains backwards compatability.
	if len(model.NotificationActions) > 0 {
		receivingType := swoClient.NotificationReceivingTypeNotSpecified
		includeDetails := true

		for _, action := range model.NotificationActions {
			actionsList := make(map[string][]string)

			for _, configId := range action.ConfigurationIds {
				// Notification Id's are formatted as id:type.
				// This is to accomidate ImportState needing a single Id to import a resource.
				actionId, notificationType, _ := ParseNotificationId(types.StringValue(configId))
				actionType := findCaseInsensitiveMatch(notificationActionTypes, notificationType)

				actionsList[actionType] = append(actionsList[actionType], actionId)
			}

			for actionType, actionIds := range actionsList {
				resendInterval := int(action.ResendIntervalSeconds.ValueInt64())
				inputs = append(inputs, swoClient.AlertActionInput{
					Type:                  actionType,
					ConfigurationIds:      actionIds,
					ResendIntervalSeconds: &resendInterval,
					ReceivingType:         &receivingType,
					IncludeDetails:        &includeDetails,
				})
			}

		}
	} else {
		actionTypes := map[string][]string{}
		for _, configId := range model.Notifications {
			actionId, notificationType, err := ParseNotificationId(types.StringValue(configId))
			actionType := findCaseInsensitiveMatch(notificationActionTypes, notificationType)

			if err == nil {
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
	}

	return inputs
}
