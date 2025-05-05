package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"reflect"

	"github.com/cenkalti/backoff/v5"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	swoClient "github.com/solarwinds/swo-client-go/pkg/client"
	swoClientTypes "github.com/solarwinds/swo-client-go/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &uriResource{}
	_ resource.ResourceWithConfigure   = &uriResource{}
	_ resource.ResourceWithImportState = &uriResource{}
)

func NewUriResource() resource.Resource {
	return &uriResource{}
}

// Defines the resource implementation.
type uriResource struct {
	client *swoClient.Client
}

func (r *uriResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "uri"
}

func (r *uriResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, _ := req.ProviderData.(providerClients)
	r.client = client.SwoClient
}

func (r *uriResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var tfPlan uriResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &tfPlan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var planOptions uriResourceOptions
	tfPlan.Options.As(ctx, &planOptions, basetypes.ObjectAsOptions{})
	var tcpOptions uriResourceTcpOptions
	tfPlan.TcpOptions.As(ctx, &tcpOptions, basetypes.ObjectAsOptions{})
	var testDefinitions uriResourceTestDefinitions
	tfPlan.TestDefinitions.As(ctx, &testDefinitions, basetypes.ObjectAsOptions{})
	var planPlatformOptions uriResourcePlatformOptions
	testDefinitions.PlatformOptions.As(ctx, &planPlatformOptions, basetypes.ObjectAsOptions{})
	var locationOptions []uriResourceProbeLocation
	testDefinitions.LocationOptions.ElementsAs(ctx, &locationOptions, false)

	// Create our input request.
	createInput := swoClient.CreateUriInput{
		Name:       tfPlan.Name.ValueString(),
		IpOrDomain: tfPlan.Host.ValueString(),
		PingOptions: &swoClient.UriPingOptionsInput{
			Enabled: planOptions.IsPingEnabled.ValueBool(),
		},
		TcpOptions: &swoClient.UriTcpOptionsInput{
			Enabled:        planOptions.IsTcpEnabled.ValueBool(),
			Port:           int(tcpOptions.Port.ValueInt64()),
			StringToExpect: tcpOptions.StringToExpect.ValueStringPointer(),
			StringToSend:   tcpOptions.StringToSend.ValueStringPointer(),
		},
		TestDefinitions: swoClient.UriTestDefinitionsInput{
			PlatformOptions: &swoClient.ProbePlatformOptionsInput{
				TestFromAll: planPlatformOptions.TestFromAll.ValueBoolPointer(),
				ProbePlatforms: convertArray(planPlatformOptions.Platforms.Elements(),
					func(s attr.Value) swoClient.ProbePlatform { return swoClient.ProbePlatform(attrValueToString(s)) }),
			},
			TestFrom: swoClient.ProbeLocationInput{
				Type: swoClient.ProbeLocationType(testDefinitions.TestFromLocation.ValueString()),
				Values: convertArray(locationOptions,
					func(v uriResourceProbeLocation) string { return v.Value.ValueString() }),
			},
			TestIntervalInSeconds: swoClientTypes.TestIntervalInSeconds(int(testDefinitions.TestIntervalInSeconds.ValueInt64())),
		},
	}

	// Create the Uri...
	newUri, err := r.client.UriService().Create(ctx, createInput)
	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error creating uri '%s' - error: %s", tfPlan.Name, err))
		return
	}

	tfPlan.Id = types.StringValue(newUri.Id)
	resp.Diagnostics.Append(resp.State.Set(ctx, tfPlan)...)
}

