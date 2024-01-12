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
	_ resource.Resource                = &websiteResource{}
	_ resource.ResourceWithConfigure   = &websiteResource{}
	_ resource.ResourceWithImportState = &websiteResource{}
)

func NewWebsiteResource() resource.Resource {
	return &websiteResource{}
}

// Defines the resource implementation.
type websiteResource struct {
	client *swoClient.Client
}

func (r *websiteResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "website"
}

func (r *websiteResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, _ := req.ProviderData.(*swoClient.Client)
	r.client = client
}

func (r *websiteResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var tfPlan websiteResourceModel
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
				TestFromAll: tfPlan.Monitoring.Availability.PlatformOptions.TestFromAll.ValueBoolPointer(),
				ProbePlatforms: convertArray(tfPlan.Monitoring.Availability.PlatformOptions.Platforms, func(s string) swoClient.ProbePlatform {
					return swoClient.ProbePlatform(s)
				}),
			},
			TestFrom: swoClient.ProbeLocationInput{
				Type: swoClient.ProbeLocationType(tfPlan.Monitoring.Availability.TestFromLocation.ValueString()),
				Values: convertArray(tfPlan.Monitoring.Availability.LocationOptions, func(p probeLocation) string {
					return p.Value.ValueString()
				}),
			},
			Ssl: &swoClient.SslMonitoringInput{
				Enabled:                        tfPlan.Monitoring.Availability.SSL.Enabled.ValueBoolPointer(),
				DaysPriorToExpiration:          swoClient.Ptr(int(tfPlan.Monitoring.Availability.SSL.DaysPriorToExpiration.ValueInt64())),
				IgnoreIntermediateCertificates: tfPlan.Monitoring.Availability.SSL.IgnoreIntermediateCertificates.ValueBoolPointer(),
			},
			CustomHeaders: convertArray(tfPlan.Monitoring.CustomHeaders, func(h customHeader) swoClient.CustomHeaderInput {
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
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error creating website '%s' - error: %s", tfPlan.Name, err))
		return
	}

	// Get the latest Website state from the server so we can get the 'snippet' field. Ideally we need to update
	// the API to return the 'snippet' field in the create response.
	if website, err := r.client.WebsiteService().Read(ctx, result.Id); err != nil {
		resp.Diagnostics.AddWarning("Client Error",
			fmt.Sprintf("error capturing RUM snippit for Website '%s' - error: %s", tfPlan.Name, err))
	} else {
		tfPlan.Monitoring.Rum.Snippet = types.StringValue(*website.Monitoring.Rum.Snippet)
	}

	tfPlan.Id = types.StringValue(result.Id)
	resp.Diagnostics.Append(resp.State.Set(ctx, tfPlan)...)
}

