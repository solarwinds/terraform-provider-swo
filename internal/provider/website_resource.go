package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/cenkalti/backoff/v5"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	swoClient "github.com/solarwinds/swo-client-go/pkg/client"
	swoClientTypes "github.com/solarwinds/swo-client-go/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                     = &websiteResource{}
	_ resource.ResourceWithConfigure        = &websiteResource{}
	_ resource.ResourceWithImportState      = &websiteResource{}
	_ resource.ResourceWithConfigValidators = &websiteResource{}
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

func getAvailabilityOrRum(monitoring *websiteMonitoring) (*swoClient.AvailabilityCheckSettingsInput, *swoClient.RumMonitoringInput) {

	var availabilityCheckSettings *swoClient.AvailabilityCheckSettingsInput
	if monitoring.Availability != nil {
		var checkForString *swoClient.CheckForStringInput
		if monitoring.Availability.CheckForString != nil {
			checkForString = &swoClient.CheckForStringInput{
				Operator: swoClient.CheckStringOperator(monitoring.Availability.CheckForString.Operator.ValueString()),
				Value:    monitoring.Availability.CheckForString.Value.ValueString(),
			}
		}

		var ssl *swoClient.SslMonitoringInput
		if monitoring.Availability.SSL != nil {
			ssl = &swoClient.SslMonitoringInput{
				Enabled:                        monitoring.Availability.SSL.Enabled.ValueBoolPointer(),
				DaysPriorToExpiration:          swoClient.Ptr(int(monitoring.Availability.SSL.DaysPriorToExpiration.ValueInt64())),
				IgnoreIntermediateCertificates: monitoring.Availability.SSL.IgnoreIntermediateCertificates.ValueBoolPointer(),
			}
		}

		availabilityCheckSettings = &swoClient.AvailabilityCheckSettingsInput{
			CheckForString:        checkForString,
			TestIntervalInSeconds: swoClientTypes.TestIntervalInSeconds(int(monitoring.Availability.TestIntervalInSeconds.ValueInt64())),
			Protocols: convertArray(monitoring.Availability.Protocols, func(s string) swoClient.WebsiteProtocol {
				return swoClient.WebsiteProtocol(s)
			}),
			PlatformOptions: &swoClient.ProbePlatformOptionsInput{
				TestFromAll: monitoring.Availability.PlatformOptions.TestFromAll.ValueBoolPointer(),
				ProbePlatforms: convertArray(monitoring.Availability.PlatformOptions.Platforms, func(s string) swoClient.ProbePlatform {
					return swoClient.ProbePlatform(s)
				}),
			},
			TestFrom: swoClient.ProbeLocationInput{
				Type: swoClient.ProbeLocationType(monitoring.Availability.TestFromLocation.ValueString()),
				Values: convertArray(monitoring.Availability.LocationOptions, func(p probeLocation) string {
					return p.Value.ValueString()
				}),
			},
			Ssl: ssl,
			CustomHeaders: convertArray(monitoring.CustomHeaders, func(h customHeader) swoClient.CustomHeaderInput {
				return swoClient.CustomHeaderInput{
					Name:  h.Name.ValueString(),
					Value: h.Value.ValueString(),
				}
			})}
	}
	var rum *swoClient.RumMonitoringInput
	if monitoring.Rum != nil {
		rum = &swoClient.RumMonitoringInput{
			ApdexTimeInSeconds: swoClient.Ptr(int(monitoring.Rum.ApdexTimeInSeconds.ValueInt64())),
			Spa:                swoClient.Ptr(monitoring.Rum.Spa.ValueBool()),
		}
	}

	return availabilityCheckSettings, rum
}

