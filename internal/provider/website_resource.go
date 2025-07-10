package provider

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/cenkalti/backoff/v5"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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
	ErrNoWebsiteDataReturned          = errors.New("no website data returned")
	ErrWebsiteEntityNoData            = errors.New("website entity exists but has no data")
	ErrNoWebsiteEntityReturned        = errors.New("no website entity returned")
	ErrAvailabilitySettingsNotUpdated = errors.New("availability settings not yet updated")
	ErrRUMSettingsNotUpdated          = errors.New("RUM settings not yet updated")
	ErrWebsiteNameNotUpdated          = errors.New("website name not yet updated")
	ErrWebsiteURLNotUpdated           = errors.New("website URL not yet updated")
	ErrUnsupportedOperator            = errors.New("unsupported operator")
	ErrUnsupportedProtocol            = errors.New("unsupported protocol")
	ErrUnsupportedPlatform            = errors.New("unsupported platform")
	ErrUnsupportedTestFromType        = errors.New("unsupported test from type")
	ErrFailedParseCheckForString      = errors.New("failed to parse check_for_string")
	ErrFailedParseSSLSettings         = errors.New("failed to parse SSL settings")
	ErrFailedParseLocationOptions     = errors.New("failed to parse location options")
	ErrFailedParsePlatformOptions     = errors.New("failed to parse platform options")
	ErrFailedParseAvailabilityHeaders = errors.New("failed to parse availability custom headers")
	ErrFailedParseMonitoringHeaders   = errors.New("failed to parse monitoring custom headers")
	ErrFailedParseOutageConfig        = errors.New("failed to parse outage configuration")
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

	// Build the website input
	var tags []websiteTag
	d := tfPlan.Tags.ElementsAs(ctx, &tags, false)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	createInput := components.Website{
		Name: tfPlan.Name.ValueString(),
		URL:  tfPlan.Url.ValueString(),
		Tags: convertArray(tags, func(e websiteTag) components.Tag {
			return components.Tag{
				Key:   e.Key.ValueString(),
				Value: e.Value.ValueString(),
			}
		}),
	}

	// Parse monitoring configuration
	var tfMonitoring websiteMonitoring
	d = tfPlan.Monitoring.As(ctx, &tfMonitoring, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Configure availability monitoring
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

	// Configure RUM
	if !tfMonitoring.Rum.IsNull() {
		var tfRum rumMonitoring
		d = tfMonitoring.Rum.As(ctx, &tfRum, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}

		createInput.Rum = mapRumSettings(tfRum)
	}

	// Create the website in swo
	res, err := r.client.Dem.CreateWebsite(ctx, createInput)
	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error creating website '%s' - error: %s", tfPlan.Name.ValueString(), err))
		return
	}

	if res.EntityID == nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error creating website '%s' - no entity ID returned", tfPlan.Name.ValueString()))
		return
	}

	// Set the ID from the creation response
	tfPlan.Id = types.StringValue(res.EntityID.GetID())

	// Set computed monitoring options based on what was configured
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

	// Get the created website to get computed fields
	websiteResp, err := r.client.Dem.GetWebsite(ctx, operations.GetWebsiteRequest{
		EntityID: tfPlan.Id.ValueString(),
	})

	// Update RUM with computed snippet field
	if !tfMonitoring.Rum.IsNull() {
		var rum rumMonitoring
		d = tfMonitoring.Rum.As(ctx, &rum, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}

		// Set RUM snippet from server response if available
		if err == nil && websiteResp.GetWebsiteResponse != nil &&
			websiteResp.GetWebsiteResponse.Rum != nil &&
			websiteResp.GetWebsiteResponse.Rum.Snippet != nil {
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

	// Warn if we couldn't read the website after creation
	if err != nil {
		resp.Diagnostics.AddWarning("Client Error",
			fmt.Sprintf("error reading website after create '%s' - error: %s", tfPlan.Name.ValueString(), err))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, tfPlan)...)
}

