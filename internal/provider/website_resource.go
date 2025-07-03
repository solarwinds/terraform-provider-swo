package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/cenkalti/backoff/v5"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/solarwinds/swo-sdk-go/swov1"
	"github.com/solarwinds/swo-sdk-go/swov1/models/components"
	"github.com/solarwinds/swo-sdk-go/swov1/models/operations"
)

const (
	websiteExpBackoffMaxInterval = 30 * time.Second
	websiteExpBackoffMaxElapsed  = 2 * time.Minute
)

var (
	operatorMap = map[string]components.Operator{
		string(components.OperatorContains):       components.OperatorContains,
		string(components.OperatorDoesNotContain): components.OperatorDoesNotContain,
	}

	websiteProtocolMap = map[string]components.WebsiteProtocol{
		string(components.WebsiteProtocolHTTP):  components.WebsiteProtocolHTTP,
		string(components.WebsiteProtocolHTTPS): components.WebsiteProtocolHTTPS,
	}

	probePlatformMap = map[string]components.ProbePlatform{
		string(components.ProbePlatformAws):         components.ProbePlatformAws,
		string(components.ProbePlatformAzure):       components.ProbePlatformAzure,
		string(components.ProbePlatformGoogleCloud): components.ProbePlatformGoogleCloud,
	}

	testFromTypeMap = map[string]components.Type{
		string(components.TypeRegion):  components.TypeRegion,
		string(components.TypeCountry): components.TypeCountry,
		string(components.TypeCity):    components.TypeCity,
	}
)

var (
	_ resource.Resource                = &websiteResource{}
	_ resource.ResourceWithConfigure   = &websiteResource{}
	_ resource.ResourceWithImportState = &websiteResource{}
)

func NewWebsiteResource() resource.Resource {
	return &websiteResource{}
}

type websiteResource struct {
	client *swov1.Swo
}

func (r *websiteResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "website"
}

func (r *websiteResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, _ := req.ProviderData.(providerClients)
	r.client = client.SwoV1Client
}

func (r *websiteResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var tfPlan websiteResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &tfPlan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createInput := components.Website{
		Name: tfPlan.Name.ValueString(),
		URL:  tfPlan.Url.ValueString(),
	}

	var tfMonitoring websiteMonitoring
	d := tfPlan.Monitoring.As(ctx, &tfMonitoring, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !tfMonitoring.Availability.IsNull() {
		var tfAvailability availabilityMonitoring
		d = tfMonitoring.Availability.As(ctx, &tfAvailability, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}

		availabilitySettings, err := mapAvailabilitySettings(ctx, tfAvailability, tfMonitoring)
		if err != nil {
			resp.Diagnostics.AddError("Invalid Availability Configuration", err.Error())
			return
		}

		createInput.AvailabilityCheckSettings = availabilitySettings
	}

	if !tfMonitoring.Rum.IsNull() {
		var tfRum rumMonitoring
		d = tfMonitoring.Rum.As(ctx, &tfRum, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}

		createInput.Rum = mapRumSettings(ctx, tfRum)
	}

	result, err := r.client.Dem.CreateWebsite(ctx, createInput)
	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error creating website '%s' - error: %s", tfPlan.Name, err))
		return
	}

	if result.EntityID == nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error creating website '%s' - no entity ID returned", tfPlan.Name))
		return
	}

	tfPlan.Id = types.StringValue(result.EntityID.GetID())

	userMonitoringOptions := monitoringOptions{
		IsAvailabilityActive: types.BoolValue(!tfMonitoring.Availability.IsNull()),
		IsRumActive:          types.BoolValue(!tfMonitoring.Rum.IsNull()),
	}

	userOptions, dOpts := types.ObjectValueFrom(ctx, MonitoringOptionsAttributeTypes(), userMonitoringOptions)
	resp.Diagnostics.Append(dOpts...)
	if resp.Diagnostics.HasError() {
		return
	}
	tfMonitoring.Options = userOptions

	websiteResp, err := r.client.Dem.GetWebsite(ctx, operations.GetWebsiteRequest{
		EntityID: tfPlan.Id.ValueString(),
	})

	if !tfMonitoring.Rum.IsNull() {
		var rum rumMonitoring
		d = tfMonitoring.Rum.As(ctx, &rum, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}

		if err == nil && websiteResp.GetWebsiteResponse != nil && websiteResp.GetWebsiteResponse.Rum != nil && websiteResp.GetWebsiteResponse.Rum.Snippet != nil {
			rum.Snippet = types.StringValue(*websiteResp.GetWebsiteResponse.Rum.Snippet)
		} else {
			rum.Snippet = types.StringValue("")
		}

		rumObject, dRum := types.ObjectValueFrom(ctx, RumMonitoringAttributeTypes(), rum)
		resp.Diagnostics.Append(dRum...)
		if resp.Diagnostics.HasError() {
			return
		}
		tfMonitoring.Rum = rumObject
	}

	monitoringObject, dMonitor := types.ObjectValueFrom(ctx, WebsiteMonitoringAttributeTypes(), tfMonitoring)
	resp.Diagnostics.Append(dMonitor...)
	if resp.Diagnostics.HasError() {
		return
	}
	tfPlan.Monitoring = monitoringObject

	if err != nil {
		resp.Diagnostics.AddWarning("Client Error",
			fmt.Sprintf("error reading website after create '%s' - error: %s", tfPlan.Name, err))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, tfPlan)...)
}

