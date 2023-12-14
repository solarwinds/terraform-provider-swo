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
var _ resource.Resource = &UriResource{}
var _ resource.ResourceWithConfigure = &UriResource{}
var _ resource.ResourceWithImportState = &UriResource{}

func NewUriResource() resource.Resource {
	return &UriResource{}
}

// Defines the resource implementation.
type UriResource struct {
	client *swoClient.Client
}

func (r *UriResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_uri"
}

func (r *UriResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	tflog.Trace(ctx, "UriResource: Configure")

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

func (r *UriResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Trace(ctx, "UriResource: Create")

	var tfPlan UriResourceModel

	// Read the Terraform plan.
	resp.Diagnostics.Append(req.Plan.Get(ctx, &tfPlan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create our input request.
	createInput := swoClient.CreateUriInput{
		Name:       tfPlan.Name.ValueString(),
		IpOrDomain: tfPlan.Host.ValueString(),
		PingOptions: &swoClient.UriPingOptionsInput{
			Enabled: tfPlan.Options.IsPingEnabled.ValueBool(),
		},
		TcpOptions: &swoClient.UriTcpOptionsInput{
			Enabled:        tfPlan.Options.IsTcpEnabled.ValueBool(),
			Port:           int(tfPlan.TcpOptions.Port.ValueInt64()),
			StringToExpect: swoClient.Ptr(tfPlan.TcpOptions.StringToExpect.ValueString()),
			StringToSend:   swoClient.Ptr(tfPlan.TcpOptions.StringToSend.ValueString()),
		},
		TestDefinitions: swoClient.UriTestDefinitionsInput{
			PlatformOptions: &swoClient.ProbePlatformOptionsInput{
				TestFromAll: swoClient.Ptr(tfPlan.TestDefinitions.PlatformOptions.TestFromAll.ValueBool()),
				ProbePlatforms: convertArray(tfPlan.TestDefinitions.PlatformOptions.Platforms,
					func(v string) swoClient.ProbePlatform { return swoClient.ProbePlatform(v) }),
			},
			TestFrom: swoClient.ProbeLocationInput{
				Type: swoClient.ProbeLocationType(tfPlan.TestDefinitions.TestFromLocation.ValueString()),
				Values: convertArray(tfPlan.TestDefinitions.LocationOptions,
					func(v UriResourceProbeLocation) string { return v.Value.ValueString() }),
			},
			TestIntervalInSeconds: int(tfPlan.TestDefinitions.TestIntervalInSeconds.ValueInt64()),
		},
	}

	// Create the Uri...
	newUri, err := r.client.UriService().Create(ctx, createInput)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("error creating uri '%s' - error: %s",
			tfPlan.Name,
			err))
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("uri %s created successfully: id=%s", tfPlan.Name, newUri.Id))
	tfPlan.Id = types.StringValue(newUri.Id)
	resp.Diagnostics.Append(resp.State.Set(ctx, tfPlan)...)
}

func (r *UriResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Trace(ctx, "UriResource: Read")

	var tfState UriResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &tfState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read the Uri...
	tflog.Trace(ctx, fmt.Sprintf("read uri: id=%s", tfState.Id))
	uri, err := r.client.UriService().Read(ctx, tfState.Id.ValueString())

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("error reading uri %s. error: %s",
			tfState.Id, err))
		return
	}

	// Update the Terraform state.
	tfState.Host = types.StringValue(uri.Host)

	if uri.Name != nil {
		tfState.Name = types.StringValue(*uri.Name)
	}

	// Options
	options := uri.Options
	tfState.Options = &UriResourceOptions{
		IsPingEnabled: types.BoolValue(options.IsPingEnabled),
		IsTcpEnabled:  types.BoolValue(options.IsTcpEnabled),
	}

	// TcpOptions
	if uri.TcpOptions != nil {
		tcpOptions := uri.TcpOptions
		tfState.TcpOptions = &UriResourceTcpOptions{
			Port: types.Int64Value(int64(tcpOptions.Port)),
		}
		if tcpOptions.StringToSend != nil {
			tfState.TcpOptions.StringToSend = types.StringValue(*tcpOptions.StringToSend)
		}
		if tcpOptions.StringToExpect != nil {
			tfState.TcpOptions.StringToExpect = types.StringValue(*tcpOptions.StringToExpect)
		}
	} else {
		tfState.TcpOptions = nil
	}

	// TestDefinitions
	testDefs := uri.TestDefinitions
	tfState.TestDefinitions = &UriResourceTestDefinitions{}

	if testDefs.PlatformOptions != nil {
		tfState.TestDefinitions.PlatformOptions = &UriResourcePlatformOptions{
			TestFromAll: types.BoolValue(testDefs.PlatformOptions.TestFromAll),
			Platforms:   testDefs.PlatformOptions.Platforms,
		}
	} else {
		tfState.TestDefinitions.PlatformOptions = nil
	}
	if testDefs.TestFromLocation != nil {
		tfState.TestDefinitions.TestFromLocation = types.StringValue(string(*testDefs.TestFromLocation))
	}
	if testDefs.TestIntervalInSeconds != nil {
		tfState.TestDefinitions.TestIntervalInSeconds = types.Int64Value(int64(*testDefs.TestIntervalInSeconds))
	}
	var locOpts = []UriResourceProbeLocation{}
	for _, x := range testDefs.LocationOptions {
		locOpts = append(locOpts, UriResourceProbeLocation{
			Type:  types.StringValue(string(x.Type)),
			Value: types.StringValue(x.Value),
		})
	}
	tfState.TestDefinitions.LocationOptions = locOpts

	tflog.Trace(ctx, fmt.Sprintf("read uri success: id=%s", uri.Id))
	resp.Diagnostics.Append(resp.State.Set(ctx, tfState)...)
}