func (r *websiteResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var tfState websiteResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &tfState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read operation for retry logic
	readOperation := func(ctx context.Context, id string) (*components.GetWebsiteResponse, error) {
		websiteResp, err := r.client.Dem.GetWebsite(ctx, operations.GetWebsiteRequest{
			EntityID: id,
		})

		if err != nil {
			return nil, err
		}

		if websiteResp.GetWebsiteResponse == nil {
			return nil, ErrNoWebsiteDataReturned
		}

		return websiteResp.GetWebsiteResponse, nil
	}

	// GET website data with retry
	website, err := websiteReadRetry(ctx, tfState.Id.ValueString(), readOperation)
	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error reading website %s. error: %s", tfState.Name, err))
		return
	}

	// Update basic website fields
	tfState.Url = types.StringValue(website.URL)
	tfState.Name = types.StringValue(website.Name)
	var tagElements []attr.Value
	for _, x := range website.Tags {
		objectValue, d := types.ObjectValueFrom(
			ctx,
			WebsiteTagAttributeTypes(),
			websiteTag{
				Key:   types.StringValue(x.Key),
				Value: types.StringValue(x.Value),
			},
		)

		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}
		tagElements = append(tagElements, objectValue)
	}
	tags, d := types.SetValueFrom(ctx, types.ObjectType{AttrTypes: WebsiteTagAttributeTypes()}, tagElements)
	tfState.Tags = tags

	// Build monitoring configuration from server response
	monitoring, d := r.buildMonitoringFromServerResponse(ctx, website)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	tfState.Monitoring = monitoring

	// Save the updated state
	resp.Diagnostics.Append(resp.State.Set(ctx, tfState)...)
}

func (r *websiteResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var tfPlan, tfState *websiteResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &tfPlan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &tfState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build the update input
	var tags []websiteTag
	d := tfPlan.Tags.ElementsAs(ctx, &tags, false)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	updateInput := components.Website{
		Name: tfPlan.Name.ValueString(),
		URL:  tfPlan.Url.ValueString(),
		Tags: convertArray(tags, func(e websiteTag) components.Tag {
			return components.Tag{
				Key:   e.Key.ValueString(),
				Value: e.Value.ValueString(),
			}
		}),
	}

	// Parse monitoring configuration from the plan
	var tfMonitoring websiteMonitoring
	d = tfPlan.Monitoring.As(ctx, &tfMonitoring, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Configure availability monitoring
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

	// Configure RUM
	if !tfMonitoring.Rum.IsNull() {
		var tfRum rumMonitoring
		d = tfMonitoring.Rum.As(ctx, &tfRum, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}

		updateInput.Rum = mapRumSettings(tfRum)
	}

	// POST the update
	_, err := r.client.Dem.UpdateWebsite(ctx, operations.UpdateWebsiteRequest{
		EntityID: tfState.Id.ValueString(),
		Website:  updateInput,
	})

	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error updating website %s. err: %s", tfState.Id.ValueString(), err))
		return
	}

	// Read operation with eventual consistency website validation
	readOperation := func(ctx context.Context, id string) (*components.GetWebsiteResponse, error) {
		websiteResp, err := r.client.Dem.GetWebsite(ctx, operations.GetWebsiteRequest{
			EntityID: id,
		})

		if err != nil {
			return nil, err
		}

		if websiteResp.GetWebsiteResponse == nil {
			return nil, ErrNoWebsiteDataReturned
		}

		website := websiteResp.GetWebsiteResponse

		// Validate that the basic fields have been updated
		expectedName := tfPlan.Name.ValueString()
		if website.Name != expectedName {
			return nil, ErrWebsiteNameNotUpdated
		}

		if website.URL != tfPlan.Url.ValueString() {
			return nil, ErrWebsiteURLNotUpdated
		}

		// Validate monitoring settings exist if configured in the plan
		if !tfMonitoring.Availability.IsNull() && website.AvailabilityCheckSettings == nil {
			return nil, ErrAvailabilitySettingsNotUpdated
		}

		if !tfMonitoring.Rum.IsNull() && website.Rum == nil {
			return nil, ErrRUMSettingsNotUpdated
		}

		return website, nil
	}

	// Read the updated website with retry for eventual consistency
	website, err := websiteReadRetry(ctx, tfState.Id.ValueString(), readOperation)
	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error reading website after update %s. error: %s", tfState.Id.ValueString(), err))
		return
	}

	// Update plan with values from the server response
	tfPlan.Name = types.StringValue(website.Name)
	tfPlan.Url = types.StringValue(website.URL)
	var tagElements []attr.Value
	for _, x := range website.Tags {
		objectValue, d := types.ObjectValueFrom(
			ctx,
			WebsiteTagAttributeTypes(),
			websiteTag{
				Key:   types.StringValue(x.Key),
				Value: types.StringValue(x.Value),
			},
		)

		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}
		tagElements = append(tagElements, objectValue)
	}
	websiteTags, d := types.SetValueFrom(ctx, types.ObjectType{AttrTypes: WebsiteTagAttributeTypes()}, tagElements)
	tfState.Tags = websiteTags

	// Set computed monitoring options based on user's plan configuration
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

	// Update RUM with the snippet if RUM is configured
	if !tfMonitoring.Rum.IsNull() && website.Rum != nil {
		var tfRum rumMonitoring
		d = tfMonitoring.Rum.As(ctx, &tfRum, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}

		// Set RUM snippet from server response
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
	return "", ErrUnsupportedOperator
}

