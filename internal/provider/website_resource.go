package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cenkalti/backoff/v5"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"reflect"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	swoClient "github.com/solarwinds/swo-client-go/pkg/client"
	swoClientTypes "github.com/solarwinds/swo-client-go/types"
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
	client, _ := req.ProviderData.(providerClients)
	r.client = client.SwoClient
}

func (r *websiteResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var tfPlan websiteResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &tfPlan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createInput := swoClient.CreateWebsiteInput{
		Name: tfPlan.Name.ValueString(),
		Url:  tfPlan.Url.ValueString(),
	}

	if tfPlan.Monitoring.Availability != nil {
		var checkForString *swoClient.CheckForStringInput
		if !tfPlan.Monitoring.Availability.CheckForString.IsNull() {
			var planCheckForString checkForStringType
			tfPlan.Monitoring.Availability.CheckForString.As(ctx, &planCheckForString, basetypes.ObjectAsOptions{})

			checkForString = &swoClient.CheckForStringInput{
				Operator: swoClient.CheckStringOperator(planCheckForString.Operator.ValueString()),
				Value:    planCheckForString.Value.ValueString(),
			}
		}

		var ssl *swoClient.SslMonitoringInput
		if tfPlan.Monitoring.Availability.SSL != nil && tfPlan.Monitoring.Availability.SSL.Enabled.ValueBool() {
			ssl = &swoClient.SslMonitoringInput{
				Enabled:                        tfPlan.Monitoring.Availability.SSL.Enabled.ValueBoolPointer(),
				DaysPriorToExpiration:          swoClient.Ptr(int(tfPlan.Monitoring.Availability.SSL.DaysPriorToExpiration.ValueInt64())),
				IgnoreIntermediateCertificates: tfPlan.Monitoring.Availability.SSL.IgnoreIntermediateCertificates.ValueBoolPointer(),
			}
		}

		var tfPlanCustomHeaders []customHeader

		//monitoring.custom_headers is deprecated. Both custom_headers fields cannot be set at the same time.
		if tfPlan.Monitoring.Availability.CustomHeaders != nil {
			tfPlanCustomHeaders = *tfPlan.Monitoring.Availability.CustomHeaders
		} else {
			tfPlanCustomHeaders = *tfPlan.Monitoring.CustomHeaders
		}

		var customHeaders []swoClient.CustomHeaderInput
		if len(tfPlanCustomHeaders) > 0 {
			customHeaders = convertArray(tfPlanCustomHeaders, func(h customHeader) swoClient.CustomHeaderInput {
				return swoClient.CustomHeaderInput{
					Name:  h.Name.ValueString(),
					Value: h.Value.ValueString(),
				}
			})
		}

		createInput.AvailabilityCheckSettings = &swoClient.AvailabilityCheckSettingsInput{
			CheckForString:        checkForString,
			TestIntervalInSeconds: swoClientTypes.TestIntervalInSeconds(int(tfPlan.Monitoring.Availability.TestIntervalInSeconds.ValueInt64())),
			Protocols: convertArray(tfPlan.Monitoring.Availability.Protocols.Elements(), func(s attr.Value) swoClient.WebsiteProtocol {
				return swoClient.WebsiteProtocol(attrValueToString(s))
			}),
			PlatformOptions: &swoClient.ProbePlatformOptionsInput{
				TestFromAll: tfPlan.Monitoring.Availability.PlatformOptions.TestFromAll.ValueBoolPointer(),
				ProbePlatforms: convertArray(tfPlan.Monitoring.Availability.PlatformOptions.Platforms, func(s types.String) swoClient.ProbePlatform {
					return swoClient.ProbePlatform(strings.Trim(s.String(), "\""))
				}),
			},
			TestFrom: swoClient.ProbeLocationInput{
				Type: swoClient.ProbeLocationType(tfPlan.Monitoring.Availability.TestFromLocation.ValueString()),
				Values: convertArray(tfPlan.Monitoring.Availability.LocationOptions, func(p probeLocation) string {
					return p.Value.ValueString()
				}),
			},
			Ssl:           ssl,
			CustomHeaders: customHeaders,
		}
	}

	if tfPlan.Monitoring.Rum != nil {
		createInput.Rum = &swoClient.RumMonitoringInput{
			ApdexTimeInSeconds: swoClient.Ptr(int(tfPlan.Monitoring.Rum.ApdexTimeInSeconds.ValueInt64())),
			Spa:                swoClient.Ptr(tfPlan.Monitoring.Rum.Spa.ValueBool()),
		}
	}

	// Create the Website...
	result, err := r.client.WebsiteService().Create(ctx, createInput)
	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error creating website '%s' - error: %s", tfPlan.Name, err))
		return
	}

	website, err := ReadRetry(ctx, result.Id, r.client.WebsiteService().Read)

	if err != nil {
		resp.Diagnostics.AddWarning("Client Error",
			fmt.Sprintf("error reading webiste after create '%s' - error: %s", tfPlan.Name, err))
		return
	}

	// Get the latest Website state from the server so we can get the 'snippet' field. Ideally we need to update
	// the API to return the 'snippet' field in the create response.
	// only set the snippet field if the user has RUM enabled.
	if tfPlan.Monitoring.Rum != nil {
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
			fmt.Sprintf("error reading website %s. error: %s", tfState.Name, err))
		return
	}

	tfStateCopy := tfState

	// Update the Terraform state with latest values from the server.
	tfState.Url = types.StringValue(website.Url)
	tfState.Name = types.StringPointerValue(website.Name)

	if website.Monitoring != nil {
		monitoring := website.Monitoring
		tfState.Monitoring = &websiteMonitoring{}

		availability := monitoring.Availability
		if availability != nil && website.Monitoring.Options.IsAvailabilityActive {
			tfState.Monitoring.Availability = &availabilityMonitoring{}
			if availability.CheckForString != nil {
				elementTypes := map[string]attr.Type{
					"operator": types.StringType,
					"value":    types.StringType,
				}
				elements := map[string]attr.Value{
					"operator": types.StringValue(string(availability.CheckForString.Operator)),
					"value":    types.StringValue(availability.CheckForString.Value),
				}
				checkForString, _ := types.ObjectValue(elementTypes, elements)

				tfState.Monitoring.Availability.CheckForString = checkForString
			} else {
				elementTypes := map[string]attr.Type{
					"operator": types.StringType,
					"value":    types.StringType,
				}
				checkForString := types.ObjectNull(elementTypes)
				tfState.Monitoring.Availability.CheckForString = checkForString
			}

			if availability.TestIntervalInSeconds != nil {
				tfState.Monitoring.Availability.TestIntervalInSeconds = types.Int64Value(int64(*availability.TestIntervalInSeconds))
			}

			if len(availability.Protocols) > 0 {
				tfState.Monitoring.Availability.Protocols = sliceToStringList(availability.Protocols, func(s swoClient.WebsiteProtocol) string {
					return string(s)
				})
			}

			if availability.PlatformOptions != nil {
				tfState.Monitoring.Availability.PlatformOptions = platformOptions{
					TestFromAll: types.BoolValue(availability.PlatformOptions.TestFromAll),
					Platforms: convertArray(availability.PlatformOptions.Platforms, func(p string) types.String {
						return types.StringValue(p)
					}),
				}
			}

			if availability.TestFromLocation != nil {
				tfState.Monitoring.Availability.TestFromLocation = types.StringValue(string(*availability.TestFromLocation))
			}

			if len(availability.LocationOptions) > 0 {
				var locOpts []probeLocation
				for _, p := range availability.LocationOptions {
					locOpts = append(locOpts, probeLocation{
						Type:  types.StringValue(string(p.Type)),
						Value: types.StringValue(p.Value),
					})
				}
				tfState.Monitoring.Availability.LocationOptions = locOpts
			}

			if availability.Ssl != nil && availability.Ssl.Enabled {
				tfState.Monitoring.Availability.SSL = &sslMonitoring{
					Enabled:                        types.BoolValue(availability.Ssl.Enabled),
					IgnoreIntermediateCertificates: types.BoolValue(availability.Ssl.IgnoreIntermediateCertificates),
				}
				if availability.Ssl.DaysPriorToExpiration != nil {
					tfState.Monitoring.Availability.SSL.DaysPriorToExpiration = types.Int64Value(int64(*availability.Ssl.DaysPriorToExpiration))
				} else {
					tfState.Monitoring.Availability.SSL.DaysPriorToExpiration = types.Int64Null()
				}
			}
		}

		if len(monitoring.CustomHeaders) > 0 {
			var customHeaders []customHeader
			for _, h := range monitoring.CustomHeaders {
				customHeaders = append(customHeaders, customHeader{
					Name:  types.StringValue(h.Name),
					Value: types.StringValue(h.Value),
				})
			}
			if tfStateCopy.Monitoring.CustomHeaders != nil {
				tfState.Monitoring.CustomHeaders = &customHeaders
			} else {
				tfState.Monitoring.Availability.CustomHeaders = &customHeaders
			}
		}

		if monitoring.Options.IsRumActive && monitoring.Rum != nil {
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

	updateInput := swoClient.UpdateWebsiteInput{
		Id:   tfState.Id.ValueString(),
		Name: tfPlan.Name.ValueString(),
		Url:  tfPlan.Url.ValueString(),
	}

	if tfPlan.Monitoring.Availability != nil {
		var checkForString *swoClient.CheckForStringInput
		if !tfPlan.Monitoring.Availability.CheckForString.IsNull() {
			var planCheckForString checkForStringType
			tfPlan.Monitoring.Availability.CheckForString.As(ctx, &planCheckForString, basetypes.ObjectAsOptions{})

			checkForString = &swoClient.CheckForStringInput{
				Operator: swoClient.CheckStringOperator(planCheckForString.Operator.ValueString()),
				Value:    planCheckForString.Value.ValueString(),
			}
		}
		var ssl *swoClient.SslMonitoringInput
		if tfPlan.Monitoring.Availability.SSL != nil && tfPlan.Monitoring.Availability.SSL.Enabled.ValueBool() {
			ssl = &swoClient.SslMonitoringInput{
				Enabled:                        tfPlan.Monitoring.Availability.SSL.Enabled.ValueBoolPointer(),
				DaysPriorToExpiration:          swoClient.Ptr(int(tfPlan.Monitoring.Availability.SSL.DaysPriorToExpiration.ValueInt64())),
				IgnoreIntermediateCertificates: tfPlan.Monitoring.Availability.SSL.IgnoreIntermediateCertificates.ValueBoolPointer(),
			}
		}

		var tfPlanCustomHeaders []customHeader

		//monitoring.custom_headers is deprecated. Both custom_headers fields cannot be set at the same time.
		if tfPlan.Monitoring.Availability.CustomHeaders != nil {
			tfPlanCustomHeaders = *tfPlan.Monitoring.Availability.CustomHeaders
		} else {
			tfPlanCustomHeaders = *tfPlan.Monitoring.CustomHeaders
		}

		var customHeaders []swoClient.CustomHeaderInput
		if len(tfPlanCustomHeaders) > 0 {
			customHeaders = convertArray(tfPlanCustomHeaders, func(h customHeader) swoClient.CustomHeaderInput {
				return swoClient.CustomHeaderInput{
					Name:  h.Name.ValueString(),
					Value: h.Value.ValueString(),
				}
			})
		}

		updateInput.AvailabilityCheckSettings = &swoClient.AvailabilityCheckSettingsInput{
			CheckForString:        checkForString,
			TestIntervalInSeconds: swoClientTypes.TestIntervalInSeconds(int(tfPlan.Monitoring.Availability.TestIntervalInSeconds.ValueInt64())),
			Protocols: convertArray(tfPlan.Monitoring.Availability.Protocols.Elements(), func(s attr.Value) swoClient.WebsiteProtocol {
				return swoClient.WebsiteProtocol(attrValueToString(s))
			}),
			PlatformOptions: &swoClient.ProbePlatformOptionsInput{
				TestFromAll: swoClient.Ptr(tfPlan.Monitoring.Availability.PlatformOptions.TestFromAll.ValueBool()),
				ProbePlatforms: convertArray(tfPlan.Monitoring.Availability.PlatformOptions.Platforms, func(s types.String) swoClient.ProbePlatform {
					return swoClient.ProbePlatform(attrValueToString(s))
				}),
			},
			TestFrom: swoClient.ProbeLocationInput{
				Type: swoClient.ProbeLocationType(tfPlan.Monitoring.Availability.TestFromLocation.ValueString()),
				Values: convertArray(tfPlan.Monitoring.Availability.LocationOptions, func(p probeLocation) string {
					return p.Value.ValueString()
				}),
			},
			Ssl:           ssl,
			CustomHeaders: customHeaders,
		}
	}

	if tfPlan.Monitoring.Rum != nil {
		updateInput.Rum = &swoClient.RumMonitoringInput{
			ApdexTimeInSeconds: swoClient.Ptr(int(tfPlan.Monitoring.Rum.ApdexTimeInSeconds.ValueInt64())),
			Spa:                tfPlan.Monitoring.Rum.Spa.ValueBoolPointer(),
		}
	} else {
		updateInput.Rum = nil
	}

	websiteMonitoring := map[string]interface{}{}
	if updateInput.AvailabilityCheckSettings != nil {
		websiteMonitoring["customHeaders"] = updateInput.AvailabilityCheckSettings.CustomHeaders
		websiteMonitoring["availability"] = map[string]interface{}{
			"protocols":             updateInput.AvailabilityCheckSettings.Protocols,
			"testIntervalInSeconds": updateInput.AvailabilityCheckSettings.TestIntervalInSeconds,
			"testFromLocation":      updateInput.AvailabilityCheckSettings.TestFrom.Type,
			"platformOptions": map[string]interface{}{
				"testFromAll": updateInput.AvailabilityCheckSettings.PlatformOptions.TestFromAll,
				"platforms":   updateInput.AvailabilityCheckSettings.PlatformOptions.ProbePlatforms,
			},
			"ssl": updateInput.AvailabilityCheckSettings.Ssl,
		}
	}

	bWebsiteToMatch, err := json.Marshal(map[string]interface{}{
		"id":   updateInput.Id,
		"name": updateInput.Name,
		"url":  updateInput.Url,

		"monitoring": websiteMonitoring,
		"rum":        updateInput.Rum,
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

	// Updates are eventually consistent. Retry until the Website we read and the Website we are updating match.
	website, err := BackoffRetry(func() (*swoClient.ReadWebsiteResult, error) {
		// Read the Uri...
		website, err := r.client.WebsiteService().Read(ctx, tfState.Id.ValueString())

		if err != nil {
			if errors.Is(err, swoClient.ErrEntityIdNil) {
				return website, swoClient.ErrEntityIdNil
			} else {
				return nil, backoff.Permanent(err)
			}
		}

		websiteToMatch.Typename = website.Typename
		websiteToMatch.Monitoring.Options = website.Monitoring.Options

		if websiteToMatch.Monitoring.Availability != nil {
			websiteToMatch.Monitoring.Availability.LocationOptions = website.Monitoring.Availability.LocationOptions
			websiteToMatch.Monitoring.Availability.CheckForString = website.Monitoring.Availability.CheckForString
		}

		if websiteToMatch.Monitoring.Rum == nil {
			websiteToMatch.Monitoring.Rum = website.Monitoring.Rum
		} else {
			websiteToMatch.Monitoring.Rum.Snippet = website.Monitoring.Rum.Snippet
		}

		// default values for availability are returned if availability is not set
		if tfPlan.Monitoring.Availability == nil {
			websiteToMatch.Monitoring.Availability = website.Monitoring.Availability
		}

		// default values for ssl are returned if ssl is not set
		if tfPlan.Monitoring.Availability != nil && tfPlan.Monitoring.Availability.SSL == nil {
			websiteToMatch.Monitoring.Availability.Ssl = website.Monitoring.Availability.Ssl
		}

		if websiteToMatch.Monitoring.CustomHeaders == nil && len(website.Monitoring.CustomHeaders) == 0 {
			websiteToMatch.Monitoring.CustomHeaders = website.Monitoring.CustomHeaders
		}

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

	if tfPlan.Monitoring.Rum != nil && website.Monitoring.Options.IsRumActive {
		tfPlan.Monitoring.Rum.Snippet = types.StringValue(*website.Monitoring.Rum.Snippet)
	}

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
