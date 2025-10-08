package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
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
	clients, _ := req.ProviderData.(providerClients)
	r.client = clients.SwoClient
}

func (r *alertResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var tfPlan *alertResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &tfPlan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create the alert from the provided Terraform model...
	input := tfPlan.toAlertDefinitionInput(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

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
	input := tfPlan.toAlertDefinitionInput(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

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

// Transforms a full Alert condition (with logical operators if needed) from
// form values to a flat condition tree format that is accepted by the API.
//
// An example of a resulting condition tree (displayed as a nested tree):
//
//	     AND
//	  /   |   \
//	Con  Con  Con
//
// AND - Binary logical operator
// Con. - Simple condition (comparison operator, metric, threshold, ...)
//
// The function will first create the required number of logical operator nodes with
// pre-computed operator IDs. Then it'll transform all condition items to alert conditions.
func (model *alertResourceModel) toAlertDefinitionInput(ctx context.Context, diags *diag.Diagnostics) swoClient.AlertDefinitionInput {

	var conditions []swoClient.AlertConditionNodeInput

	var planConditions []alertConditionModel
	d := model.Conditions.ElementsAs(ctx, &planConditions, false)
	diags.Append(d...)
	if diags.HasError() {
		return swoClient.AlertDefinitionInput{}
	}
	if len(planConditions) == 1 {
		conditions = planConditions[0].toAlertConditionInputs(ctx, diags, 0)
		if diags.HasError() {
			return swoClient.AlertDefinitionInput{}
		}
	} else {

		// Currently we only allow one AND logical operator type on the top level
		// These logical operators can be n-ary, so we just need one logicalOperator
		rootLogicalOperator := swoClient.AlertConditionNodeInput{}
		rootLogicalOperator.Id = 0
		rootLogicalOperator.Type = string(swoClient.AlertLogicalOperatorType)
		logicalOperator := string(swoClient.AlertOperatorAnd)
		rootLogicalOperator.Operator = &logicalOperator

		// Pre-computed child operator IDs
		numConditions := len(planConditions)
		rootOperandIds := make([]int, numConditions)
		// We need to create 5 flat tree nodes to build an alert condition:
		// - binaryOperator (comparisonOperator)
		// - aggregationOperator
		// - constantValue (threshold)
		// - metricField
		// - constantValue (time frame)
		// and condition limit is 5 (for nested conditions). Total 30 = parent condition(10 = 2:relationship operator & scope-field + remaining for logical operators) + nested conditions max(5)(5*5 = 25)
		alertConditionCountBuffer := 65
		for i := 0; i < numConditions; i++ {
			rootOperandIds[i] = (alertConditionCountBuffer * i) + 1
		}
		rootLogicalOperator.OperandIds = rootOperandIds
		conditions = []swoClient.AlertConditionNodeInput{rootLogicalOperator}

		// Create an alert condition for each
		for i := 0; i < numConditions; i++ {
			childRootNodeId := rootOperandIds[i]
			childConditions := planConditions[i].toAlertConditionInputs(ctx, diags, childRootNodeId)
			if diags.HasError() {
				return swoClient.AlertDefinitionInput{}
			}
			conditions = append(conditions, childConditions...)
		}
	}

	triggerDelay := int(model.TriggerDelaySeconds.ValueInt64())
	var noDataResetSeconds *int
	if p := model.NoDataResetSeconds.ValueInt64Pointer(); p != nil {
		value := int(*p)
		noDataResetSeconds = &value
	}

	// The API forces a match between NoDataResetSeconds and TimeRangeSeconds in EntityFilter.
	for _, condition := range conditions {
		if condition.Type != string(swoClient.AlertAttributeType) || condition.EntityFilter == nil {
			continue
		}
		condition.EntityFilter.TimeRangeSeconds = noDataResetSeconds
	}

	actions := model.toAlertActionInput(ctx, diags)
	if diags.HasError() {
		return swoClient.AlertDefinitionInput{}
	}
	return swoClient.AlertDefinitionInput{
		Name:                model.Name.ValueString(),
		Description:         model.Description.ValueStringPointer(),
		Enabled:             model.Enabled.ValueBool(),
		Severity:            swoClient.AlertSeverity(model.Severity.ValueString()),
		Actions:             actions,
		TriggerResetActions: model.TriggerResetActions.ValueBoolPointer(),
		Condition:           conditions,
		RunbookLink:         model.RunbookLink.ValueStringPointer(),
		TriggerDelaySeconds: &triggerDelay,
		NoDataResetSeconds:  noDataResetSeconds,
	}
}

func (model *alertResourceModel) toAlertActionInput(ctx context.Context, diags *diag.Diagnostics) []swoClient.AlertActionInput {
	var inputs []swoClient.AlertActionInput

	var notificationActions []alertActionInputModel
	d := model.NotificationActions.ElementsAs(ctx, &notificationActions, false)
	diags.Append(d...)
	if diags.HasError() {
		return inputs
	}

	// Notifications is deprecated. NotificationActions should be used instead.
	// This if/else maintains backwards compatability.
	if len(notificationActions) > 0 {
		receivingType := swoClient.NotificationReceivingTypeNotSpecified
		includeDetails := true

		for _, action := range notificationActions {
			actionsList := make(map[string][]string)

			var configurationIds []string
			dIds := action.ConfigurationIds.ElementsAs(ctx, &configurationIds, false)
			diags.Append(dIds...)
			if diags.HasError() {
				return inputs
			}

			for _, configId := range configurationIds {
				// Notification Ids are formatted as id:type.
				// This is to accommodate ImportState needing a single Id to import a resource.

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
		for _, configId := range model.Notifications.Elements() {
			actionId, notificationType, err := ParseNotificationId(types.StringValue(configId.String()))
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
