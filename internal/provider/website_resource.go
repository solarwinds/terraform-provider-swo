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
var _ resource.Resource = &WebsiteResource{}
var _ resource.ResourceWithConfigure = &WebsiteResource{}
var _ resource.ResourceWithImportState = &WebsiteResource{}

func NewWebsiteResource() resource.Resource {
	return &WebsiteResource{}
}

// Defines the resource implementation.
type WebsiteResource struct {
	client *swoClient.Client
}

func (r *WebsiteResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_website"
}

func (r *WebsiteResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	tflog.Trace(ctx, "WebsiteResource: Configure")

	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*swoClient.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Invalid Resource Client Type",
			fmt.Sprintf("expected *swoClient.Client but received: %T.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *WebsiteResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Trace(ctx, "WebsiteResource: Create")

	var tfPlan WebsiteResourceModel

	// Read the Terraform plan.
	resp.Diagnostics.Append(req.Plan.Get(ctx, &tfPlan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create our input request.
	createInput := swoClient.CreateWebsiteInput{
		Name: tfPlan.Name.ValueString(),
		Url:  tfPlan.Url.ValueString(),
		AvailabilityCheckSettings: &swoClient.AvailabilityCheckSettingsInput{
			CheckForString: &swoClient.CheckForStringInput{
				Operator: swoClient.CheckStringOperator(tfPlan.Monitoring.Availability.CheckForString.Operator.ValueString()),
				Value:    tfPlan.Monitoring.Availability.CheckForString.Value.ValueString(),
			},
			TestIntervalInSeconds: int(tfPlan.Monitoring.Availability.TestIntervalInSeconds.ValueInt64()),
			Protocols: convertArray(tfPlan.Monitoring.Availability.Protocols, func(s string) swoClient.WebsiteProtocol {
				return swoClient.WebsiteProtocol(s)
			}),
			PlatformOptions: &swoClient.ProbePlatformOptionsInput{
				TestFromAll: swoClient.Ptr(tfPlan.Monitoring.Availability.PlatformOptions.TestFromAll.ValueBool()),
				ProbePlatforms: convertArray(tfPlan.Monitoring.Availability.PlatformOptions.Platforms, func(s string) swoClient.ProbePlatform {
					return swoClient.ProbePlatform(s)
				}),
			},
			TestFrom: swoClient.ProbeLocationInput{
				Type: swoClient.ProbeLocationType(tfPlan.Monitoring.Availability.TestFromLocation.ValueString()),
				Values: convertArray(tfPlan.Monitoring.Availability.LocationOptions, func(p ProbeLocation) string {
					return p.Value.ValueString()
				}),
			},
			Ssl: &swoClient.SslMonitoringInput{
				Enabled:                        swoClient.Ptr(tfPlan.Monitoring.Availability.SSL.Enabled.ValueBool()),
				DaysPriorToExpiration:          swoClient.Ptr(int(tfPlan.Monitoring.Availability.SSL.DaysPriorToExpiration.ValueInt64())),
				IgnoreIntermediateCertificates: swoClient.Ptr(tfPlan.Monitoring.Availability.SSL.IgnoreIntermediateCertificates.ValueBool()),
			},
			CustomHeaders: convertArray(tfPlan.Monitoring.CustomHeaders, func(h CustomHeader) swoClient.CustomHeaderInput {
				return swoClient.CustomHeaderInput{
					Name:  h.Name.ValueString(),
					Value: h.Value.ValueString(),
				}
			}),
		},
		Rum: &swoClient.RumMonitoringInput{
			ApdexTimeInSeconds: swoClient.Ptr(int(tfPlan.Monitoring.Rum.ApdexTimeInSeconds.ValueInt64())),
			Spa:                swoClient.Ptr(tfPlan.Monitoring.Rum.Spa.ValueBool()),
		},
	}

	// Create the Website...
	result, err := r.client.WebsiteService().Create(ctx, createInput)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("error creating website '%s' - error: %s",
			tfPlan.Name,
			err))
		return
	}

	// Get the latest Website state from the server so we can get the 'snippet' field. Ideally we need to update
	// the API to return the 'snippet' field in the create response.
	if website, err := r.client.WebsiteService().Read(ctx, result.Id); err != nil {
		resp.Diagnostics.AddWarning("Client Error", fmt.Sprintf("error capturing RUM snippit for Website '%s' - error: %s",
			tfPlan.Name,
			err))
	} else {
		tfPlan.Monitoring.Rum.Snippet = types.StringValue(*website.Monitoring.Rum.Snippet)
	}

	tflog.Trace(ctx, fmt.Sprintf("website %s created successfully - id=%s", tfPlan.Name, result.Id))
	tfPlan.Id = types.StringValue(result.Id)
	resp.Diagnostics.Append(resp.State.Set(ctx, tfPlan)...)
}

func (r *WebsiteResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Trace(ctx, "WebsiteResource: Read")

	var tfState WebsiteResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &tfState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read the Website...
	tflog.Trace(ctx, fmt.Sprintf("read website with id: %s", tfState.Id))
	website, err := r.client.WebsiteService().Read(ctx, tfState.Id.ValueString())

	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error reading Website %s. error: %s", tfState.Id, err))
		return
	}

	// Update the Terraform state with latest values from the server.
	tfState.Url = types.StringValue(website.Url)
	if website.Name != nil {
		tfState.Name = types.StringValue(*website.Name)
	}

	if website.Monitoring != nil {
		monitoring := website.Monitoring
		tfState.Monitoring = &WebsiteMonitoring{}

		if monitoring.Availability != nil {
			tfState.Monitoring.Availability = AvailabilityMonitoring{}
			availability := monitoring.Availability

			if availability.CheckForString != nil {
				tfState.Monitoring.Availability.CheckForString = CheckForStringType{
					Operator: types.StringValue(string(availability.CheckForString.Operator)),
					Value:    types.StringValue(availability.CheckForString.Value),
				}
			}

			if availability.TestIntervalInSeconds != nil {
				tfState.Monitoring.Availability.TestIntervalInSeconds = types.Int64Value(int64(*availability.TestIntervalInSeconds))
			}

			tfState.Monitoring.Availability.Protocols = convertArray(availability.Protocols, func(s swoClient.WebsiteProtocol) string {
				return string(s)
			})

			if availability.PlatformOptions != nil {
				tfState.Monitoring.Availability.PlatformOptions = PlatformOptions{
					TestFromAll: types.BoolValue(availability.PlatformOptions.TestFromAll),
					Platforms:   availability.PlatformOptions.Platforms,
				}
			}

			if availability.TestFromLocation != nil {
				tfState.Monitoring.Availability.TestFromLocation = types.StringValue(string(*availability.TestFromLocation))
			}

			if availability.LocationOptions != nil {
				var locOpts []ProbeLocation
				for _, p := range availability.LocationOptions {
					locOpts = append(locOpts, ProbeLocation{
						Type:  types.StringValue(string(p.Type)),
						Value: types.StringValue(string(p.Value)),
					})
				}
				tfState.Monitoring.Availability.LocationOptions = locOpts
			}

			if availability.Ssl != nil {
				tfState.Monitoring.Availability.SSL = SslMonitoring{
					Enabled:                        types.BoolValue(availability.Ssl.Enabled),
					IgnoreIntermediateCertificates: types.BoolValue(availability.Ssl.IgnoreIntermediateCertificates),
				}
				if availability.Ssl.DaysPriorToExpiration != nil {
					tfState.Monitoring.Availability.SSL.DaysPriorToExpiration = types.Int64Value(int64(*availability.Ssl.DaysPriorToExpiration))
				}
			}
		}

		if monitoring.CustomHeaders != nil {
			var customHeaders []CustomHeader
			for _, h := range monitoring.CustomHeaders {
				customHeaders = append(customHeaders, CustomHeader{
					Name:  types.StringValue(h.Name),
					Value: types.StringValue(h.Value),
				})
			}
			tfState.Monitoring.CustomHeaders = customHeaders
		}

		if monitoring.Rum != nil {
			tfState.Monitoring.Rum = RumMonitoring{
				Spa: types.BoolValue(monitoring.Rum.Spa),
			}

			if monitoring.Rum.ApdexTimeInSeconds != nil {
				tfState.Monitoring.Rum.ApdexTimeInSeconds = types.Int64Value(int64(*monitoring.Rum.ApdexTimeInSeconds))
			}

			if monitoring.Rum.Snippet != nil {
				tfState.Monitoring.Rum.Snippet = types.StringValue(*monitoring.Rum.Snippet)
			}
		}
	}

	tflog.Trace(ctx, fmt.Sprintf("read Website success: %s", *website.Name))
	resp.Diagnostics.Append(resp.State.Set(ctx, tfState)...)
}

