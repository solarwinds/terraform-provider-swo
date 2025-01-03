package provider

import (
	"context"
	"errors"
	"fmt"
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
	client, _ := req.ProviderData.(*swoClient.Client)
	r.client = client
}

func (r *uriResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var tfPlan uriResourceModel
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
			StringToExpect: tfPlan.TcpOptions.StringToExpect.ValueStringPointer(),
			StringToSend:   tfPlan.TcpOptions.StringToSend.ValueStringPointer(),
		},
		TestDefinitions: swoClient.UriTestDefinitionsInput{
			PlatformOptions: &swoClient.ProbePlatformOptionsInput{
				TestFromAll: tfPlan.TestDefinitions.PlatformOptions.TestFromAll.ValueBoolPointer(),
				ProbePlatforms: convertArray(tfPlan.TestDefinitions.PlatformOptions.Platforms,
					func(v string) swoClient.ProbePlatform { return swoClient.ProbePlatform(v) }),
			},
			TestFrom: swoClient.ProbeLocationInput{
				Type: swoClient.ProbeLocationType(tfPlan.TestDefinitions.TestFromLocation.ValueString()),
				Values: convertArray(tfPlan.TestDefinitions.LocationOptions,
					func(v uriResourceProbeLocation) string { return v.Value.ValueString() }),
			},
			TestIntervalInSeconds: swoClientTypes.TestIntervalInSeconds(int(tfPlan.TestDefinitions.TestIntervalInSeconds.ValueInt64())),
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

	// Creates and Updates are eventually consistant. Retry until the URI's entity Id is returned.
	uri, err := r.backoffRetry(func() (*swoClient.ReadUriResult, error) {
		// Read the Uri...
		uri, err := r.client.UriService().Read(ctx, tfState.Id.ValueString())

		if err != nil {
			// The entity is still being created, retry
			if errors.Is(err, swoClient.ErrEntityIdNil) {
				return nil, swoClient.ErrEntityIdNil
			}

			return nil, backoff.Permanent(err)
		}

		return uri, nil
	})

	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error reading uri %s. error: %s", tfState.Id, err))
		return
	}

	// Update the Terraform state.
	tfState.Host = types.StringValue(uri.Host)
	tfState.Name = types.StringPointerValue(uri.Name)

	// Options
	options := uri.Options
	tfState.Options = &uriResourceOptions{
		IsPingEnabled: types.BoolValue(options.IsPingEnabled),
		IsTcpEnabled:  types.BoolValue(options.IsTcpEnabled),
	}

	// TcpOptions
	if uri.TcpOptions != nil {
		tcpOptions := uri.TcpOptions
		tfState.TcpOptions = &uriResourceTcpOptions{
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
	tfState.TestDefinitions = &uriResourceTestDefinitions{}

	if testDefs.PlatformOptions != nil {
		tfState.TestDefinitions.PlatformOptions = &uriResourcePlatformOptions{
			TestFromAll: types.BoolValue(testDefs.PlatformOptions.TestFromAll),
			Platforms:   testDefs.PlatformOptions.Platforms,
		}
	} else {
		tfState.TestDefinitions.PlatformOptions = nil
	}

	if testDefs.TestFromLocation != nil {
		tfState.TestDefinitions.TestFromLocation = types.StringValue(string(*testDefs.TestFromLocation))
	} else {
		tfState.TestDefinitions.TestFromLocation = types.StringNull()
	}

	if testDefs.TestIntervalInSeconds != nil {
		tfState.TestDefinitions.TestIntervalInSeconds = types.Int64Value(int64(*testDefs.TestIntervalInSeconds))
	} else {
		tfState.TestDefinitions.TestIntervalInSeconds = types.Int64Null()
	}

	var locOpts = []uriResourceProbeLocation{}
	for _, x := range testDefs.LocationOptions {
		locOpts = append(locOpts, uriResourceProbeLocation{
			Type:  types.StringValue(string(x.Type)),
			Value: types.StringValue(x.Value),
		})
	}
	tfState.TestDefinitions.LocationOptions = locOpts

	resp.Diagnostics.Append(resp.State.Set(ctx, tfState)...)
}

func (r *uriResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var tfPlan, tfState *uriResourceModel
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
			StringToExpect: tfPlan.TcpOptions.StringToExpect.ValueStringPointer(),
			StringToSend:   tfPlan.TcpOptions.StringToSend.ValueStringPointer(),
		},
		TestDefinitions: swoClient.UriTestDefinitionsInput{
			PlatformOptions: &swoClient.ProbePlatformOptionsInput{
				TestFromAll: tfPlan.TestDefinitions.PlatformOptions.TestFromAll.ValueBoolPointer(),
				ProbePlatforms: convertArray(tfPlan.TestDefinitions.PlatformOptions.Platforms,
					func(v string) swoClient.ProbePlatform { return swoClient.ProbePlatform(v) }),
			},
			TestFrom: swoClient.ProbeLocationInput{
				Type: swoClient.ProbeLocationType(tfPlan.TestDefinitions.TestFromLocation.ValueString()),
				Values: convertArray(tfPlan.TestDefinitions.LocationOptions,
					func(v uriResourceProbeLocation) string { return v.Value.ValueString() }),
			},

			TestIntervalInSeconds: swoClientTypes.TestIntervalInSeconds(int(tfPlan.TestDefinitions.TestIntervalInSeconds.ValueInt64())),
		},
	}

	// Update the Uri...
	err := r.client.UriService().Update(ctx, updateInput)

	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error updating uri %s. err: %s", tfState.Id, err))
		return
	}

	// Updates are eventually consistant. Retry until the URI we read and the URI we are updating match.
	_, err = r.backoffRetry(func() (*swoClient.ReadUriResult, error) {
		// Read the Uri...
		uri, err := r.client.UriService().Read(ctx, tfState.Id.ValueString())

		if err != nil {
			return nil, backoff.Permanent(err)
		}

		match := (updateInput.Name == *uri.Name) &&
			(updateInput.IpOrDomain == uri.Host) &&
			(updateInput.PingOptions.Enabled == uri.Options.IsPingEnabled) &&
			(updateInput.TcpOptions.Enabled == uri.Options.IsTcpEnabled) &&
			(updateInput.TcpOptions.Port == uri.TcpOptions.Port) &&
			(updateInput.TcpOptions.StringToExpect == uri.TcpOptions.StringToExpect) &&
			(updateInput.TcpOptions.StringToSend == uri.TcpOptions.StringToSend) &&
			(updateInput.TestDefinitions.TestIntervalInSeconds == *uri.TestDefinitions.TestIntervalInSeconds) &&
			(updateInput.TestDefinitions.PlatformOptions.TestFromAll == &uri.TestDefinitions.PlatformOptions.TestFromAll) &&
			(reflect.DeepEqual(updateInput.TestDefinitions.PlatformOptions.ProbePlatforms, uri.TestDefinitions.PlatformOptions.Platforms)) &&
			(updateInput.TestDefinitions.TestFrom.GetType() == *uri.TestDefinitions.GetTestFromLocation()) &&
			(reflect.DeepEqual(updateInput.TestDefinitions.TestFrom.GetValues(), uri.TestDefinitions.GetLocationOptions()))

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

func (r *uriResource) backoffRetry(operation func() (*swoClient.ReadUriResult, error)) (*swoClient.ReadUriResult, error) {
	expBackoff := backoff.NewExponentialBackOff()
	expBackoff.MaxInterval = expBackoffMaxInterval

	return backoff.Retry(context.Background(), operation, backoff.WithBackOff(expBackoff), backoff.WithMaxElapsedTime(expBackoffMaxElapsed))
}