func (r *uriResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var tfState uriResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &tfState)...)

	if resp.Diagnostics.HasError() {
		return
	}

	uri, err := ReadRetry(ctx, tfState.Id.ValueString(), r.client.UriService().Read)

	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error reading uri %s. error: %s", tfState.Id, err))
		return
	}

	// Update the Terraform state.
	tfState.Host = types.StringValue(uri.Host)
	tfState.Name = types.StringPointerValue(uri.Name)

	// Options
	optionsElementTypes := UriResourceOptionsAttributeTypes()
	optionsElement := uriResourceOptions{
		IsPingEnabled: types.BoolValue(uri.Options.IsPingEnabled),
		IsTcpEnabled:  types.BoolValue(uri.Options.IsTcpEnabled),
	}
	tfOptions, _ := types.ObjectValueFrom(ctx, optionsElementTypes, optionsElement)
	tfState.Options = tfOptions

	// TcpOptions
	tcpElementTypes := UriTcpOptionsAttributeTypes()
	if uri.TcpOptions != nil {
		tcpOptions := uri.TcpOptions
		tcpElement := uriResourceTcpOptions{
			Port:           types.Int64Value(int64(tcpOptions.Port)),
			StringToExpect: types.StringNull(),
			StringToSend:   types.StringNull(),
		}
		if tcpOptions.StringToExpect != nil {
			tcpElement.StringToExpect = types.StringValue(*tcpOptions.StringToExpect)
		}
		if tcpOptions.StringToSend != nil {
			tcpElement.StringToSend = types.StringValue(*tcpOptions.StringToSend)
		}

		tfTcpOptions, _ := types.ObjectValueFrom(ctx, tcpElementTypes, tcpElement)
		tfState.TcpOptions = tfTcpOptions
	} else {
		tfTcpOptions := types.ObjectNull(tcpElementTypes)
		tfState.TcpOptions = tfTcpOptions
	}

	// TestDefinitions
	testDefs := uri.TestDefinitions
	platformElementTypes := UriPlatformOptionsAttributeTypes()
	locationOptsElementTypes := UriProbeLocationAttributeTypes()
	testDefsElementTypes := UriTestDefAttributeTypes()

	testDefsElements := uriResourceTestDefinitions{
		TestFromLocation:      types.StringNull(),
		LocationOptions:       types.SetUnknown(types.ObjectType{AttrTypes: locationOptsElementTypes}),
		TestIntervalInSeconds: types.Int64Null(),
		PlatformOptions:       types.ObjectNull(platformElementTypes),
	}
	if testDefs.TestFromLocation != nil {
		testDefsElements.TestFromLocation = types.StringValue(string(*testDefs.TestFromLocation))
	}

	var diags diag.Diagnostics
	var locationOptsElements []attr.Value
	for _, x := range testDefs.LocationOptions {
		objectValue, objectDiags := types.ObjectValueFrom(
			ctx,
			locationOptsElementTypes,
			uriResourceProbeLocation{
				Type:  types.StringValue(string(x.Type)),
				Value: types.StringValue(x.Value),
			},
		)

		locationOptsElements = append(locationOptsElements, objectValue)
		diags = append(diags, objectDiags...)
	}
	locationOpts, _ := types.SetValueFrom(ctx, types.ObjectType{AttrTypes: locationOptsElementTypes}, locationOptsElements)
	testDefsElements.LocationOptions = locationOpts

	if testDefs.TestIntervalInSeconds != nil {
		testDefsElements.TestIntervalInSeconds = types.Int64Value(int64(*testDefs.TestIntervalInSeconds))
	}

	if testDefs.PlatformOptions != nil {
		listValue, _ := types.SetValueFrom(ctx, types.StringType, testDefs.PlatformOptions.Platforms)
		platformElements := uriResourcePlatformOptions{
			TestFromAll: types.BoolValue(testDefs.PlatformOptions.TestFromAll),
			Platforms:   listValue,
		}

		tfPlatformOptions, _ := types.ObjectValueFrom(ctx, platformElementTypes, platformElements)
		testDefsElements.PlatformOptions = tfPlatformOptions
	}

	objectValue, _ := types.ObjectValueFrom(ctx, testDefsElementTypes, testDefsElements)
	tfState.TestDefinitions = objectValue
	resp.Diagnostics.Append(resp.State.Set(ctx, tfState)...)
}

