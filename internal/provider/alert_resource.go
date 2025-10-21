package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	swoClient "github.com/solarwinds/swo-client-go/pkg/client"
	"github.com/solarwinds/terraform-provider-swo/internal/alerts"
	"github.com/solarwinds/terraform-provider-swo/internal/typex"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &alertResource{}
	_ resource.ResourceWithConfigure   = &alertResource{}
	_ resource.ResourceWithImportState = &alertResource{}
)

func NewAlertResource() resource.Resource {
	return &alertResource{}
}

type alertResource struct {
	client *swoClient.Client
}

func (r *alertResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "alert"
}

func (r *alertResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
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
	tfPlan.ForceUpdate = types.BoolValue(false)
	resp.Diagnostics.Append(resp.State.Set(ctx, &tfPlan)...)
}

func (r *alertResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var tfState *alertResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &tfState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	alertId := tfState.Id.ValueString()
	alertDef, err := r.client.AlertsService().Read(ctx, alertId)

	if errors.Is(err, swoClient.ErrNotFound) {
		resp.State.RemoveResource(ctx)
		return
	} else if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error getting alert %s. error: %s", alertId, err))
		return
	}

	r.updateState(ctx, tfState, alertDef, &resp.Diagnostics)
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

	if errors.Is(err, swoClient.ErrNotFound) {
		resp.State.RemoveResource(ctx)
		return
	} else if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error updating alert definition %s. error: %s", alertId, err))
		return
	}

	// Save and log the model into Terraform state.
	tfPlan.ForceUpdate = types.BoolValue(false)
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

	if err != nil && !errors.Is(err, swoClient.ErrNotFound) {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error deleting alert definition %s. error: %s", alertDefId, err))
	}
}

