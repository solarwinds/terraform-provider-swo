package provider

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	swoClient "github.com/solarwinds/swo-client-go/pkg/client"
	"github.com/solarwinds/terraform-provider-swo/internal/typex"
)

const (
	// This is used when the deprecated notifications attribute is in use.
	defaultResendIntervalInSecs = 600
)

// alertActionDescription abstracts the getters for the properties of notifications that
// we are interested in: those that are actually exposed in Terraform. The provider will
// ignore the others, even if they changed in the API, to avoid meaningless drift.
type alertActionDescription interface {
	GetType() string
	GetConfigurationIds() []string
	GetResendIntervalSeconds() *int
}

var (
	// Ensure that both input and result types implement this interface.
	_ alertActionDescription = &swoClient.AlertActionInput{}
	_ alertActionDescription = &swoClient.ReadAlertActionResult{}
)

type alertActionDescriptions []alertActionDescription

// actionDescriptionsFromInput translates an array of action inputs to an alert descriptions
// type, normalizing the order of inputs and configuration IDs in each input. Note that input
// is assumed to have a single item for each combination of type and the other parameters
// other than the configuration IDs. If this didn't hold, the result would not be normalized.
func actionDescriptionsFromInput(input []swoClient.AlertActionInput) alertActionDescriptions {
	actions := typex.Map(input, func(a swoClient.AlertActionInput) alertActionDescription {
		a.ConfigurationIds = typex.SliceShallowClone(a.ConfigurationIds)
		sort.Strings(a.ConfigurationIds)
		return &a
	})
	sort.Slice(actions, func(i, j int) bool {
		return actions[i].GetType() < actions[j].GetType() ||
			actions[i].GetType() == actions[j].GetType() &&
				typex.PtrCompare(actions[i].GetResendIntervalSeconds(), actions[j].GetResendIntervalSeconds(), typex.Less)
	})
	return actions
}

// actionDescriptionsFromResult builds an alert descriptions type from API read results.
// The payload is normalized such that there's a single action description for each
// combination of notification type and the rest of parameters other than configuration
// IDs, configuration IDs are sorted and there's a global sort as well.
func actionDescriptionsFromResult(result []swoClient.ReadAlertActionResult) alertActionDescriptions {
	// We need to normalize, just like we do when converting from the Terraform model to the
	// input types. All action IDs (ConfigurationIds) sharing the same action parameters go
	// in a single list under one action result.
	type actionParameters struct {
		actionType            string
		resendIntervalSeconds int
	}
	resultIndexByParams := make(map[actionParameters][]int)
	needsCollapsing := false

	for idx, action := range result {
		// We may change this array later (through sorting and maybe collapsing). Cloning
		// avoids the side effect on the original payload sent from the API.
		action.ConfigurationIds = typex.SliceShallowClone(action.ConfigurationIds)

		params := actionParameters{
			actionType:            action.Type,
			resendIntervalSeconds: typex.DerefOrDefault(action.ResendIntervalSeconds, 0),
		}
		indices := resultIndexByParams[params]
		if len(indices) > 0 {
			needsCollapsing = true
		}
		resultIndexByParams[params] = append(indices, idx)
	}

	normalizedResult := result
	if needsCollapsing {
		// At least one set of parameters occurs more than once in result. We go over the map
		// and collapse subsequent occurrences of the same parameters in result into the first
		// that we found. Note that the map doesn't preserver order, but we don't care. The
		// slice will be sorted into normal form below, anyway.
		normalizedResult = make([]swoClient.ReadAlertActionResult, 0, len(result))
		for _, indices := range resultIndexByParams {
			// One element is guaranteed by construction.
			base := result[indices[0]]

			// Just for good hygiene. Not strictly needed, as we don't use them.
			base.IncludeDetails = nil
			base.ReceivingType = nil

			for _, extraIdx := range indices[1:] {
				base.ConfigurationIds = append(base.ConfigurationIds, result[extraIdx].ConfigurationIds...)
			}
			normalizedResult = append(normalizedResult, base)
		}
	}

	// Finally we sort all configuration ID slices and sort result by action.
	actions := typex.Map(normalizedResult, func(a swoClient.ReadAlertActionResult) alertActionDescription {
		sort.Strings(a.ConfigurationIds)
		return &a
	})
	sort.Slice(actions, func(i, j int) bool {
		return actions[i].GetType() < actions[j].GetType() ||
			actions[i].GetType() == actions[j].GetType() &&
				typex.PtrCompare(actions[i].GetResendIntervalSeconds(), actions[j].GetResendIntervalSeconds(), typex.Less)
	})
	return actions
}

func (actions alertActionDescriptions) toModelActions() types.Set {
	result := make([]attr.Value, len(actions))

	for i, action := range actions {
		configurationIds := make([]attr.Value, len(action.GetConfigurationIds()))
		for j, id := range action.GetConfigurationIds() {
			// We lowercase the type to match the validation in the schema.
			configurationIds[j] = types.StringValue(fmt.Sprintf("%s:%s", id, strings.ToLower(action.GetType())))
		}
		configurationIdsList := types.ListValueMust(types.StringType, configurationIds)

		resendIntervalSeconds := types.Int64Null()
		if action.GetResendIntervalSeconds() != nil {
			resendIntervalSeconds = types.Int64Value(int64(*action.GetResendIntervalSeconds()))
		}

		result[i] = types.ObjectValueMust(alertActionAttributeTypes(), map[string]attr.Value{
			"configuration_ids":       configurationIdsList,
			"resend_interval_seconds": resendIntervalSeconds,
		})
	}

	// The use of the ...Must() variants here and above helps avoid needless reflection
	// and error checking. We are building the values right here. They are correctly typed
	// by construction.
	return types.SetValueMust(types.ObjectType{AttrTypes: alertActionAttributeTypes()}, result)
}