func (r *websiteResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var tfState websiteResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &tfState)...)

	if resp.Diagnostics.HasError() {
		return
	}

	readOperation := func(ctx context.Context, id string) (*components.GetWebsiteResponse, error) {
		websiteResp, err := r.client.Dem.GetWebsite(ctx, operations.GetWebsiteRequest{
			EntityID: id,
		})

		if err != nil {
			return nil, err
		}

		if websiteResp.GetWebsiteResponse == nil {
			return nil, fmt.Errorf("no website data returned")
		}

		return websiteResp.GetWebsiteResponse, nil
	}

	website, err := websiteReadRetry(ctx, tfState.Id.ValueString(), readOperation)

	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error reading website %s. error: %s", tfState.Name, err))
		return
	}

	tfState.Url = types.StringValue(website.URL)
	tfState.Name = types.StringValue(website.Name)

	if website != nil {
		var existingMonitoring websiteMonitoring

		if !tfState.Monitoring.IsNull() {
			d := tfState.Monitoring.As(ctx, &existingMonitoring, basetypes.ObjectAsOptions{})
			resp.Diagnostics.Append(d...)
			if resp.Diagnostics.HasError() {
				return
			}
		} else {
			existingMonitoring = websiteMonitoring{
				Options:       types.ObjectNull(MonitoringOptionsAttributeTypes()),
				Availability:  types.ObjectNull(AvailabilityMonitoringAttributeTypes()),
				Rum:           types.ObjectNull(RumMonitoringAttributeTypes()),
				CustomHeaders: types.SetNull(types.ObjectType{AttrTypes: CustomHeaderAttributeTypes()}),
			}
		}

		monitoringOpts := website.MonitoringOptions
		serverMonitoringOptions := monitoringOptions{
			IsAvailabilityActive: types.BoolValue(monitoringOpts.IsAvailabilityActive),
			IsRumActive:          types.BoolValue(monitoringOpts.IsRumActive),
		}

		serverOptions, d := types.ObjectValueFrom(ctx, MonitoringOptionsAttributeTypes(), serverMonitoringOptions)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}

		tfStateMonitoring := websiteMonitoring{
			Options:       serverOptions,
			Availability:  types.ObjectNull(AvailabilityMonitoringAttributeTypes()),
			Rum:           types.ObjectNull(RumMonitoringAttributeTypes()),
			CustomHeaders: types.SetNull(types.ObjectType{AttrTypes: CustomHeaderAttributeTypes()}),
		}

		if website.AvailabilityCheckSettings != nil && serverMonitoringOptions.IsAvailabilityActive.ValueBool() {
			availability := website.AvailabilityCheckSettings
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
				checkForString, d := types.ObjectValueFrom(ctx, CheckForStringTypeAttributeTypes(), elements)
				resp.Diagnostics.Append(d...)
				if resp.Diagnostics.HasError() {
					return
				}
				tfStateAvailability.CheckForString = checkForString
			}

			if availability.TestIntervalInSeconds != 0 {
				tfStateAvailability.TestIntervalInSeconds = types.Int64Value(int64(availability.TestIntervalInSeconds))
			}

			if len(availability.Protocols) > 0 {
				tfStateAvailability.Protocols = sliceToStringList(availability.Protocols, func(s components.WebsiteProtocol) string {
					return string(s)
				})
			}

			if availability.PlatformOptions != nil {

				var platforms []types.String
				if availability.PlatformOptions.ProbePlatforms != nil {
					for _, p := range availability.PlatformOptions.ProbePlatforms {
						platforms = append(platforms, types.StringValue(string(p)))
					}
				}
				platformValue, d := types.SetValueFrom(ctx, types.StringType, platforms)
				resp.Diagnostics.Append(d...)
				if resp.Diagnostics.HasError() {
					return
				}

				testFromAll := false
				if availability.PlatformOptions.TestFromAll != nil {
					testFromAll = *availability.PlatformOptions.TestFromAll
				}

				platformOptionsValue := platformOptions{
					TestFromAll: types.BoolValue(testFromAll),
					Platforms:   platformValue,
				}
				tfPlatformOptions, d := types.ObjectValueFrom(ctx, PlatformOptionsAttributeTypes(), platformOptionsValue)
				resp.Diagnostics.Append(d...)
				if resp.Diagnostics.HasError() {
					return
				}
				tfStateAvailability.PlatformOptions = tfPlatformOptions
			}

			if availability.TestFrom.Type != "" {
				tfStateAvailability.TestFromLocation = types.StringValue(string(availability.TestFrom.Type))

				var locOpts []attr.Value
				if len(availability.TestFrom.Values) > 0 {
					for _, value := range availability.TestFrom.Values {
						objectValue, d := types.ObjectValueFrom(
							ctx,
							ProbeLocationAttributeTypes(),
							probeLocation{
								Type:  types.StringValue(string(components.TypeRegion)),
								Value: types.StringValue(value),
							},
						)

						resp.Diagnostics.Append(d...)
						if resp.Diagnostics.HasError() {
							return
						}
						locOpts = append(locOpts, objectValue)
					}

					tfLocationOptions, d := types.SetValueFrom(ctx, types.ObjectType{AttrTypes: ProbeLocationAttributeTypes()}, locOpts)
					resp.Diagnostics.Append(d...)
					if resp.Diagnostics.HasError() {
						return
					}
					tfStateAvailability.LocationOptions = tfLocationOptions
				}
			}

			sslTypes := SslMonitoringAttributeTypes()
			if availability.Ssl != nil && availability.Ssl.Enabled != nil && *availability.Ssl.Enabled {
				sslValues := sslMonitoring{
					Enabled:                        types.BoolValue(*availability.Ssl.Enabled),
					IgnoreIntermediateCertificates: types.BoolValue(availability.Ssl.IgnoreIntermediateCertificates != nil && *availability.Ssl.IgnoreIntermediateCertificates),
					DaysPriorToExpiration:          types.Int64Null(),
				}
				if availability.Ssl.DaysPriorToExpiration != nil {
					sslValues.DaysPriorToExpiration = types.Int64Value(int64(*availability.Ssl.DaysPriorToExpiration))
				}
				objectValue, d := types.ObjectValueFrom(ctx, sslTypes, sslValues)
				resp.Diagnostics.Append(d...)
				if resp.Diagnostics.HasError() {
					return
				}
				tfStateAvailability.SSL = objectValue
			} else {
				nullValue := types.ObjectNull(sslTypes)
				tfStateAvailability.SSL = nullValue
			}

			availabilityValue, d := types.ObjectValueFrom(ctx, AvailabilityMonitoringAttributeTypes(), tfStateAvailability)
			resp.Diagnostics.Append(d...)
			if resp.Diagnostics.HasError() {
				return
			}
			tfStateMonitoring.Availability = availabilityValue
		}

		customHeaderElementTypes := CustomHeaderAttributeTypes()
		nullCustomHeader := types.SetNull(types.ObjectType{AttrTypes: customHeaderElementTypes})

		if website.AvailabilityCheckSettings != nil && len(website.AvailabilityCheckSettings.CustomHeaders) > 0 {
			var elements []attr.Value
			for _, h := range website.AvailabilityCheckSettings.CustomHeaders {
				objectValue, d := types.ObjectValueFrom(
					ctx,
					customHeaderElementTypes,
					customHeader{
						Name:  types.StringValue(h.Name),
						Value: types.StringValue(h.Value),
					},
				)

				resp.Diagnostics.Append(d...)
				if resp.Diagnostics.HasError() {
					return
				}
				elements = append(elements, objectValue)
			}
			customHeaderValue, d := types.SetValueFrom(ctx, types.ObjectType{AttrTypes: customHeaderElementTypes}, elements)
			resp.Diagnostics.Append(d...)
			if resp.Diagnostics.HasError() {
				return
			}

			var m availabilityMonitoring
			d = tfStateMonitoring.Availability.As(ctx, &m, basetypes.ObjectAsOptions{})
			resp.Diagnostics.Append(d...)
			if resp.Diagnostics.HasError() {
				return
			}

			if !m.CustomHeaders.IsNull() {
				m.CustomHeaders = customHeaderValue
				tfStateMonitoring.CustomHeaders = nullCustomHeader

				availabilityValue, dAvail := types.ObjectValueFrom(ctx, AvailabilityMonitoringAttributeTypes(), m)
				resp.Diagnostics.Append(dAvail...)
				if resp.Diagnostics.HasError() {
					return
				}
				tfStateMonitoring.Availability = availabilityValue
			} else {
				m.CustomHeaders = nullCustomHeader
				tfStateMonitoring.CustomHeaders = customHeaderValue

				availabilityValue, dAvail := types.ObjectValueFrom(ctx, AvailabilityMonitoringAttributeTypes(), m)
				resp.Diagnostics.Append(dAvail...)
				if resp.Diagnostics.HasError() {
					return
				}
				tfStateMonitoring.Availability = availabilityValue
			}
		}

		if website.Rum != nil && serverMonitoringOptions.IsRumActive.ValueBool() {
			rumAttributeTypes := RumMonitoringAttributeTypes()
			rumValue := rumMonitoring{
				Spa:                types.BoolValue(website.Rum.Spa),
				ApdexTimeInSeconds: types.Int64Null(),
				Snippet:            types.StringNull(),
			}

			if website.Rum.ApdexTimeInSeconds != nil {
				rumValue.ApdexTimeInSeconds = types.Int64Value(int64(*website.Rum.ApdexTimeInSeconds))
			}

			if website.Rum.Snippet != nil {
				rumValue.Snippet = types.StringValue(*website.Rum.Snippet)
			}

			rum, d := types.ObjectValueFrom(ctx, rumAttributeTypes, rumValue)
			resp.Diagnostics.Append(d...)
			if resp.Diagnostics.HasError() {
				return
			}
			tfStateMonitoring.Rum = rum
		} else {
			nullValues := types.ObjectNull(RumMonitoringAttributeTypes())
			tfStateMonitoring.Rum = nullValues
		}

		tfState2, d := types.ObjectValueFrom(ctx, WebsiteMonitoringAttributeTypes(), tfStateMonitoring)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}
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

	updateInput := components.Website{
		Name: tfPlan.Name.ValueString(),
		URL:  tfPlan.Url.ValueString(),
	}

	var tfMonitoring websiteMonitoring
	d := tfPlan.Monitoring.As(ctx, &tfMonitoring, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !tfMonitoring.Availability.IsNull() {
		var tfAvailability availabilityMonitoring
		d = tfMonitoring.Availability.As(ctx, &tfAvailability, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}

		availabilitySettings, err := mapAvailabilitySettings(ctx, tfAvailability, tfMonitoring)
		if err != nil {
			resp.Diagnostics.AddError("Invalid Availability Configuration", err.Error())
			return
		}

		updateInput.AvailabilityCheckSettings = availabilitySettings
	}

	if !tfMonitoring.Rum.IsNull() {
		var tfRum rumMonitoring
		d = tfMonitoring.Rum.As(ctx, &tfRum, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}

		updateInput.Rum = mapRumSettings(ctx, tfRum)
	}

	_, err := r.client.Dem.UpdateWebsite(ctx, operations.UpdateWebsiteRequest{
		EntityID: tfState.Id.ValueString(),
		Website:  updateInput,
	})

	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error updating website %s. err: %s", tfState.Id.ValueString(), err))
		return
	}

	readOperation := func(ctx context.Context, id string) (*components.GetWebsiteResponse, error) {
		websiteResp, err := r.client.Dem.GetWebsite(ctx, operations.GetWebsiteRequest{
			EntityID: id,
		})

		if err != nil {
			return nil, err
		}

		if websiteResp.GetWebsiteResponse == nil {
			return nil, fmt.Errorf("no website data returned")
		}

		expectedName := tfPlan.Name.ValueString()
		if websiteResp.GetWebsiteResponse.Name != expectedName {
			return nil, fmt.Errorf("website name not yet updated, expected '%s' but got '%s'", expectedName, websiteResp.GetWebsiteResponse.Name)
		}

		return websiteResp.GetWebsiteResponse, nil
	}

	website, err := websiteReadRetry(ctx, tfState.Id.ValueString(), readOperation)

	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error reading website after update %s. error: %s", tfState.Id.ValueString(), err))
		return
	}

	tfPlan.Name = types.StringValue(website.Name)
	tfPlan.Url = types.StringValue(website.URL)

	userMonitoringOptions := monitoringOptions{
		IsAvailabilityActive: types.BoolValue(!tfMonitoring.Availability.IsNull()),
		IsRumActive:          types.BoolValue(!tfMonitoring.Rum.IsNull()),
	}

	userOptions, dOpts := types.ObjectValueFrom(ctx, MonitoringOptionsAttributeTypes(), userMonitoringOptions)
	resp.Diagnostics.Append(dOpts...)
	if resp.Diagnostics.HasError() {
		return
	}
	tfMonitoring.Options = userOptions

	if !tfMonitoring.Rum.IsNull() && website.Rum != nil {
		var tfRum rumMonitoring
		d = tfMonitoring.Rum.As(ctx, &tfRum, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}

		if website.Rum.Snippet != nil {
			tfRum.Snippet = types.StringValue(*website.Rum.Snippet)
		}

		rumValue, d := types.ObjectValueFrom(ctx, RumMonitoringAttributeTypes(), tfRum)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}
		tfMonitoring.Rum = rumValue
	}

	monitoringValue, d := types.ObjectValueFrom(ctx, WebsiteMonitoringAttributeTypes(), tfMonitoring)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	tfPlan.Monitoring = monitoringValue

	resp.Diagnostics.Append(resp.State.Set(ctx, &tfPlan)...)
}