func stringToWebsiteProtocol(protocolStr string) (components.WebsiteProtocol, error) {
	if protocol, exists := websiteProtocolMap[protocolStr]; exists {
		return protocol, nil
	}
	return "", ErrUnsupportedProtocol
}

func stringToProbePlatform(platformStr string) (components.ProbePlatform, error) {
	if platform, exists := probePlatformMap[platformStr]; exists {
		return platform, nil
	}
	return "", ErrUnsupportedPlatform
}

func stringToTestFromType(typeStr string) (components.Type, error) {
	if testFromType, exists := testFromTypeMap[typeStr]; exists {
		return testFromType, nil
	}
	return "", ErrUnsupportedTestFromType
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
			return nil, ErrNoWebsiteEntityReturned
		}

		if result.ID == "" {
			return nil, ErrWebsiteEntityNoData
		}

		return result, nil
	}, backoff.WithBackOff(expBackoff), backoff.WithMaxElapsedTime(websiteExpBackoffMaxElapsed))

	return website, err
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
			return nil, ErrFailedParseCheckForString
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
			return nil, ErrFailedParseSSLSettings
		}
		if tfSslMonitoring.Enabled.ValueBool() {
			availabilitySettings.Ssl = &components.Ssl{
				Enabled:                        swov1.Bool(true),
				DaysPriorToExpiration:          swov1.Int(int(tfSslMonitoring.DaysPriorToExpiration.ValueInt64())),
				IgnoreIntermediateCertificates: swov1.Bool(tfSslMonitoring.IgnoreIntermediateCertificates.ValueBool()),
			}
		}
	}

	if !tfAvailability.OutageConfig.IsNull() {
		var tfOutageConfig outageConfig
		if err := tfAvailability.OutageConfig.As(ctx, &tfOutageConfig, basetypes.ObjectAsOptions{}); err != nil {
			return nil, ErrFailedParseOutageConfig
		}
		availabilitySettings.OutageConfiguration = &components.WebsiteOutageConfiguration{
			FailingTestLocations: components.WebsiteFailingTestLocations(tfOutageConfig.FailingTestLocations.ValueString()),
			ConsecutiveForDown:   int(tfOutageConfig.ConsecutiveForDown.ValueInt64()),
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
		return nil, ErrFailedParseLocationOptions
	}

	var locationValues []string
	for _, loc := range tfLocationOptions {
		locationValues = append(locationValues, loc.Value.ValueString())
	}
	availabilitySettings.TestFrom.Values = locationValues

	var tfPlatformOpts platformOptions
	if err := tfAvailability.PlatformOptions.As(ctx, &tfPlatformOpts, basetypes.ObjectAsOptions{}); err != nil {
		return nil, ErrFailedParsePlatformOptions
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
			return nil, ErrFailedParseAvailabilityHeaders
		}
	} else {
		if err := tfMonitoring.CustomHeaders.ElementsAs(ctx, &tfCustomHeaders, false); err != nil {
			return nil, ErrFailedParseMonitoringHeaders
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

func mapRumSettings(tfRum rumMonitoring) *components.Rum {
	return &components.Rum{
		ApdexTimeInSeconds: swov1.Int(int(tfRum.ApdexTimeInSeconds.ValueInt64())),
		Spa:                tfRum.Spa.ValueBool(),
	}
}

// Builds the monitoring configuration from server response
func (r *websiteResource) buildMonitoringFromServerResponse(ctx context.Context, website *components.GetWebsiteResponse) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	monitoringOpts := website.MonitoringOptions
	serverMonitoringOptions := monitoringOptions{
		IsAvailabilityActive: types.BoolValue(monitoringOpts.IsAvailabilityActive),
		IsRumActive:          types.BoolValue(monitoringOpts.IsRumActive),
	}

	serverOptions, d := types.ObjectValueFrom(ctx, MonitoringOptionsAttributeTypes(), serverMonitoringOptions)
	diags.Append(d...)
	if diags.HasError() {
		return types.ObjectNull(WebsiteMonitoringAttributeTypes()), diags
	}

	tfStateMonitoring := websiteMonitoring{
		Options:       serverOptions,
		Availability:  types.ObjectNull(AvailabilityMonitoringAttributeTypes()),
		Rum:           types.ObjectNull(RumMonitoringAttributeTypes()),
		CustomHeaders: types.SetNull(types.ObjectType{AttrTypes: CustomHeaderAttributeTypes()}),
	}

	// Build availability monitoring if active
	if website.AvailabilityCheckSettings != nil && serverMonitoringOptions.IsAvailabilityActive.ValueBool() {
		availabilityValue, d := r.buildAvailabilityMonitoring(ctx, website.AvailabilityCheckSettings)
		diags.Append(d...)
		if diags.HasError() {
			return types.ObjectNull(WebsiteMonitoringAttributeTypes()), diags
		}
		tfStateMonitoring.Availability = availabilityValue

		customHeaders, d := r.buildCustomHeaders(ctx, website.AvailabilityCheckSettings.CustomHeaders)
		diags.Append(d...)
		if diags.HasError() {
			return types.ObjectNull(WebsiteMonitoringAttributeTypes()), diags
		}

		if !customHeaders.IsNull() {
			// Update availability monitoring with custom headers
			var availability availabilityMonitoring
			d = tfStateMonitoring.Availability.As(ctx, &availability, basetypes.ObjectAsOptions{})
			diags.Append(d...)
			if diags.HasError() {
				return types.ObjectNull(WebsiteMonitoringAttributeTypes()), diags
			}

			if !availability.CustomHeaders.IsNull() {
				availability.CustomHeaders = customHeaders
				tfStateMonitoring.CustomHeaders = types.SetNull(types.ObjectType{AttrTypes: CustomHeaderAttributeTypes()})
			} else {
				availability.CustomHeaders = types.SetNull(types.ObjectType{AttrTypes: CustomHeaderAttributeTypes()})
				tfStateMonitoring.CustomHeaders = customHeaders
			}

			availabilityValue, d := types.ObjectValueFrom(ctx, AvailabilityMonitoringAttributeTypes(), availability)
			diags.Append(d...)
			if diags.HasError() {
				return types.ObjectNull(WebsiteMonitoringAttributeTypes()), diags
			}
			tfStateMonitoring.Availability = availabilityValue
		}
	}

	// Build RUM monitoring if active
	if website.Rum != nil && serverMonitoringOptions.IsRumActive.ValueBool() {
		rumValue, d := r.buildRumMonitoring(ctx, website.Rum)
		diags.Append(d...)
		if diags.HasError() {
			return types.ObjectNull(WebsiteMonitoringAttributeTypes()), diags
		}
		tfStateMonitoring.Rum = rumValue
	}

	monitoringValue, d := types.ObjectValueFrom(ctx, WebsiteMonitoringAttributeTypes(), tfStateMonitoring)
	diags.Append(d...)
	if diags.HasError() {
		return types.ObjectNull(WebsiteMonitoringAttributeTypes()), diags
	}

	return monitoringValue, diags
}

// Build availability monitoring configuration from server response
func (r *websiteResource) buildAvailabilityMonitoring(ctx context.Context, availability *components.GetWebsiteResponseAvailabilityCheckSettings) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	tfAvailability := availabilityMonitoring{
		CheckForString:        types.ObjectNull(CheckForStringTypeAttributeTypes()),
		SSL:                   types.ObjectNull(SslMonitoringAttributeTypes()),
		Protocols:             types.ListNull(types.StringType),
		TestFromLocation:      types.StringNull(),
		TestIntervalInSeconds: types.Int64Null(),
		LocationOptions:       types.SetUnknown(types.ObjectType{AttrTypes: ProbeLocationAttributeTypes()}),
		PlatformOptions:       types.ObjectNull(PlatformOptionsAttributeTypes()),
		CustomHeaders:         types.SetNull(types.ObjectType{AttrTypes: CustomHeaderAttributeTypes()}),
		OutageConfig:          types.ObjectNull(OutageConfigAttributeTypes()),
	}

	if availability.OutageConfiguration != nil {
		outageConfigValue := outageConfig{
			FailingTestLocations: types.StringValue(string(availability.OutageConfiguration.FailingTestLocations)),
			ConsecutiveForDown:   types.Int64Value(int64(availability.OutageConfiguration.ConsecutiveForDown)),
		}
		outageConfig, d := types.ObjectValueFrom(ctx, OutageConfigAttributeTypes(), outageConfigValue)
		diags.Append(d...)
		if diags.HasError() {
			return types.ObjectNull(AvailabilityMonitoringAttributeTypes()), diags
		}
		tfAvailability.OutageConfig = outageConfig
	}

	// Map check for string configuration
	if availability.CheckForString != nil {
		checkForStringValue := checkForStringType{
			Operator: types.StringValue(string(availability.CheckForString.Operator)),
			Value:    types.StringValue(availability.CheckForString.Value),
		}
		checkForString, d := types.ObjectValueFrom(ctx, CheckForStringTypeAttributeTypes(), checkForStringValue)
		diags.Append(d...)
		if diags.HasError() {
			return types.ObjectNull(AvailabilityMonitoringAttributeTypes()), diags
		}
		tfAvailability.CheckForString = checkForString
	}

	// Map test interval
	if availability.TestIntervalInSeconds != 0 {
		tfAvailability.TestIntervalInSeconds = types.Int64Value(int64(availability.TestIntervalInSeconds))
	}

	// Map protocols
	if len(availability.Protocols) > 0 {
		tfAvailability.Protocols = sliceToStringList(availability.Protocols, func(s components.WebsiteProtocol) string {
			return string(s)
		})
	}

	// Map platform options
	if availability.PlatformOptions != nil {
		platformValue, d := r.buildPlatformOptions(ctx, availability.PlatformOptions)
		diags.Append(d...)
		if diags.HasError() {
			return types.ObjectNull(AvailabilityMonitoringAttributeTypes()), diags
		}
		tfAvailability.PlatformOptions = platformValue
	}

	// Map test from location and options
	if availability.TestFrom.Type != "" {
		tfAvailability.TestFromLocation = types.StringValue(string(availability.TestFrom.Type))

		if len(availability.TestFrom.Values) > 0 {
			locationOptions, d := r.buildLocationOptions(ctx, availability.TestFrom.Values)
			diags.Append(d...)
			if diags.HasError() {
				return types.ObjectNull(AvailabilityMonitoringAttributeTypes()), diags
			}
			tfAvailability.LocationOptions = locationOptions
		}
	}

	// Map SSL configuration
	sslValue, d := r.buildSSLMonitoring(ctx, availability.Ssl)
	diags.Append(d...)
	if diags.HasError() {
		return types.ObjectNull(AvailabilityMonitoringAttributeTypes()), diags
	}
	tfAvailability.SSL = sslValue

	availabilityValue, d := types.ObjectValueFrom(ctx, AvailabilityMonitoringAttributeTypes(), tfAvailability)
	diags.Append(d...)
	if diags.HasError() {
		return types.ObjectNull(AvailabilityMonitoringAttributeTypes()), diags
	}

	return availabilityValue, diags
}

