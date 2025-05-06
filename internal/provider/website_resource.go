package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cenkalti/backoff/v5"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	swoClient "github.com/solarwinds/swo-client-go/pkg/client"
	swoClientTypes "github.com/solarwinds/swo-client-go/types"
	"reflect"
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

	var tfMonitoring websiteMonitoring
	tfPlan.Monitoring.As(ctx, &tfMonitoring, basetypes.ObjectAsOptions{})
	if !tfMonitoring.Availability.IsNull() {
		var tfAvailability availabilityMonitoring
		tfMonitoring.Availability.As(ctx, &tfAvailability, basetypes.ObjectAsOptions{})

		var checkForString *swoClient.CheckForStringInput
		if !tfAvailability.CheckForString.IsNull() {
			var tfCheckForString checkForStringType
			tfAvailability.CheckForString.As(ctx, &tfCheckForString, basetypes.ObjectAsOptions{})

			checkForString = &swoClient.CheckForStringInput{
				Operator: swoClient.CheckStringOperator(tfCheckForString.Operator.ValueString()),
				Value:    tfCheckForString.Value.ValueString(),
			}
		}

		var ssl *swoClient.SslMonitoringInput
		if !tfAvailability.SSL.IsUnknown() {
			var tfSslMonitoring sslMonitoring
			tfAvailability.SSL.As(ctx, &tfSslMonitoring, basetypes.ObjectAsOptions{})
			if tfSslMonitoring.Enabled.ValueBool() {
				ssl = &swoClient.SslMonitoringInput{
					Enabled:                        tfSslMonitoring.Enabled.ValueBoolPointer(),
					DaysPriorToExpiration:          swoClient.Ptr(int(tfSslMonitoring.DaysPriorToExpiration.ValueInt64())),
					IgnoreIntermediateCertificates: tfSslMonitoring.IgnoreIntermediateCertificates.ValueBoolPointer(),
				}
			}
		}

		var tfCustomHeaders []customHeader
		//monitoring.custom_headers is deprecated. Both custom_headers fields cannot be set at the same time.
		if !tfAvailability.CustomHeaders.IsNull() {
			tfAvailability.CustomHeaders.ElementsAs(ctx, &tfCustomHeaders, false)
		} else {
			tfMonitoring.CustomHeaders.ElementsAs(ctx, &tfCustomHeaders, false)
		}

		var customHeaders []swoClient.CustomHeaderInput
		if len(tfCustomHeaders) > 0 {
			customHeaders = convertArray(tfCustomHeaders, func(h customHeader) swoClient.CustomHeaderInput {
				return swoClient.CustomHeaderInput{
					Name:  h.Name.ValueString(),
					Value: h.Value.ValueString(),
				}
			})
		}

		var tfLocationOptions []probeLocation
		tfAvailability.LocationOptions.ElementsAs(ctx, &tfLocationOptions, false)
		var tfPlatformOpts platformOptions
		tfAvailability.PlatformOptions.As(ctx, &tfPlatformOpts, basetypes.ObjectAsOptions{})

		createInput.AvailabilityCheckSettings = &swoClient.AvailabilityCheckSettingsInput{
			CheckForString:        checkForString,
			TestIntervalInSeconds: swoClientTypes.TestIntervalInSeconds(int(tfAvailability.TestIntervalInSeconds.ValueInt64())),
			Protocols: convertArray(tfAvailability.Protocols.Elements(), func(s attr.Value) swoClient.WebsiteProtocol {
				return swoClient.WebsiteProtocol(attrValueToString(s))
			}),
			PlatformOptions: &swoClient.ProbePlatformOptionsInput{
				TestFromAll: tfPlatformOpts.TestFromAll.ValueBoolPointer(),
				ProbePlatforms: convertArray(tfPlatformOpts.Platforms.Elements(),
					func(s attr.Value) swoClient.ProbePlatform { return swoClient.ProbePlatform(attrValueToString(s)) }),
			},
			TestFrom: swoClient.ProbeLocationInput{
				Type: swoClient.ProbeLocationType(tfAvailability.TestFromLocation.ValueString()),
				Values: convertArray(tfLocationOptions, func(p probeLocation) string {
					return p.Value.ValueString()
				}),
			},
			Ssl:           ssl,
			CustomHeaders: customHeaders,
		}
	} else {
		createInput.AvailabilityCheckSettings = nil
	}

	if !tfMonitoring.Rum.IsNull() {
		var tfRum rumMonitoring
		tfMonitoring.Rum.As(ctx, &tfRum, basetypes.ObjectAsOptions{})

		createInput.Rum = &swoClient.RumMonitoringInput{
			ApdexTimeInSeconds: swoClient.Ptr(int(tfRum.ApdexTimeInSeconds.ValueInt64())),
			Spa:                tfRum.Spa.ValueBoolPointer(),
		}
	} else {
		createInput.Rum = nil
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
			fmt.Sprintf("error reading website after create '%s' - error: %s", tfPlan.Name, err))
		return
	}

	// Get the latest Website state from the server so we can get the 'snippet' field. Ideally, we need to update
	// the API to return the 'snippet' field in the create response.
	// Only set the snippet field if the user has RUM enabled.
	if !tfMonitoring.Rum.IsNull() {
		var rum rumMonitoring
		tfMonitoring.Rum.As(ctx, &rum, basetypes.ObjectAsOptions{})

		rum.Snippet = types.StringValue(*website.Monitoring.Rum.Snippet)

		rumObject, _ := types.ObjectValueFrom(ctx, RumMonitoringAttributeTypes(), rum)
		tfMonitoring.Rum = rumObject

		monitoringObject, _ := types.ObjectValueFrom(ctx, WebsiteMonitoringAttributeTypes(), tfMonitoring)
		tfPlan.Monitoring = monitoringObject
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

	// Update the Terraform state with the latest values from the server.
	tfState.Url = types.StringValue(website.Url)
	tfState.Name = types.StringPointerValue(website.Name)
	tfState.Monitoring = types.ObjectNull(WebsiteMonitoringAttributeTypes())

	if website.Monitoring != nil {
		monitoring := website.Monitoring
		tfStateMonitoring := websiteMonitoring{
			Options:       types.ObjectNull(MonitoringOptionsAttributeTypes()),
			Availability:  types.ObjectNull(AvailabilityMonitoringAttributeTypes()),
			Rum:           types.ObjectNull(RumMonitoringAttributeTypes()),
			CustomHeaders: types.SetNull(types.ObjectType{AttrTypes: CustomHeaderAttributeTypes()}),
		}

		availability := monitoring.Availability
		if availability != nil && website.Monitoring.Options.IsAvailabilityActive {
			tfStateAvailability := availabilityMonitoring{
				CheckForString:        types.ObjectNull(CheckForStringTypeAttributeTypes()),
				SSL:                   types.ObjectNull(SslMonitoringAttributeTypes()),
				Protocols:             types.ListNull(types.StringType),
				TestFromLocation:      types.StringNull(),
				TestIntervalInSeconds: types.Int64Null(),
				LocationOptions:       types.SetUnknown(types.ObjectType{AttrTypes: ProbeLocationAttributeTypes()}),
				PlatformOptions:       types.ObjectNull(PlatformOptionsAttributeTypes()),
				CustomHeaders:         types.SetNull(types.ObjectType{AttrTypes: CustomHeaderAttributeTypes()}),
			}

			if availability.CheckForString != nil {
				elements := checkForStringType{
					Operator: types.StringValue(string(availability.CheckForString.Operator)),
					Value:    types.StringValue(availability.CheckForString.Value),
				}
				checkForString, _ := types.ObjectValueFrom(ctx, CheckForStringTypeAttributeTypes(), elements)
				tfStateAvailability.CheckForString = checkForString
			}

			if availability.TestIntervalInSeconds != nil {
				tfStateAvailability.TestIntervalInSeconds = types.Int64Value(int64(*availability.TestIntervalInSeconds))
			}

			if len(availability.Protocols) > 0 {
				tfStateAvailability.Protocols = sliceToStringList(availability.Protocols, func(s swoClient.WebsiteProtocol) string {
					return string(s)
				})
			}

			if availability.PlatformOptions != nil {
				platforms := convertArray(availability.PlatformOptions.Platforms, func(p string) types.String {
					return types.StringValue(p)
				})
				platformValue, _ := types.SetValueFrom(ctx, types.StringType, platforms)

				platformOptionsValue := platformOptions{
					TestFromAll: types.BoolValue(availability.PlatformOptions.TestFromAll),
					Platforms:   platformValue,
				}
				tfPlatformOptions, _ := types.ObjectValueFrom(ctx, PlatformOptionsAttributeTypes(), platformOptionsValue)
				tfStateAvailability.PlatformOptions = tfPlatformOptions
			}

			if availability.TestFromLocation != nil {
				tfStateAvailability.TestFromLocation = types.StringValue(string(*availability.TestFromLocation))
			}

			var locOpts []attr.Value
			if len(availability.LocationOptions) > 0 {
				for _, p := range availability.LocationOptions {
					objectValue, _ := types.ObjectValueFrom(
						ctx,
						ProbeLocationAttributeTypes(),
						probeLocation{
							Type:  types.StringValue(string(p.Type)),
							Value: types.StringValue(p.Value),
						},
					)
					locOpts = append(locOpts, objectValue)
				}

				tfLocationOptions, _ := types.SetValueFrom(ctx, types.ObjectType{AttrTypes: ProbeLocationAttributeTypes()}, locOpts)
				tfStateAvailability.LocationOptions = tfLocationOptions
			}

			sslTypes := SslMonitoringAttributeTypes()
			if availability.Ssl != nil && availability.Ssl.Enabled {
				sslValues := sslMonitoring{
					Enabled:                        types.BoolValue(availability.Ssl.Enabled),
					IgnoreIntermediateCertificates: types.BoolValue(availability.Ssl.IgnoreIntermediateCertificates),
					DaysPriorToExpiration:          types.Int64Null(),
				}
				if availability.Ssl.DaysPriorToExpiration != nil {
					sslValues.DaysPriorToExpiration = types.Int64Value(int64(*availability.Ssl.DaysPriorToExpiration))
				}
				objectValue, _ := types.ObjectValueFrom(ctx, sslTypes, sslValues)
				tfStateAvailability.SSL = objectValue
			} else {
				nullValue := types.ObjectNull(sslTypes)
				tfStateAvailability.SSL = nullValue
			}

			availabilityValue, _ := types.ObjectValueFrom(ctx, AvailabilityMonitoringAttributeTypes(), tfStateAvailability)
			tfStateMonitoring.Availability = availabilityValue
		}

		customHeaderElementTypes := CustomHeaderAttributeTypes()
		nullCustomHeader := types.SetNull(types.ObjectType{AttrTypes: customHeaderElementTypes})
		if len(monitoring.CustomHeaders) > 0 {
			var diags diag.Diagnostics
			var elements []attr.Value
			for _, h := range monitoring.CustomHeaders {
				objectValue, objectDiags := types.ObjectValueFrom(
					ctx,
					customHeaderElementTypes,
					customHeader{
						Name:  types.StringValue(h.Name),
						Value: types.StringValue(h.Value),
					},
				)
				elements = append(elements, objectValue)
				diags = append(diags, objectDiags...)
			}
			customHeaderValue, setDiags := types.SetValueFrom(ctx, types.ObjectType{AttrTypes: customHeaderElementTypes}, elements)
			diags = append(diags, setDiags...)

			var m availabilityMonitoring
			tfStateMonitoring.Availability.As(ctx, &m, basetypes.ObjectAsOptions{})
			if !m.CustomHeaders.IsNull() {
				m.CustomHeaders = customHeaderValue
				tfStateMonitoring.CustomHeaders = nullCustomHeader

				availabilityValue, _ := types.ObjectValueFrom(ctx, AvailabilityMonitoringAttributeTypes(), m)
				tfStateMonitoring.Availability = availabilityValue
			} else {
				m.CustomHeaders = nullCustomHeader
				tfStateMonitoring.CustomHeaders = customHeaderValue

				availabilityValue, _ := types.ObjectValueFrom(ctx, AvailabilityMonitoringAttributeTypes(), m)
				tfStateMonitoring.Availability = availabilityValue
			}
		}

		if monitoring.Options.IsRumActive && monitoring.Rum != nil {
			rumAttributeTypes := RumMonitoringAttributeTypes()
			rumValue := rumMonitoring{
				Spa:                types.BoolValue(monitoring.Rum.Spa),
				ApdexTimeInSeconds: types.Int64Null(),
				Snippet:            types.StringNull(),
			}

			if monitoring.Rum.ApdexTimeInSeconds != nil {
				rumValue.ApdexTimeInSeconds = types.Int64Value(int64(*monitoring.Rum.ApdexTimeInSeconds))
			}

			if monitoring.Rum.Snippet != nil {
				rumValue.Snippet = types.StringValue(*monitoring.Rum.Snippet)
			}

			rum, _ := types.ObjectValueFrom(ctx, rumAttributeTypes, rumValue)
			tfStateMonitoring.Rum = rum
		} else {
			nullValues := types.ObjectNull(RumMonitoringAttributeTypes())
			tfStateMonitoring.Rum = nullValues
		}

		tfState2, _ := types.ObjectValueFrom(ctx, WebsiteMonitoringAttributeTypes(), tfStateMonitoring)
		tfState.Monitoring = tfState2
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

	var tfMonitoring websiteMonitoring
	tfPlan.Monitoring.As(ctx, &tfMonitoring, basetypes.ObjectAsOptions{})

	if !tfMonitoring.Availability.IsNull() {
		var tfAvailability availabilityMonitoring
		tfMonitoring.Availability.As(ctx, &tfAvailability, basetypes.ObjectAsOptions{})

		var checkForString *swoClient.CheckForStringInput
		if !tfAvailability.CheckForString.IsNull() {
			var tfCheckForString checkForStringType
			tfAvailability.CheckForString.As(ctx, &tfCheckForString, basetypes.ObjectAsOptions{})

			checkForString = &swoClient.CheckForStringInput{
				Operator: swoClient.CheckStringOperator(tfCheckForString.Operator.ValueString()),
				Value:    tfCheckForString.Value.ValueString(),
			}
		}
		var ssl *swoClient.SslMonitoringInput
		if !tfAvailability.SSL.IsUnknown() {
			var tfSslMonitoring sslMonitoring
			tfAvailability.SSL.As(ctx, &tfSslMonitoring, basetypes.ObjectAsOptions{})
			if tfSslMonitoring.Enabled.ValueBool() {
				ssl = &swoClient.SslMonitoringInput{
					Enabled:                        tfSslMonitoring.Enabled.ValueBoolPointer(),
					DaysPriorToExpiration:          swoClient.Ptr(int(tfSslMonitoring.DaysPriorToExpiration.ValueInt64())),
					IgnoreIntermediateCertificates: tfSslMonitoring.IgnoreIntermediateCertificates.ValueBoolPointer(),
				}
			}
		}

		var tfCustomHeaders []customHeader
		//monitoring.custom_headers is deprecated. Both custom_headers fields cannot be set at the same time.
		if !tfAvailability.CustomHeaders.IsNull() {
			tfAvailability.CustomHeaders.ElementsAs(ctx, &tfCustomHeaders, false)
		} else {
			tfMonitoring.CustomHeaders.ElementsAs(ctx, &tfCustomHeaders, false)
		}

		var customHeaders []swoClient.CustomHeaderInput
		if len(tfCustomHeaders) > 0 {
			customHeaders = convertArray(tfCustomHeaders, func(h customHeader) swoClient.CustomHeaderInput {
				return swoClient.CustomHeaderInput{
					Name:  h.Name.ValueString(),
					Value: h.Value.ValueString(),
				}
			})
		}

		var tfLocationOptions []probeLocation
		tfAvailability.LocationOptions.ElementsAs(ctx, &tfLocationOptions, false)
		var tfPlatformOpts platformOptions
		tfAvailability.PlatformOptions.As(ctx, &tfPlatformOpts, basetypes.ObjectAsOptions{})

		updateInput.AvailabilityCheckSettings = &swoClient.AvailabilityCheckSettingsInput{
			CheckForString:        checkForString,
			TestIntervalInSeconds: swoClientTypes.TestIntervalInSeconds(int(tfAvailability.TestIntervalInSeconds.ValueInt64())),
			Protocols: convertArray(tfAvailability.Protocols.Elements(), func(s attr.Value) swoClient.WebsiteProtocol {
				return swoClient.WebsiteProtocol(attrValueToString(s))
			}),
			PlatformOptions: &swoClient.ProbePlatformOptionsInput{
				TestFromAll: tfPlatformOpts.TestFromAll.ValueBoolPointer(),
				ProbePlatforms: convertArray(tfPlatformOpts.Platforms.Elements(),
					func(s attr.Value) swoClient.ProbePlatform { return swoClient.ProbePlatform(attrValueToString(s)) }),
			},
			TestFrom: swoClient.ProbeLocationInput{
				Type: swoClient.ProbeLocationType(tfAvailability.TestFromLocation.ValueString()),
				Values: convertArray(tfLocationOptions, func(p probeLocation) string {
					return p.Value.ValueString()
				}),
			},
			Ssl:           ssl,
			CustomHeaders: customHeaders,
		}
	} else {
		updateInput.AvailabilityCheckSettings = nil
	}

	if !tfMonitoring.Rum.IsNull() {
		var tfRum rumMonitoring
		tfMonitoring.Rum.As(ctx, &tfRum, basetypes.ObjectAsOptions{})

		updateInput.Rum = &swoClient.RumMonitoringInput{
			ApdexTimeInSeconds: swoClient.Ptr(int(tfRum.ApdexTimeInSeconds.ValueInt64())),
			Spa:                tfRum.Spa.ValueBoolPointer(),
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
		if tfMonitoring.Availability.IsNull() {
			websiteToMatch.Monitoring.Availability = website.Monitoring.Availability
		}

		// default values for ssl are returned if ssl is not set
		if !tfMonitoring.Availability.IsNull() {
			var tfAvailability availabilityMonitoring
			tfMonitoring.Availability.As(ctx, &tfAvailability, basetypes.ObjectAsOptions{})
			if tfAvailability.SSL.IsNull() {
				websiteToMatch.Monitoring.Availability.Ssl = website.Monitoring.Availability.Ssl
			}
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

	if !tfMonitoring.Rum.IsNull() && website.Monitoring.Options.IsRumActive {
		var tfRum rumMonitoring
		tfMonitoring.Rum.As(ctx, &tfRum, basetypes.ObjectAsOptions{})

		tfRum.Snippet = types.StringValue(*website.Monitoring.Rum.Snippet)

		rumValue, _ := types.ObjectValueFrom(ctx, RumMonitoringAttributeTypes(), tfRum)
		tfMonitoring.Rum = rumValue

		monitoringValue, _ := types.ObjectValueFrom(ctx, WebsiteMonitoringAttributeTypes(), tfMonitoring)
		tfPlan.Monitoring = monitoringValue
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