func (r *websiteResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var tfState websiteResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &tfState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read the Website...
	website, err := r.client.WebsiteService().Read(ctx, tfState.Id.ValueString())

	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error reading Website %s. error: %s", tfState.Id, err))
		return
	}

	// Update the Terraform state with latest values from the server.
	tfState.Url = types.StringValue(website.Url)
	tfState.Name = types.StringPointerValue(website.Name)

	if website.Monitoring != nil {
		monitoring := website.Monitoring
		tfState.Monitoring = &websiteMonitoring{}

		if monitoring.Availability != nil {
			tfState.Monitoring.Availability = availabilityMonitoring{}
			availability := monitoring.Availability

			if availability.CheckForString != nil {
				tfState.Monitoring.Availability.CheckForString = &checkForStringType{
					Operator: types.StringValue(string(availability.CheckForString.Operator)),
					Value:    types.StringValue(availability.CheckForString.Value),
				}
			} else {
				tfState.Monitoring.Availability.CheckForString = nil
			}

			if availability.TestIntervalInSeconds != nil {
				tfState.Monitoring.Availability.TestIntervalInSeconds = types.Int64Value(int64(*availability.TestIntervalInSeconds))
			} else {
				tfState.Monitoring.Availability.TestIntervalInSeconds = types.Int64Null()
			}

			tfState.Monitoring.Availability.Protocols = convertArray(availability.Protocols, func(s swoClient.WebsiteProtocol) string {
				return string(s)
			})

			if availability.PlatformOptions != nil {
				tfState.Monitoring.Availability.PlatformOptions = platformOptions{
					TestFromAll: types.BoolValue(availability.PlatformOptions.TestFromAll),
					Platforms:   availability.PlatformOptions.Platforms,
				}
			} else {
				tfState.Monitoring.Availability.PlatformOptions = platformOptions{}
			}

			if availability.TestFromLocation != nil {
				tfState.Monitoring.Availability.TestFromLocation = types.StringValue(string(*availability.TestFromLocation))
			} else {
				tfState.Monitoring.Availability.TestFromLocation = types.StringNull()
			}

			if availability.LocationOptions != nil {
				var locOpts []probeLocation
				for _, p := range availability.LocationOptions {
					locOpts = append(locOpts, probeLocation{
						Type:  types.StringValue(string(p.Type)),
						Value: types.StringValue(string(p.Value)),
					})
				}
				tfState.Monitoring.Availability.LocationOptions = locOpts
			} else {
				tfState.Monitoring.Availability.LocationOptions = nil
			}

			if availability.Ssl != nil {
				tfState.Monitoring.Availability.SSL = &sslMonitoring{
					Enabled:                        types.BoolValue(availability.Ssl.Enabled),
					IgnoreIntermediateCertificates: types.BoolValue(availability.Ssl.IgnoreIntermediateCertificates),
				}
				if availability.Ssl.DaysPriorToExpiration != nil {
					tfState.Monitoring.Availability.SSL.DaysPriorToExpiration = types.Int64Value(int64(*availability.Ssl.DaysPriorToExpiration))
				} else {
					tfState.Monitoring.Availability.SSL.DaysPriorToExpiration = types.Int64Null()
				}
			} else {
				tfState.Monitoring.Availability.SSL = nil
			}
		}

		var customHeaders []customHeader
		if monitoring.CustomHeaders != nil {
			for _, h := range monitoring.CustomHeaders {
				customHeaders = append(customHeaders, customHeader{
					Name:  types.StringValue(h.Name),
					Value: types.StringValue(h.Value),
				})
			}
		}
		tfState.Monitoring.CustomHeaders = customHeaders

		if monitoring.Rum != nil {
			tfState.Monitoring.Rum = rumMonitoring{
				Spa: types.BoolValue(monitoring.Rum.Spa),
			}

			if monitoring.Rum.ApdexTimeInSeconds != nil {
				tfState.Monitoring.Rum.ApdexTimeInSeconds = types.Int64Value(int64(*monitoring.Rum.ApdexTimeInSeconds))
			}

			if monitoring.Rum.Snippet != nil {
				tfState.Monitoring.Rum.Snippet = types.StringValue(*monitoring.Rum.Snippet)
			}
		} else {
			tfState.Monitoring.Rum = rumMonitoring{}
		}
	} else {
		tfState.Monitoring = nil
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, tfState)...)
}

func (r *websiteResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var tfPlan, tfState *websiteResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &tfPlan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &tfState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update the Website...
	err := r.client.WebsiteService().Update(ctx,
		swoClient.UpdateWebsiteInput{
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
					Values: convertArray(tfPlan.Monitoring.Availability.LocationOptions, func(p probeLocation) string {
						return p.Value.ValueString()
					}),
				},
				Ssl: &swoClient.SslMonitoringInput{
					Enabled:                        tfPlan.Monitoring.Availability.SSL.Enabled.ValueBoolPointer(),
					DaysPriorToExpiration:          swoClient.Ptr(int(tfPlan.Monitoring.Availability.SSL.DaysPriorToExpiration.ValueInt64())),
					IgnoreIntermediateCertificates: tfPlan.Monitoring.Availability.SSL.IgnoreIntermediateCertificates.ValueBoolPointer(),
				},
				CustomHeaders: convertArray(tfPlan.Monitoring.CustomHeaders, func(h customHeader) swoClient.CustomHeaderInput {
					return swoClient.CustomHeaderInput{
						Name:  h.Name.ValueString(),
						Value: h.Value.ValueString(),
					}
				}),
			},
			Rum: &swoClient.RumMonitoringInput{
				ApdexTimeInSeconds: swoClient.Ptr(int(tfPlan.Monitoring.Rum.ApdexTimeInSeconds.ValueInt64())),
				Spa:                tfPlan.Monitoring.Rum.Spa.ValueBoolPointer(),
			},
		})

	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error updating website %s. err: %s", tfState.Id, err))
		return
	}

	// Save to Terraform state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &tfPlan)...)
}

func (r *websiteResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var tfState websiteResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &tfState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the Website...
	if err := r.client.WebsiteService().
		Delete(ctx, tfState.Id.ValueString()); err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error deleting website %s - %s", tfState.Id, err))
	}
}

func (r *websiteResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