// Builds platform options from server response
func (r *websiteResource) buildPlatformOptions(ctx context.Context, platformOpts *components.GetWebsiteResponsePlatformOptions) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	var platforms []types.String
	if platformOpts.ProbePlatforms != nil {
		for _, p := range platformOpts.ProbePlatforms {
			platforms = append(platforms, types.StringValue(string(p)))
		}
	}

	platformValue, d := types.SetValueFrom(ctx, types.StringType, platforms)
	diags.Append(d...)
	if diags.HasError() {
		return types.ObjectNull(PlatformOptionsAttributeTypes()), diags
	}

	testFromAll := false
	if platformOpts.TestFromAll != nil {
		testFromAll = *platformOpts.TestFromAll
	}

	platformOptionsValue := platformOptions{
		TestFromAll: types.BoolValue(testFromAll),
		Platforms:   platformValue,
	}

	platformObject, d := types.ObjectValueFrom(ctx, PlatformOptionsAttributeTypes(), platformOptionsValue)
	diags.Append(d...)
	if diags.HasError() {
		return types.ObjectNull(PlatformOptionsAttributeTypes()), diags
	}

	return platformObject, diags
}

// Builds location options from server response
func (r *websiteResource) buildLocationOptions(ctx context.Context, values []string) (types.Set, diag.Diagnostics) {
	var diags diag.Diagnostics
	var locOpts []attr.Value

	for _, value := range values {
		objectValue, d := types.ObjectValueFrom(
			ctx,
			ProbeLocationAttributeTypes(),
			probeLocation{
				Type:  types.StringValue(string(components.TypeRegion)),
				Value: types.StringValue(value),
			},
		)

		diags.Append(d...)
		if diags.HasError() {
			return types.SetNull(types.ObjectType{AttrTypes: ProbeLocationAttributeTypes()}), diags
		}
		locOpts = append(locOpts, objectValue)
	}

	locationOptions, d := types.SetValueFrom(ctx, types.ObjectType{AttrTypes: ProbeLocationAttributeTypes()}, locOpts)
	diags.Append(d...)
	if diags.HasError() {
		return types.SetNull(types.ObjectType{AttrTypes: ProbeLocationAttributeTypes()}), diags
	}

	return locationOptions, diags
}