func (r *websiteResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var tfState websiteResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &tfState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Dem.DeleteWebsite(ctx, operations.DeleteWebsiteRequest{
		EntityID: tfState.Id.ValueString(),
	})

	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error deleting website %s - %s", tfState.Id.ValueString(), err))
	}
}

func (r *websiteResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func stringToOperator(operatorStr string) (components.Operator, error) {
	if operator, exists := operatorMap[operatorStr]; exists {
		return operator, nil
	}
	return "", fmt.Errorf("unsupported operator '%s', valid values: %v", operatorStr, getMapKeys(operatorMap))
}

func stringToWebsiteProtocol(protocolStr string) (components.WebsiteProtocol, error) {
	if protocol, exists := websiteProtocolMap[protocolStr]; exists {
		return protocol, nil
	}
	return "", fmt.Errorf("unsupported protocol '%s', valid values: %v", protocolStr, getMapKeys(websiteProtocolMap))
}

func stringToProbePlatform(platformStr string) (components.ProbePlatform, error) {
	if platform, exists := probePlatformMap[platformStr]; exists {
		return platform, nil
	}
	return "", fmt.Errorf("unsupported platform '%s', valid values: %v", platformStr, getMapKeys(probePlatformMap))
}

func stringToTestFromType(typeStr string) (components.Type, error) {
	if testFromType, exists := testFromTypeMap[typeStr]; exists {
		return testFromType, nil
	}
	return "", fmt.Errorf("unsupported test from type '%s', valid values: %v", typeStr, getMapKeys(testFromTypeMap))
}

func websiteReadRetry(ctx context.Context, id string, operation func(context.Context, string) (*components.GetWebsiteResponse, error)) (*components.GetWebsiteResponse, error) {
	var website *components.GetWebsiteResponse

	expBackoff := backoff.NewExponentialBackOff()
	expBackoff.MaxInterval = websiteExpBackoffMaxInterval

	website, err := backoff.Retry(ctx, func() (*components.GetWebsiteResponse, error) {
		result, err := operation(ctx, id)
		if err != nil {
			return nil, err
		}

		if result == nil {
			return nil, fmt.Errorf("no website entity returned for id %s", id)
		}

		if result.ID == "" {
			return nil, fmt.Errorf("website entity %s exists but has no data", id)
		}

		return result, nil
	}, backoff.WithBackOff(expBackoff), backoff.WithMaxElapsedTime(websiteExpBackoffMaxElapsed))

	return website, err
}

func getMapKeys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func mapProtocolsFromTerraform(tfProtocols types.List) ([]components.WebsiteProtocol, error) {
	var protocols []components.WebsiteProtocol
	for _, p := range tfProtocols.Elements() {
		protocolStr := p.String()
		// Remove quotes from the string
		protocolStr = protocolStr[1 : len(protocolStr)-1]
		protocol, err := stringToWebsiteProtocol(protocolStr)
		if err != nil {
			return nil, err
		}
		protocols = append(protocols, protocol)
	}
	return protocols, nil
}

func mapPlatformsFromTerraform(tfPlatforms types.Set) ([]components.ProbePlatform, error) {
	var probePlatforms []components.ProbePlatform
	for _, p := range tfPlatforms.Elements() {
		platformStr := p.String()
		// Remove quotes from the string
		platformStr = platformStr[1 : len(platformStr)-1]
		platform, err := stringToProbePlatform(platformStr)
		if err != nil {
			return nil, err
		}
		probePlatforms = append(probePlatforms, platform)
	}
	return probePlatforms, nil
}

func mapAvailabilitySettings(ctx context.Context, tfAvailability availabilityMonitoring, tfMonitoring websiteMonitoring) (*components.AvailabilityCheckSettings, error) {
	availabilitySettings := &components.AvailabilityCheckSettings{
		TestIntervalInSeconds: float64(tfAvailability.TestIntervalInSeconds.ValueInt64()),
	}

	if !tfAvailability.CheckForString.IsNull() {
		var tfCheckForString checkForStringType
		if err := tfAvailability.CheckForString.As(ctx, &tfCheckForString, basetypes.ObjectAsOptions{}); err != nil {
			return nil, fmt.Errorf("failed to parse check_for_string: %s", err)
		}

		operator, err := stringToOperator(tfCheckForString.Operator.ValueString())
		if err != nil {
			return nil, err
		}

		availabilitySettings.CheckForString = &components.CheckForString{
			Operator: operator,
			Value:    tfCheckForString.Value.ValueString(),
		}
	}

	if !tfAvailability.SSL.IsNull() {
		var tfSslMonitoring sslMonitoring
		if err := tfAvailability.SSL.As(ctx, &tfSslMonitoring, basetypes.ObjectAsOptions{}); err != nil {
			return nil, fmt.Errorf("failed to parse SSL settings: %s", err)
		}
		if tfSslMonitoring.Enabled.ValueBool() {
			availabilitySettings.Ssl = &components.Ssl{
				Enabled:                        swov1.Bool(true),
				DaysPriorToExpiration:          swov1.Int(int(tfSslMonitoring.DaysPriorToExpiration.ValueInt64())),
				IgnoreIntermediateCertificates: swov1.Bool(tfSslMonitoring.IgnoreIntermediateCertificates.ValueBool()),
			}
		}
	}

	protocols, err := mapProtocolsFromTerraform(tfAvailability.Protocols)
	if err != nil {
		return nil, err
	}
	availabilitySettings.Protocols = protocols

	testFromType, err := stringToTestFromType(tfAvailability.TestFromLocation.ValueString())
	if err != nil {
		return nil, err
	}
	availabilitySettings.TestFrom = components.TestFrom{
		Type: testFromType,
	}

	var tfLocationOptions []probeLocation
	if err := tfAvailability.LocationOptions.ElementsAs(ctx, &tfLocationOptions, false); err != nil {
		return nil, fmt.Errorf("failed to parse location options: %s", err)
	}

	var locationValues []string
	for _, loc := range tfLocationOptions {
		locationValues = append(locationValues, loc.Value.ValueString())
	}
	availabilitySettings.TestFrom.Values = locationValues

	var tfPlatformOpts platformOptions
	if err := tfAvailability.PlatformOptions.As(ctx, &tfPlatformOpts, basetypes.ObjectAsOptions{}); err != nil {
		return nil, fmt.Errorf("failed to parse platform options: %s", err)
	}

	probePlatforms, err := mapPlatformsFromTerraform(tfPlatformOpts.Platforms)
	if err != nil {
		return nil, err
	}

	availabilitySettings.PlatformOptions = &components.WebsitePlatformOptions{
		ProbePlatforms: probePlatforms,
		TestFromAll:    swov1.Bool(tfPlatformOpts.TestFromAll.ValueBool()),
	}

	var tfCustomHeaders []customHeader
	if !tfAvailability.CustomHeaders.IsNull() {
		if err := tfAvailability.CustomHeaders.ElementsAs(ctx, &tfCustomHeaders, false); err != nil {
			return nil, fmt.Errorf("failed to parse availability custom headers: %s", err)
		}
	} else {
		if err := tfMonitoring.CustomHeaders.ElementsAs(ctx, &tfCustomHeaders, false); err != nil {
			return nil, fmt.Errorf("failed to parse monitoring custom headers: %s", err)
		}
	}

	if len(tfCustomHeaders) > 0 {
		var customHeaders []components.CustomHeaders
		for _, h := range tfCustomHeaders {
			customHeaders = append(customHeaders, components.CustomHeaders{
				Name:  h.Name.ValueString(),
				Value: h.Value.ValueString(),
			})
		}
		availabilitySettings.CustomHeaders = customHeaders
	}

	return availabilitySettings, nil
}

func mapRumSettings(ctx context.Context, tfRum rumMonitoring) *components.Rum {
	return &components.Rum{
		ApdexTimeInSeconds: swov1.Int(int(tfRum.ApdexTimeInSeconds.ValueInt64())),
		Spa:                tfRum.Spa.ValueBool(),
	}
}