func (r *uriResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var tfPlan, tfState *uriResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &tfPlan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &tfState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var planOptions uriResourceOptions
	tfPlan.Options.As(ctx, &planOptions, basetypes.ObjectAsOptions{})
	var tcpOptions uriResourceTcpOptions
	tfPlan.TcpOptions.As(ctx, &tcpOptions, basetypes.ObjectAsOptions{})
	var testDefinitions uriResourceTestDefinitions
	tfPlan.TestDefinitions.As(ctx, &testDefinitions, basetypes.ObjectAsOptions{})
	var planPlatformOptions uriResourcePlatformOptions
	testDefinitions.PlatformOptions.As(ctx, &planPlatformOptions, basetypes.ObjectAsOptions{})
	var locationOptions []uriResourceProbeLocation
	testDefinitions.LocationOptions.ElementsAs(ctx, &locationOptions, false)

	// Create our input request.
	updateInput := swoClient.UpdateUriInput{
		Id:         tfState.Id.ValueString(),
		Name:       tfPlan.Name.ValueString(),
		IpOrDomain: tfPlan.Host.ValueString(),
		PingOptions: &swoClient.UriPingOptionsInput{
			Enabled: planOptions.IsPingEnabled.ValueBool(),
		},
		TcpOptions: &swoClient.UriTcpOptionsInput{
			Enabled:        planOptions.IsTcpEnabled.ValueBool(),
			Port:           int(tcpOptions.Port.ValueInt64()),
			StringToExpect: tcpOptions.StringToExpect.ValueStringPointer(),
			StringToSend:   tcpOptions.StringToSend.ValueStringPointer(),
		},
		TestDefinitions: swoClient.UriTestDefinitionsInput{
			PlatformOptions: &swoClient.ProbePlatformOptionsInput{
				TestFromAll: planPlatformOptions.TestFromAll.ValueBoolPointer(),
				ProbePlatforms: convertArray(planPlatformOptions.Platforms.Elements(),
					func(s attr.Value) swoClient.ProbePlatform { return swoClient.ProbePlatform(attrValueToString(s)) }),
			},
			TestFrom: swoClient.ProbeLocationInput{
				Type: swoClient.ProbeLocationType(testDefinitions.TestFromLocation.ValueString()),
				Values: convertArray(locationOptions,
					func(v uriResourceProbeLocation) string { return v.Value.ValueString() }),
			},

			TestIntervalInSeconds: swoClientTypes.TestIntervalInSeconds(int(testDefinitions.TestIntervalInSeconds.ValueInt64())),
		},
	}

	var platforms []string
	planPlatformOptions.Platforms.ElementsAs(ctx, &platforms, false)

	bUriToMatch, err := json.Marshal(map[string]interface{}{
		"id":   updateInput.Id,
		"name": updateInput.Name,
		"host": updateInput.IpOrDomain,
		"options": map[string]interface{}{
			"isPingEnabled": updateInput.PingOptions.Enabled,
			"isTcpEnabled":  updateInput.TcpOptions.Enabled,
		},
		"tcpOptions": map[string]interface{}{
			"port":           updateInput.TcpOptions.Port,
			"stringToExpect": updateInput.TcpOptions.StringToExpect,
			"stringToSend":   updateInput.TcpOptions.StringToSend,
		},
		"testDefinitions": map[string]interface{}{
			"testFromLocation":      updateInput.TestDefinitions.TestFrom.Type,
			"testIntervalInSeconds": updateInput.TestDefinitions.TestIntervalInSeconds,
			"platformOptions": map[string]interface{}{
				"testFromAll": planPlatformOptions.TestFromAll.ValueBoolPointer(),
				"platforms":   platforms,
			},
		},
	})

	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error marshaling uri result to match %s - %s", tfState.Id, err))
		return
	}

	var readUriResultToMatch swoClient.ReadUriResult

	err = json.Unmarshal(bUriToMatch, &readUriResultToMatch)
	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error unmarshaling uri result to match %s - %s", tfState.Id, err))
		return
	}

	// Update the Uri...
	err = r.client.UriService().Update(ctx, updateInput)

	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error updating uri %s. err: %s", tfState.Id, err))
		return
	}

	// Updates are eventually consistent. Retry until the URI we read and the URI we are updating match.
	_, err = BackoffRetry(func() (*swoClient.ReadUriResult, error) {
		// Read the Uri...
		uri, err := r.client.UriService().Read(ctx, tfState.Id.ValueString())

		if err != nil {
			return nil, backoff.Permanent(err)
		}

		//Set unsupported values
		readUriResultToMatch.Typename = uri.Typename
		readUriResultToMatch.Options.IsHttpEnabled = uri.Options.IsHttpEnabled
		readUriResultToMatch.HttpOptions = uri.HttpOptions
		readUriResultToMatch.Tags = uri.Tags
		readUriResultToMatch.TestDefinitions.LocationOptions = uri.TestDefinitions.LocationOptions

		match := reflect.DeepEqual(&readUriResultToMatch, uri)

		// Updated entity properties don't match, retry
		if !match {
			return nil, ErrNonMatchingEntites
		}

		return uri, nil
	})

	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error updating uri %s. err: %s", tfState.Id, err))
		return
	}

	// Save to Terraform state.
	tfPlan.Id = tfState.Id
	resp.Diagnostics.Append(resp.State.Set(ctx, &tfPlan)...)
}

func (r *uriResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var tfState uriResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &tfState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the Uri...
	if err := r.client.UriService().Delete(ctx, tfState.Id.ValueString()); err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error deleting uri %s - %s", tfState.Id, err))
	}
}

func (r *uriResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