// equals indicates whether descriptions given as provider and argument are equal. Equality
// here is defined as: the descriptions have the same elements in the same order, with
// individual elements compared by all methods defined in the alertActionsDescription type.
func (actions alertActionDescriptions) equals(others alertActionDescriptions) bool {
	return typex.SliceEqualFunc(actions, others, func(a, b alertActionDescription) bool {
		return a.GetType() == b.GetType() &&
			typex.SliceEqual(a.GetConfigurationIds(), b.GetConfigurationIds()) &&
			typex.PtrEqual(a.GetResendIntervalSeconds(), b.GetResendIntervalSeconds())
	})
}

// deprecatedNotificationsToActions converts the given deprecated 'notifications' attribute in
// the alert resource schema to the set of objects required by the new 'notification_actions'
// that replaced it. The argument to this method is expected to be a list of types.StringType.
func deprecatedNotificationsToActions(notifications types.List) (types.Set, diag.Diagnostics) {
	var diags diag.Diagnostics

	items := notifications.Elements()
	actions := make([]attr.Value, len(items))
	actionObjectType := types.ObjectType{AttrTypes: alertActionAttributeTypes()}

	for i, configId := range items {
		configIds, d := types.ListValue(types.StringType, []attr.Value{configId})
		if diags = append(diags, d...); diags.HasError() {
			continue
		}

		actions[i], d = types.ObjectValue(alertActionAttributeTypes(), map[string]attr.Value{
			"configuration_ids":       configIds,
			"resend_interval_seconds": types.Int64Value(defaultResendIntervalInSecs),
		})
		diags = append(diags, d...)
	}

	if diags.HasError() {
		return types.SetUnknown(actionObjectType), diags
	}

	// The use of the ...Must() variant here helps avoid needless reflection and error
	// checking. We are building the values right here. They are correctly typed by
	// construction.
	return types.SetValueMust(actionObjectType, actions), diags
}

// deprecatedActionsToNotifications converts a set of notification actions, as given by the
// Terraform model for the notification_actions attribute, into the format expected by the
// deprecated notifications attribute.
func deprecatedActionsToNotifications(actions types.Set) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	var result []attr.Value

	for idx, action := range actions.Elements() {
		obj, ok := action.(types.Object)
		if !ok {
			diags.AddError("Invalid Set Element Type",
				fmt.Sprintf("Action #%d is not an object, but %T", idx, action))
			continue
		}

		tfConfigIds, found := obj.Attributes()["configuration_ids"]
		if !found {
			diags.AddError("Missing Attribute In Object",
				fmt.Sprintf("Action #%d is missing the configuration_ids attribute", idx))
			continue
		}

		configIds, ok := tfConfigIds.(types.List)
		if !ok {
			diags.AddError("Invalid Configuration Attribute",
				fmt.Sprintf("Action #%d's configuration_ids is not a list, but %T", idx, tfConfigIds))
			continue
		}

		result = append(result, configIds.Elements()...)
	}

	if diags.HasError() {
		return types.ListNull(types.StringType), diags
	}
	return types.ListValue(types.StringType, result)
}

// modelActionsToInput converts the notification actions, as defined in our Terraform schema
// to the input structure for our SWO client. The result is arranged so that there's only one
// element in the slice for each combination of action type and the rest of the parameters
// other than the SWO API's configuration IDs. Each such element includes all applicable
// configuration IDs. This is not yet a normalized form, since order is not defined.
func modelActionsToInput(ctx context.Context, actions types.Set) ([]swoClient.AlertActionInput, diag.Diagnostics) {
	// Used to group notification actions by their parameters for normalization.
	type actionParameters struct {
		actionType            string
		resendIntervalSeconds int
	}
	actionIdsByParams := make(map[actionParameters][]string)

	var notificationActions []alertActionInputModel
	diags := actions.ElementsAs(ctx, &notificationActions, false)
	if diags.HasError() {
		return []swoClient.AlertActionInput{}, diags
	}

	for _, action := range notificationActions {
		resendInterval := int(action.ResendIntervalSeconds.ValueInt64())

		var configIds []types.String
		d := action.ConfigurationIds.ElementsAs(ctx, &configIds, false)
		diags.Append(d...)
		if diags.HasError() {
			continue
		}

		for _, configId := range configIds {
			actionId, rawActionType, err := ParseNotificationId(configId)
			if err != nil {
				diags.AddError("Invalid Configuration ID", err.Error())
				continue
			}

			actionType, d := canonicalNotificationActionType(rawActionType)
			diags.Append(d...)
			if diags.HasError() {
				continue
			}

			// Add the actionId to the list under the same parameters.
			params := actionParameters{
				actionType:            actionType,
				resendIntervalSeconds: resendInterval,
			}
			actionIdsByParams[params] = append(actionIdsByParams[params], actionId)
		}
	}

	if diags.HasError() {
		return []swoClient.AlertActionInput{}, diags
	}

	// Fixed notification parameters, not currently exposed via Terraform.
	receivingType := swoClient.NotificationReceivingTypeNotSpecified
	includeDetails := true

	var inputs []swoClient.AlertActionInput
	for actionParams, actionIds := range actionIdsByParams {
		resendIntervalSeconds := actionParams.resendIntervalSeconds
		inputs = append(inputs, swoClient.AlertActionInput{
			Type:                  actionParams.actionType,
			ConfigurationIds:      actionIds,
			ResendIntervalSeconds: &resendIntervalSeconds,
			ReceivingType:         &receivingType,
			IncludeDetails:        &includeDetails,
		})
	}

	return inputs, nil
}