// Builds SSL monitoring configuration from server response
func (r *websiteResource) buildSSLMonitoring(ctx context.Context, ssl *components.GetWebsiteResponseSsl) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics
	sslTypes := SslMonitoringAttributeTypes()

	if ssl != nil && ssl.Enabled != nil && *ssl.Enabled {
		sslValues := sslMonitoring{
			Enabled:                        types.BoolValue(*ssl.Enabled),
			IgnoreIntermediateCertificates: types.BoolValue(ssl.IgnoreIntermediateCertificates != nil && *ssl.IgnoreIntermediateCertificates),
			DaysPriorToExpiration:          types.Int64Null(),
		}
		if ssl.DaysPriorToExpiration != nil {
			sslValues.DaysPriorToExpiration = types.Int64Value(int64(*ssl.DaysPriorToExpiration))
		}

		objectValue, d := types.ObjectValueFrom(ctx, sslTypes, sslValues)
		diags.Append(d...)
		if diags.HasError() {
			return types.ObjectNull(sslTypes), diags
		}

		return objectValue, diags
	}

	return types.ObjectNull(sslTypes), diags
}

// Build custom headers from server response
func (r *websiteResource) buildCustomHeaders(ctx context.Context, headers []components.CustomHeaders) (types.Set, diag.Diagnostics) {
	var diags diag.Diagnostics
	customHeaderElementTypes := CustomHeaderAttributeTypes()

	if len(headers) == 0 {
		return types.SetNull(types.ObjectType{AttrTypes: customHeaderElementTypes}), diags
	}

	var elements []attr.Value
	for _, h := range headers {
		objectValue, d := types.ObjectValueFrom(
			ctx,
			customHeaderElementTypes,
			customHeader{
				Name:  types.StringValue(h.Name),
				Value: types.StringValue(h.Value),
			},
		)

		diags.Append(d...)
		if diags.HasError() {
			return types.SetNull(types.ObjectType{AttrTypes: customHeaderElementTypes}), diags
		}
		elements = append(elements, objectValue)
	}

	customHeaderValue, d := types.SetValueFrom(ctx, types.ObjectType{AttrTypes: customHeaderElementTypes}, elements)
	diags.Append(d...)
	if diags.HasError() {
		return types.SetNull(types.ObjectType{AttrTypes: customHeaderElementTypes}), diags
	}

	return customHeaderValue, diags
}

// Build RUM monitoring configuration from server response
func (r *websiteResource) buildRumMonitoring(ctx context.Context, rum *components.GetWebsiteResponseRum) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	rumValue := rumMonitoring{
		Spa:                types.BoolValue(rum.Spa),
		ApdexTimeInSeconds: types.Int64Null(),
		Snippet:            types.StringNull(),
	}

	if rum.ApdexTimeInSeconds != nil {
		rumValue.ApdexTimeInSeconds = types.Int64Value(int64(*rum.ApdexTimeInSeconds))
	}

	if rum.Snippet != nil {
		rumValue.Snippet = types.StringValue(*rum.Snippet)
	}

	rumObject, d := types.ObjectValueFrom(ctx, RumMonitoringAttributeTypes(), rumValue)
	diags.Append(d...)
	if diags.HasError() {
		return types.ObjectNull(RumMonitoringAttributeTypes()), diags
	}

	return rumObject, diags
}