func (r *websiteResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var tfPlan websiteResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &tfPlan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create our input request.
	availabilityCheckSettings, rum := getAvailabilityOrRum(tfPlan.Monitoring)

	createInput := swoClient.CreateWebsiteInput{
		Name:                      tfPlan.Name.ValueString(),
		Url:                       tfPlan.Url.ValueString(),
		AvailabilityCheckSettings: availabilityCheckSettings,
		Rum:                       rum,
	}

	// Create the Website...
	result, err := r.client.WebsiteService().Create(ctx, createInput)
	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error creating website '%s' - error: %s", tfPlan.Name, err))
		return
	}

	website, err := ReadRetry(ctx, result.Id, r.client.WebsiteService().Read)

	// Get the latest Website state from the server so we can get the 'snippet' field. Ideally we need to update
	// the API to return the 'snippet' field in the create response.
	if err != nil {
		resp.Diagnostics.AddWarning("Client Error",
			fmt.Sprintf("error capturing RUM snippit for Website '%s' - error: %s", tfPlan.Name, err))
	} else if website.Monitoring.Rum != nil && tfPlan.Monitoring.Rum != nil {
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

	website, err := ReadRetry(ctx, tfState.Id.ValueString(), r.client.WebsiteService().Read)

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

		if tfState.Monitoring.Availability != nil && monitoring.Availability != nil {
			tfState.Monitoring.Availability = &availabilityMonitoring{}
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

		if tfState.Monitoring.Rum != nil && monitoring.Rum != nil {
			tfState.Monitoring.Rum = &rumMonitoring{
				Spa: types.BoolValue(monitoring.Rum.Spa),
			}

			if monitoring.Rum.ApdexTimeInSeconds != nil {
				tfState.Monitoring.Rum.ApdexTimeInSeconds = types.Int64Value(int64(*monitoring.Rum.ApdexTimeInSeconds))
			}

			if monitoring.Rum.Snippet != nil {
				tfState.Monitoring.Rum.Snippet = types.StringValue(*monitoring.Rum.Snippet)
			}
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

	availabilityCheckSettings, rum := getAvailabilityOrRum(tfPlan.Monitoring)

	updateInput := swoClient.UpdateWebsiteInput{
		Id:                        tfState.Id.ValueString(),
		Name:                      tfPlan.Name.ValueString(),
		Url:                       tfPlan.Url.ValueString(),
		AvailabilityCheckSettings: availabilityCheckSettings,
		Rum:                       rum,
	}

	bWebsiteToMatch, err := json.Marshal(map[string]interface{}{
		"id":   updateInput.Id,
		"name": updateInput.Name,
		"url":  updateInput.Url,

		"monitoring": map[string]interface{}{
			"availability": map[string]interface{}{
				"protocols":             updateInput.AvailabilityCheckSettings.Protocols,
				"testIntervalInSeconds": updateInput.AvailabilityCheckSettings.TestIntervalInSeconds,
				"testFromLocation":      updateInput.AvailabilityCheckSettings.TestFrom.Type,
				"platformOptions": map[string]interface{}{
					"testFromAll": updateInput.AvailabilityCheckSettings.PlatformOptions.TestFromAll,
					"platforms":   updateInput.AvailabilityCheckSettings.PlatformOptions.ProbePlatforms,
				},
				"ssl": updateInput.AvailabilityCheckSettings.Ssl,
			},
			"rum":           updateInput.Rum,
			"customHeaders": updateInput.AvailabilityCheckSettings.CustomHeaders,
		},
	})

	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error marshaling website result to match %s - %s", tfState.Id, err))
		return
	}

	var websiteToMatch swoClient.ReadWebsiteResult

	err = json.Unmarshal(bWebsiteToMatch, &websiteToMatch)
	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error unmarshaling uri result to match %s - %s", tfState.Id, err))
		return
	}

	// Update the Website...
	err = r.client.WebsiteService().Update(ctx, updateInput)

	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error updating website %s. err: %s", tfState.Id, err))
		return
	}

	// Updates are eventually consistant. Retry until the Website we read and the Website we are updating match.
	_, err = BackoffRetry(func() (*swoClient.ReadWebsiteResult, error) {
		// Read the Uri...
		website, err := r.client.WebsiteService().Read(ctx, tfState.Id.ValueString())

		if err != nil {
			return nil, backoff.Permanent(err)
		}

		websiteToMatch.Typename = website.Typename
		if websiteToMatch.Monitoring.Rum != nil {
			websiteToMatch.Monitoring.Rum.Snippet = website.Monitoring.Rum.Snippet
		} else {
			websiteToMatch.Monitoring.Rum = website.Monitoring.Rum
		}
		websiteToMatch.Monitoring.Options = website.Monitoring.Options
		websiteToMatch.Monitoring.Availability.LocationOptions = website.Monitoring.Availability.LocationOptions
		websiteToMatch.Monitoring.Availability.CheckForString = website.Monitoring.Availability.CheckForString

		match := reflect.DeepEqual(&websiteToMatch, website)

		// Updated entity properties don't match, retry
		if !match {
			return nil, ErrNonMatchingEntites
		}

		return website, nil
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

func (r *websiteResource) ConfigValidators(context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.AtLeastOneOf(
			path.MatchRoot("monitoring").AtName("availability"),
			path.MatchRoot("monitoring").AtName("rum"),
		),
	}
}