func (r *WebsiteResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var tfPlan, tfState *WebsiteResourceModel

	// Read the Terraform plan and state data.
	resp.Diagnostics.Append(req.Plan.Get(ctx, &tfPlan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &tfState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Update the Website...
	tflog.Trace(ctx, fmt.Sprintf("updating Website with id: %s", tfState.Id))
	err := r.client.WebsiteService().Update(ctx, swoClient.UpdateWebsiteInput{
		Id:   tfState.Id.ValueString(),
		Name: tfPlan.Name.ValueString(),
		Url:  tfPlan.Url.ValueString(),
		AvailabilityCheckSettings: &swoClient.AvailabilityCheckSettingsInput{
			CheckForString: &swoClient.CheckForStringInput{
				Operator: swoClient.CheckStringOperator(tfPlan.Monitoring.Availability.CheckForString.Operator.ValueString()),
				Value:    tfPlan.Monitoring.Availability.CheckForString.Value.ValueString(),
			},
			TestIntervalInSeconds: int(tfPlan.Monitoring.Availability.TestIntervalInSeconds.ValueInt64()),
			Protocols: convertArray(tfPlan.Monitoring.Availability.Protocols, func(s string) swoClient.WebsiteProtocol {
				return swoClient.WebsiteProtocol(s)
			}),
			PlatformOptions: &swoClient.ProbePlatformOptionsInput{
				TestFromAll: swoClient.Ptr(tfPlan.Monitoring.Availability.PlatformOptions.TestFromAll.ValueBool()),
				ProbePlatforms: convertArray(tfPlan.Monitoring.Availability.PlatformOptions.Platforms, func(s string) swoClient.ProbePlatform {
					return swoClient.ProbePlatform(s)
				}),
			},
			TestFrom: swoClient.ProbeLocationInput{
				Type: swoClient.ProbeLocationType(tfPlan.Monitoring.Availability.TestFromLocation.ValueString()),
				Values: convertArray(tfPlan.Monitoring.Availability.LocationOptions, func(p ProbeLocation) string {
					return p.Value.ValueString()
				}),
			},
			Ssl: &swoClient.SslMonitoringInput{
				Enabled:                        swoClient.Ptr(tfPlan.Monitoring.Availability.SSL.Enabled.ValueBool()),
				DaysPriorToExpiration:          swoClient.Ptr(int(tfPlan.Monitoring.Availability.SSL.DaysPriorToExpiration.ValueInt64())),
				IgnoreIntermediateCertificates: swoClient.Ptr(tfPlan.Monitoring.Availability.SSL.IgnoreIntermediateCertificates.ValueBool()),
			},
			CustomHeaders: convertArray(tfPlan.Monitoring.CustomHeaders, func(h CustomHeader) swoClient.CustomHeaderInput {
				return swoClient.CustomHeaderInput{
					Name:  h.Name.ValueString(),
					Value: h.Value.ValueString(),
				}
			}),
		},
		Rum: &swoClient.RumMonitoringInput{
			ApdexTimeInSeconds: swoClient.Ptr(int(tfPlan.Monitoring.Rum.ApdexTimeInSeconds.ValueInt64())),
			Spa:                swoClient.Ptr(tfPlan.Monitoring.Rum.Spa.ValueBool()),
		},
	})

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("error updating website %s. err: %s", tfState.Id, err))
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("website '%s' updated successfully", tfState.Id))

	// Save to Terraform state.
	tfPlan.Id = tfState.Id

	if tfPlan.Monitoring != nil {
		tfPlan.Monitoring.Rum.Snippet = tfState.Monitoring.Rum.Snippet
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &tfPlan)...)
}

func (r *WebsiteResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var tfState WebsiteResourceModel

	// Read existing Terraform state.
	resp.Diagnostics.Append(req.State.Get(ctx, &tfState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the Website...
	tflog.Trace(ctx, fmt.Sprintf("deleting website - id=%s", tfState.Id))
	if err := r.client.WebsiteService().
		Delete(ctx, tfState.Id.ValueString()); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("error deleting website %s - %s", tfState.Id, err))
		return
	}
	tflog.Trace(ctx, fmt.Sprintf("website deleted: id=%s", tfState.Id))
}

func (r *WebsiteResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