func (r *UriResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var tfPlan, tfState *UriResourceModel

	// Read the Terraform plan and state data.
	resp.Diagnostics.Append(req.Plan.Get(ctx, &tfPlan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &tfState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create our input request.
	updateInput := swoClient.UpdateUriInput{
		Id:         tfState.Id.ValueString(),
		Name:       tfPlan.Name.ValueString(),
		IpOrDomain: tfPlan.Host.ValueString(),
		PingOptions: &swoClient.UriPingOptionsInput{
			Enabled: tfPlan.Options.IsPingEnabled.ValueBool(),
		},
		TcpOptions: &swoClient.UriTcpOptionsInput{
			Enabled:        tfPlan.Options.IsTcpEnabled.ValueBool(),
			Port:           int(tfPlan.TcpOptions.Port.ValueInt64()),
			StringToExpect: swoClient.Ptr(tfPlan.TcpOptions.StringToExpect.ValueString()),
			StringToSend:   swoClient.Ptr(tfPlan.TcpOptions.StringToSend.ValueString()),
		},
		TestDefinitions: swoClient.UriTestDefinitionsInput{
			PlatformOptions: &swoClient.ProbePlatformOptionsInput{
				TestFromAll: swoClient.Ptr(tfPlan.TestDefinitions.PlatformOptions.TestFromAll.ValueBool()),
				ProbePlatforms: convertArray(tfPlan.TestDefinitions.PlatformOptions.Platforms,
					func(v string) swoClient.ProbePlatform { return swoClient.ProbePlatform(v) }),
			},
			TestFrom: swoClient.ProbeLocationInput{
				Type: swoClient.ProbeLocationType(tfPlan.TestDefinitions.TestFromLocation.ValueString()),
				Values: convertArray(tfPlan.TestDefinitions.LocationOptions,
					func(v UriResourceProbeLocation) string { return v.Value.ValueString() }),
			},
			TestIntervalInSeconds: int(tfPlan.TestDefinitions.TestIntervalInSeconds.ValueInt64()),
		},
	}

	// Update the Uri...
	tflog.Trace(ctx, fmt.Sprintf("updating uri with id: %s", tfState.Id))
	err := r.client.UriService().Update(ctx, updateInput)

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("error updating uri %s. err: %s", tfState.Id, err))
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("uri '%s' updated successfully", tfState.Id))

	// Save to Terraform state.
	tfPlan.Id = tfState.Id
	resp.Diagnostics.Append(resp.State.Set(ctx, &tfPlan)...)
}

func (r *UriResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state UriResourceModel

	// Read existing Terraform state.
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// // Delete the Uri...
	tflog.Trace(ctx, fmt.Sprintf("deleting uri: id=%s", state.Id))
	if err := r.client.UriService().Delete(ctx, state.Id.ValueString()); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("error deleting uri %s - %s", state.Id, err))
		return
	}
	tflog.Trace(ctx, fmt.Sprintf("uri deleted: id=%s", state.Id))
}

func (r *UriResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