func (r *alertResource) ImportState(ctx context.Context,
	req resource.ImportStateRequest, resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (*alertResource) updateState(ctx context.Context,
	state *alertResourceModel, result *swoClient.ReadAlertDefinitionResult, diags *diag.Diagnostics,
) {
	actionsInResponse := actionDescriptionsFromResult(result.Actions)
	actionsInState := actionDescriptionsFromInput(state.notificationsToInput(ctx, diags))
	if diags.HasError() {
		return
	}

	notificationActions := state.NotificationActions
	deprecatedNotifications := state.Notifications
	if !actionsInResponse.equals(actionsInState) {
		// Drift detected. We use the response for the new state. We decide which attribute
		// to update based on whether notificationActions is known (was set).
		if !notificationActions.IsNull() {
			notificationActions = actionsInResponse.toModelActions()
		} else {
			var d diag.Diagnostics
			deprecatedNotifications, d = deprecatedActionsToNotifications(actionsInResponse.toModelActions())

			if diags.HasError() {
				diags.Append(d...)
				return
			}
		}
	}

	conditionsInResponse := conditionsFromResult(result.FlatCondition, diags)
	conditionsInState := conditionsFromInput(state.toAlertDefinitionInput(ctx, diags).Condition, diags)
	if diags.HasError() {
		return
	}

	conditionsSet := state.Conditions
	if !conditionsInResponse.Equals(conditionsInState) {
		// Drift detected. We use the response for the new state.
		conditionsSet = toModelConditions(conditionsInResponse, state.Conditions, diags)
		if diags.HasError() {
			return
		}
	}

	*state = alertResourceModel{
		Id:                  types.StringValue(result.Id),
		Name:                types.StringValue(result.Name),
		NotificationActions: notificationActions,
		Description:         types.StringPointerValue(result.Description),
		Severity:            types.StringValue(string(result.Severity)),
		Enabled:             types.BoolValue(result.Enabled),
		Conditions:          conditionsSet,
		Notifications:       deprecatedNotifications,
		TriggerResetActions: types.BoolValue(result.TriggerResetActions),
		RunbookLink:         types.StringPointerValue(result.RunbookLink),
		TriggerDelaySeconds: types.Int64Value(int64(result.TriggerDelaySeconds)),
		NoDataResetSeconds:  types.Int64PointerValue(typex.CastIntPtr[int, int64](result.NoDataResetSeconds)),
		ForceUpdate:         types.BoolValue(conditionsSet.IsNull()),
	}
}

func conditionsFromResult(result []swoClient.ReadAlertConditionResult, diags *diag.Diagnostics) alerts.Condition {
	condition, err := alerts.ConditionsFromResult(result)
	if err != nil {
		diags.AddError("Unexpected Conditions Format",
			fmt.Sprintf("error parsing conditions from the alert definition response: %s", err))
		return nil
	}

	return condition
}

func conditionsFromInput(result []swoClient.AlertConditionNodeInput, diags *diag.Diagnostics) alerts.Condition {
	condition, err := alerts.ConditionsFromInput(result)
	if err != nil {
		diags.AddError("Unexpected Conditions Format",
			fmt.Sprintf("error parsing conditions from the alert state: %s", err))
		return nil
	}

	return condition
}

// toModelConditions translates the given alert condition tree into the Terraform model. The
// condition is expected to have the form produced by this provider on create/update. It has to
// be either a single, top-level attribute or metric condition, or a conjunction (AND) of such
// conditions. If this pattern is violated, then the condition cannot be stored in the Terraform
// model and a null set is returned. This indicates that the resource should be replaced.
func toModelConditions(response alerts.Condition, state types.Set, diags *diag.Diagnostics) types.Set {
	// Given the current model and our construction of conditions, we map individual terms
	// in a top-level AND to individual entries in the model's condition set. If the top level
	// is not AND, then a single condition will result in the set.
	sourceConditions := []alerts.Condition{response}
	if isAndOperator(response) {
		sourceConditions = response.GetOperands()
	}

	warnUnsupported := func() {
		diags.AddWarning("Unsupported Condition",
			"the condition as set is not supported by this provider")
	}
	setObjectType := types.ObjectType{AttrTypes: AlertConditionAttributeTypes()}

	stateConditions := state.Elements()
	var modelConditions []attr.Value
	for idx, cond := range sourceConditions {
		// We build only two types of conditions: either attribute or entity. We can tell them
		// apart based on the type for the first operand for cond. Once detected, that's what
		// we try to parse. If we can't get that operand, then it's not a condition we could
		// have built from Terraform. In that case we simply abort because the condition must
		// be fully reset.
		if !isBinaryOperator(cond) || len(cond.GetOperands()) != 2 {
			warnUnsupported()
			return types.SetNull(setObjectType)
		}
		firstOperand := cond.GetOperands()[0]
		modelCondition := types.ObjectNull(AlertConditionAttributeTypes())

		switch swoClient.AlertOperatorType(firstOperand.GetType()) {
		case swoClient.AlertAttributeType:
			modelCondition = attributeConditionToModel(cond, diags)
		case swoClient.AlertAggregationOperatorType:
			modelCondition = metricConditionToModel(cond, diags)
		}

		if diags.HasError() {
			// Error parsing the condition. We continue to allow for other errors to show up.
			// Processing will still be terminated in that case.
			continue
		}
		if modelCondition.IsNull() {
			// We don't recognize this condition or cannot map it to the Terraform model, so
			// there's no point in pushing this further. We'll just have to reset it.
			warnUnsupported()
			return types.SetNull(setObjectType)
		}

		// We try to use values in the current state when they are equivalent, to avoid
		// unnecessary and meaningless differences. Note that we don't try to match the actual
		// conditions with the state, but just refer to them by position. This could end up
		// looking at different terms, but this is only an effort at trying to reduce noise.
		// If the conditions are otherwise different, the drift will be caught anyway.
		condInState := types.ObjectNull(AlertConditionAttributeTypes())
		if idx < len(stateConditions) {
			if c, isObject := stateConditions[idx].(types.Object); isObject {
				condInState = c
			}
		}
		modelConditions = append(modelConditions, typex.ObjectAttributesCoalesce(condInState, modelCondition))
	}

	set, d := types.SetValue(setObjectType, modelConditions)
	diags.Append(d...)
	return set
}

// attributeConditionToModel translates an attribute-based condition to the Terraform model. The
// condition must follow the structure built by this provider on the create or update requests,
// for it to be successfully translated. A null object will be returned if it does not, to
// indicate that the resource should be replaced instead. This function assumes that condition
// is a binary operator with two operands (it will try to access both) and that the first of them
// is an attribute type.
func attributeConditionToModel(condition alerts.Condition, diags *diag.Diagnostics) types.Object {
	// We know already that there are two operands and that the first is an attribute. We
	// need to check that the second is a constant and none of them have operands themselves.
	attribute := condition.GetOperands()[0]
	constant := condition.GetOperands()[1]
	if len(attribute.GetOperands()) != 0 || !isConstantOperator(constant) || len(constant.GetOperands()) != 0 {
		return types.ObjectNull(AlertConditionAttributeTypes())
	}

	attributeName := types.StringPointerValue(attribute.GetFieldName())
	attributeOperator := types.StringPointerValue(condition.GetOperator())
	inOperator := types.StringValue(string(swoClient.AlertOperatorIn))

	entityIds, entityTypes, entityQuery, isValid := unpackEntityFilter(attribute.GetEntityFilter(), diags)
	if diags.HasError() || !isValid {
		return types.ObjectNull(AlertConditionAttributeTypes())
	}

	attributeValue := types.StringNull()
	attributeValues := types.ListNull(types.StringType)
	var computedDataType *string
	if attributeOperator.Equal(inOperator) {
		// Builds the values list; attributeValue remains unset.
		attributeValues = typex.StringSliceToList(constant.GetValues(), diags)
		if diags.HasError() {
			return types.ObjectNull(AlertConditionAttributeTypes())
		}

		// Computes the data type for the received list of values, in the same way that it
		// is computed when creating or updating the resource.
		if vs := constant.GetValues(); len(vs) > 0 {
			dataType := GetStringDataType(vs[0])
			computedDataType = &dataType
		}
	} else {
		// Sets a single value; the attributeValues list remains unset.
		attributeValue = types.StringPointerValue(constant.GetValue())
		if constant.GetValue() != nil {
			dataType := GetStringDataType(*constant.GetValue())
			computedDataType = &dataType
		}
	}

	if !typex.PtrEqual(constant.GetDataType(), computedDataType) {
		// The only reliable approach here is replacing the resource.
		diags.AddWarning("Unexpected Data Type",
			fmt.Sprintf("unexpected attribute value(s) data type was found: %v", constant.GetDataType()))
		return types.ObjectNull(AlertConditionAttributeTypes())
	}

	result, d := types.ObjectValue(AlertConditionAttributeTypes(),
		map[string]attr.Value{
			"aggregation_type":    types.StringNull(),
			"attribute_name":      attributeName,
			"attribute_operator":  attributeOperator,
			"attribute_value":     attributeValue,
			"attribute_values":    attributeValues,
			"duration":            types.StringNull(),
			"entity_ids":          entityIds,
			"exclude_tags":        types.SetNull(types.ObjectType{AttrTypes: AlertTagAttributeTypes()}),
			"group_by_metric_tag": types.ListNull(types.StringType),
			"include_tags":        types.SetNull(types.ObjectType{AttrTypes: AlertTagAttributeTypes()}),
			"metric_name":         types.StringNull(),
			"not_reporting":       types.BoolValue(false),
			"query_search":        entityQuery,
			"target_entity_types": entityTypes,
			"threshold":           types.StringNull(),
		})
	diags.Append(d...)
	return result
}

// unpackEntityFilter goes through the given entity filter and returns (in this order) the entity
// ids, entity types and query. If the filter is nil, then all of these are set to null. An
// additional boolean return value indicates whether the filter is valid. If it is not, then there
// is no way to map it back to the Terraform model, so the resource must be replaced.
func unpackEntityFilter(
	filter alerts.EntityFilter, diags *diag.Diagnostics,
) (types.List, types.List, types.String, bool) {
	entityIds := types.ListNull(types.StringType)
	entityTypes := types.ListNull(types.StringType)
	entityQuery := types.StringNull()

	if filter != nil {
		if len(filter.GetFields()) > 0 {
			return entityIds, entityTypes, entityQuery, false
		}
		entityIds = typex.StringSliceToList(filter.GetIds(), diags)
		entityTypes = typex.StringSliceToList(filter.GetTypes(), diags)
		entityQuery = types.StringPointerValue(filter.GetQuery())
	}

	return entityIds, entityTypes, entityQuery, true
}

// metricConditionToModel translates a metric-based condition to the Terraform model. The
// condition must follow the structure built by this provider on the create or update requests,
// for it to be successfully translated. A null object will be returned if it does not, to
// indicate that the resource should be replaced instead. This function assumes that condition
// is a binary operator with two operands (it will try to access both) and that the first of them
// is an aggregation type.
func metricConditionToModel(condition alerts.Condition, diags *diag.Diagnostics) types.Object {
	// Read the first-level condition nodes: aggregation and threshold.
	aggregation := condition.GetOperands()[0]
	threshold := condition.GetOperands()[1]
	if len(aggregation.GetOperands()) != 2 || !isConstantOperator(threshold) || len(threshold.GetOperands()) != 0 {
		return types.ObjectNull(AlertConditionAttributeTypes())
	}

	// Read the second-level nodes, under aggregation: metric and duration. We do this early to
	// fail fast in case that the condition doesn't have a form that we can handle in Terraform.
	metric := aggregation.GetOperands()[0]
	duration := aggregation.GetOperands()[1]
	if !isAggregationOperator(aggregation) || len(metric.GetOperands()) != 0 ||
		!isConstantOperator(duration) || len(duration.GetOperands()) != 0 {
		return types.ObjectNull(AlertConditionAttributeTypes())
	}

	aggregationType := types.StringPointerValue(aggregation.GetOperator())
	thresholdValue := types.StringNull()
	notReporting := types.BoolValue(false)
	var expectedThresholdDataType *string
	if op, v := condition.GetOperator(), threshold.GetValue(); op != nil && v != nil {
		// In case we find a zero (without units), we set not_reporting to true. The forward
		// translation is not entirely reversible, because the user could have set "=0" as the
		// threshold and false for not_reporting, and we'd end up exactly in the same place.
		// We assume this case to be set by a true not_reporting instead. Ideally, this should
		// be handled differently, so that these semantically equivalent specs produce no plan
		// differences. But this adds more complexity and I have to draw the line somewhere.
		valueWithUnits := *op + *v
		if valueWithUnits == "0" {
			notReporting = types.BoolValue(true)
		}
		thresholdValue = types.StringValue(*op + *v)
		dataType := GetStringDataType(*v)
		expectedThresholdDataType = &dataType
	}
	if !typex.PtrEqual(threshold.GetDataType(), expectedThresholdDataType) {
		// The only reliable approach here is replacing the resource.
		diags.AddWarning("Unexpected Data Type",
			fmt.Sprintf("unexpected threshold data type was found: %v", threshold.GetDataType()))
		return types.ObjectNull(AlertConditionAttributeTypes())
	}

	metricName := types.StringPointerValue(metric.GetFieldName())
	entityIds, entityTypes, entityQuery, isEValid := unpackEntityFilter(metric.GetEntityFilter(), diags)
	includeTags, excludeTags, isMValid := unpackMetricFilter(metric.GetMetricFilter(), diags)
	if diags.HasError() || !isEValid || !isMValid {
		return types.ObjectNull(AlertConditionAttributeTypes())
	}

	groupByMetricTag := typex.StringSliceToList(metric.GetGroupByMetricTag(), diags)
	if diags.HasError() {
		return types.ObjectNull(AlertConditionAttributeTypes())
	}

	durationValue := types.StringNull()
	var expectedDurationDataType *string
	if d := duration.GetValue(); d != nil {
		durationValue = types.StringValue(*d)
		dataType := GetStringDataType(*d)
		expectedDurationDataType = &dataType
	}
	if !typex.PtrEqual(duration.GetDataType(), expectedDurationDataType) {
		// The only reliable approach here is replacing the resource.
		diags.AddWarning("Unexpected Data Type",
			fmt.Sprintf("unexpected duration data type was found: %v", duration.GetDataType()))
		return types.ObjectNull(AlertConditionAttributeTypes())
	}

	result, d := types.ObjectValue(AlertConditionAttributeTypes(),
		map[string]attr.Value{
			"aggregation_type":    aggregationType,
			"attribute_name":      types.StringNull(),
			"attribute_operator":  types.StringNull(),
			"attribute_value":     types.StringNull(),
			"attribute_values":    types.ListNull(types.StringType),
			"duration":            durationValue,
			"entity_ids":          entityIds,
			"exclude_tags":        excludeTags,
			"group_by_metric_tag": groupByMetricTag,
			"include_tags":        includeTags,
			"metric_name":         metricName,
			"not_reporting":       notReporting,
			"query_search":        entityQuery,
			"target_entity_types": entityTypes,
			"threshold":           thresholdValue,
		})
	diags.Append(d...)
	return result
}

// unpackMetricFilter goes through the given metric filter and extracts the include and exclude tags.
// The filter must be in the form of a single tag node (possibly negated) or a conjunction (AND) of
// multiple such nodes. The function returns the set of include and exclude tags (in this order) and
// a boolean value, indicating whether the filter is valid. When not valid, there is no way to map
// the filter back to the Terraform model, so the resource must be replaced.
func unpackMetricFilter(filter alerts.MetricFilter, diags *diag.Diagnostics) (types.Set, types.Set, bool) {
	includeTagsSet := types.SetNull(types.ObjectType{AttrTypes: AlertTagAttributeTypes()})
	excludeTagsSet := types.SetNull(types.ObjectType{AttrTypes: AlertTagAttributeTypes()})

	if filter == nil {
		return includeTagsSet, excludeTagsSet, true
	}

	// We allow either a single filter node, or an AND operation with multiple tag nodes.
	var tagNodes []alerts.MetricFilter
	if filter.GetOperation() == swoClient.FilterOperationAnd {
		tagNodes = filter.GetOperands()
	} else {
		tagNodes = []alerts.MetricFilter{filter}
	}

	includeTags := make([]attr.Value, 0)
	excludeTags := make([]attr.Value, 0)

	for _, item := range tagNodes {
		tagNode := item
		negated := false
		if tagNode.GetOperation() == swoClient.FilterOperationNot {
			if len(tagNode.GetOperands()) != 1 {
				diags.AddWarning("Missing Operands", "metric filter NOT must have exactly one argument")
				return includeTagsSet, excludeTagsSet, false
			}
			tagNode = tagNode.GetOperands()[0]
			negated = true
		}

		if len(tagNode.GetOperands()) > 0 {
			// Nested constructions are not supported here; we need to replace the alert.
			return includeTagsSet, excludeTagsSet, false
		}
		if tagNode.GetPropertyName() == nil {
			// This should not happen; the only fix at this point is replacing the resource.
			diags.AddWarning("Unexpected Metric Filter", "metric property name is missing")
			return includeTagsSet, excludeTagsSet, false
		}
		propertyName := types.StringValue(*tagNode.GetPropertyName())

		// When we build the API request in create/update, we always use the IN operator and the
		// list of values, regardless of how many values there are. Here we try to be a bit more
		// flexible and accept the EQ operator with a single value as well, that is translated
		// to a single value list.
		var propertyValues types.List
		switch tagNode.GetOperation() {
		case swoClient.FilterOperationEq:
			propertyValues = typex.StringPtrSliceToList([]*string{tagNode.GetPropertyValue()}, diags)
		case swoClient.FilterOperationIn:
			propertyValues = typex.StringPtrSliceToList(tagNode.GetPropertyValues(), diags)
		default:
			diags.AddWarning("Unexpected Metric Filter",
				fmt.Sprintf("unexpected filter operation: %v", tagNode.GetOperation()))
			return includeTagsSet, excludeTagsSet, false
		}

		tagsObj := types.ObjectValueMust(AlertTagAttributeTypes(), map[string]attr.Value{
			"name":   propertyName,
			"values": propertyValues,
		})
		if negated {
			excludeTags = append(excludeTags, tagsObj)
		} else {
			includeTags = append(includeTags, tagsObj)
		}
	}

	var d diag.Diagnostics
	includeTagsSet, d = types.SetValue(types.ObjectType{AttrTypes: AlertTagAttributeTypes()}, includeTags)
	diags.Append(d...)
	excludeTagsSet, d = types.SetValue(types.ObjectType{AttrTypes: AlertTagAttributeTypes()}, excludeTags)
	diags.Append(d...)
	return includeTagsSet, excludeTagsSet, true
}

func isAndOperator(condition alerts.Condition) bool {
	return swoClient.AlertOperatorType(condition.GetType()) == swoClient.AlertLogicalOperatorType &&
		condition.GetOperator() != nil &&
		*condition.GetOperator() == string(swoClient.AlertOperatorAnd)
}

func isBinaryOperator(condition alerts.Condition) bool {
	return swoClient.AlertOperatorType(condition.GetType()) == swoClient.AlertBinaryOperatorType
}

func isAggregationOperator(condition alerts.Condition) bool {
	return swoClient.AlertOperatorType(condition.GetType()) == swoClient.AlertAggregationOperatorType
}

func isConstantOperator(condition alerts.Condition) bool {
	return swoClient.AlertOperatorType(condition.GetType()) == swoClient.AlertConstantValueType
}

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
	noDataResetSeconds := typex.CastIntPtr[int64, int](model.NoDataResetSeconds.ValueInt64Pointer())

	// The API forces a match between NoDataResetSeconds and TimeRangeSeconds in EntityFilter.
	for _, condition := range conditions {
		if condition.Type != string(swoClient.AlertAttributeType) || condition.EntityFilter == nil {
			continue
		}
		condition.EntityFilter.TimeRangeSeconds = noDataResetSeconds
	}

	actions := model.notificationsToInput(ctx, diags)
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

// normalizedNotifications returns notifications in the model, either from NotificationActions,
// when defined, or from the deprecated Notifications field, if the former wasn't provided.
func (model *alertResourceModel) normalizedNotifications() (types.Set, diag.Diagnostics) {
	if !model.NotificationActions.IsNull() {
		return model.NotificationActions, nil
	}

	// We need to migrate from the old field.
	return deprecatedNotificationsToActions(model.Notifications)
}

// notificationsToInput translates the notification actions element in the model to the
// appropriate type to use with the SWO client. This includes processing deprecated fields
// when appropriate.
func (model *alertResourceModel) notificationsToInput(ctx context.Context, diags *diag.Diagnostics) []swoClient.AlertActionInput {
	// Convert from deprecated fields when needed. This can be removed once no more deprecated
	// fields remain.
	notifications, d := model.normalizedNotifications()
	diags.Append(d...)
	if diags.HasError() {
		return []swoClient.AlertActionInput{}
	}

	// Compute the SWO client input type.
	inputs, d := modelActionsToInput(ctx, notifications)
	diags.Append(d...)
	if diags.HasError() {
		return []swoClient.AlertActionInput{}
	}

	return inputs
}
