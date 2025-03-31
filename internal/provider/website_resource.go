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
	"github.com/solarwinds/swo-sdk-go/swov1"
	"github.com/solarwinds/swo-sdk-go/swov1/models/components"
	"github.com/solarwinds/swo-sdk-go/swov1/models/operations"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &websiteResource{}
	_ resource.ResourceWithConfigure   = &websiteResource{}
	_ resource.ResourceWithImportState = &websiteResource{}

	ErrWebsiteNotFound = errors.New("website not found")
)

const websiteValidationErrSummary = "Website Resource Validation Error"

func NewWebsiteResource() resource.Resource {
	return &websiteResource{}
}

// Defines the resource implementation.
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

	if tfPlan.Monitoring.Availability == nil && tfPlan.Monitoring.Rum == nil {
		resp.Diagnostics.AddError(websiteValidationErrSummary,
			fmt.Sprintf("error creating website '%s' - at least one monitoring configuration must be provided: rum, or availability.", tfPlan.Name))
		return
	}

	var availabilityCheckSettings *components.AvailabilityCheckSettings
	if tfPlan.Monitoring.Availability != nil {

		var protocols []components.WebsiteProtocol
		tfPlan.Monitoring.Availability.Protocols.ElementsAs(ctx, &protocols, false)

		var probePlatforms []components.ProbePlatform
		tfPlan.Monitoring.Availability.PlatformOptions.Platforms.ElementsAs(ctx, &protocols, false)

		availabilityCheckSettings = &components.AvailabilityCheckSettings{
			TestIntervalInSeconds: int(tfPlan.Monitoring.Availability.TestIntervalInSeconds.ValueInt64()),
			Protocols:             protocols,
			PlatformOptions: &components.ProbePlatformOptions{
				TestFromAll:    tfPlan.Monitoring.Availability.PlatformOptions.TestFromAll.ValueBoolPointer(),
				ProbePlatforms: probePlatforms,
			},
			TestFrom: components.TestFrom{
				Type: components.ProbeLocationType(tfPlan.Monitoring.Availability.TestFromLocation.ValueString()),
				Values: convertArray(tfPlan.Monitoring.Availability.LocationOptions, func(p probeLocation) string {
					return p.Value.ValueString()
				}),
			},
			CustomHeaders: convertArray(tfPlan.Monitoring.CustomHeaders, func(h customHeader) components.CustomHeaders {
				return components.CustomHeaders{
					Name:  h.Name.ValueString(),
					Value: h.Value.ValueString(),
				}
			}),
		}

		if tfPlan.Monitoring.Availability.CheckForString != nil {
			availabilityCheckSettings.CheckForString = &components.CheckForString{
				Operator: components.CheckForStringOperator(tfPlan.Monitoring.Availability.CheckForString.Operator.ValueString()),
				Value:    tfPlan.Monitoring.Availability.CheckForString.Value.ValueString(),
			}
		}

		if tfPlan.Monitoring.Availability.SSL != nil {
			availabilityCheckSettings.Ssl = &components.Ssl{
				Enabled:                        tfPlan.Monitoring.Availability.SSL.Enabled.ValueBoolPointer(),
				DaysPriorToExpiration:          swoClient.Ptr(int(tfPlan.Monitoring.Availability.SSL.DaysPriorToExpiration.ValueInt64())),
				IgnoreIntermediateCertificates: tfPlan.Monitoring.Availability.SSL.IgnoreIntermediateCertificates.ValueBoolPointer(),
			}
		}
	}

	var rum *components.Rum
	if tfPlan.Monitoring.Rum != nil {

		rum = &components.Rum{
			Spa: tfPlan.Monitoring.Rum.Spa.ValueBool(),
		}

		if tfPlan.Monitoring.Rum.ApdexTimeInSeconds.ValueInt32Pointer() != nil {
			rumApdexTimeInSeconds := int(tfPlan.Monitoring.Rum.ApdexTimeInSeconds.ValueInt32())
			rum.ApdexTimeInSeconds = &rumApdexTimeInSeconds
		}
	}

	websiteInput := components.Website{
		Name:                      tfPlan.Name.ValueString(),
		URL:                       tfPlan.Url.ValueString(),
		AvailabilityCheckSettings: availabilityCheckSettings,
		Rum:                       rum,
	}

	res, err := r.client.Dem.CreateWebsite(ctx, websiteInput)
	if err != nil {
		resp.Diagnostics.AddError(clientErrSummary,
			fmt.Sprintf("error creating website '%s' - error: %s", tfPlan.Name, err))
		return
	}

	website, err := r.RetryGetWebsite(ctx, res.EntityID.ID)

	// Create the Website...
	if err != nil {
		resp.Diagnostics.AddError(clientErrSummary,
			fmt.Sprintf("error reading webiste after create. '%s' - error: %s", tfPlan.Name, err))
		return
	}

	// Get the latest Website state from the server so we can get the 'snippet' field. Ideally we need to update
	// the API to return the 'snippet' field in the create response.
	tfPlan.Monitoring.Rum.Snippet = types.StringValue(*website.Object.Rum.Snippet)

	tfPlan.Id = types.StringValue(website.Object.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, tfPlan)...)
}

func (r *websiteResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var tfState websiteResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &tfState)...)

	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.RetryGetWebsite(ctx, tfState.Id.ValueString())

	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error reading Website %s. error: %s", tfState.Id, err))
		return
	}

	ws := res.Object

	// Update the Terraform state with latest values from the server.
	tfState.Url = types.StringValue(ws.URL)
	tfState.Name = types.StringValue(ws.Name)

	tfState.Monitoring = websiteMonitoring{}
	tfState.Monitoring.Availability = &availabilityMonitoring{}
	availability := ws.AvailabilityCheckSettings

	if availability.CheckForString != nil {
		tfState.Monitoring.Availability.CheckForString = &checkForStringType{
			Operator: types.StringValue(string(availability.CheckForString.Operator)),
			Value:    types.StringValue(availability.CheckForString.Value),
		}
	} else {
		tfState.Monitoring.Availability.CheckForString = nil
	}

	if availability.TestIntervalInSeconds != int(*types.Int64Null().ValueInt64Pointer()) {
		tfState.Monitoring.Availability.TestIntervalInSeconds = types.Int64Value(int64(availability.TestIntervalInSeconds))
	} else {
		tfState.Monitoring.Availability.TestIntervalInSeconds = types.Int64Null()
	}

	tfState.Monitoring.Availability.Protocols = sliceToStringList(availability.Protocols, func(s components.WebsiteProtocol) string {
		return string(s)
	})

	if availability.PlatformOptions != nil {
		tfState.Monitoring.Availability.PlatformOptions = platformOptions{
			TestFromAll: types.BoolValue(*availability.PlatformOptions.TestFromAll),
			Platforms: sliceToStringList(availability.PlatformOptions.ProbePlatforms, func(s components.ProbePlatform) string {
				return string(s)
			}),
		}
	} else {
		tfState.Monitoring.Availability.PlatformOptions = platformOptions{}
	}

	if availability.TestFrom.Type.ToPointer() != nil {
		tfState.Monitoring.Availability.TestFromLocation = types.StringValue(string(availability.TestFrom.Type))
	} else {
		tfState.Monitoring.Availability.TestFromLocation = types.StringNull()
	}

	if availability.TestFrom.Values != nil {
		var locOpts []probeLocation
		for _, p := range availability.TestFrom.Values {
			locOpts = append(locOpts, probeLocation{
				Type:  types.StringValue(string(availability.TestFrom.Type)),
				Value: types.StringValue(p),
			})
		}
		tfState.Monitoring.Availability.LocationOptions = locOpts
	} else {
		tfState.Monitoring.Availability.LocationOptions = nil
	}

	if availability.Ssl != nil {
		tfState.Monitoring.Availability.SSL = &sslMonitoring{
			Enabled:                        types.BoolValue(*availability.Ssl.Enabled),
			IgnoreIntermediateCertificates: types.BoolValue(*availability.Ssl.IgnoreIntermediateCertificates),
		}
		if availability.Ssl.DaysPriorToExpiration != nil {
			tfState.Monitoring.Availability.SSL.DaysPriorToExpiration = types.Int64Value(int64(*availability.Ssl.DaysPriorToExpiration))
		} else {
			tfState.Monitoring.Availability.SSL.DaysPriorToExpiration = types.Int64Null()
		}
	} else {
		tfState.Monitoring.Availability.SSL = nil
	}

	var customHeaders []customHeader
	if availability.CustomHeaders != nil {
		for _, h := range availability.CustomHeaders {
			customHeaders = append(customHeaders, customHeader{
				Name:  types.StringValue(h.Name),
				Value: types.StringValue(h.Value),
			})
		}
	}
	tfState.Monitoring.CustomHeaders = customHeaders

	if ws.Rum != nil {
		tfState.Monitoring.Rum = &rumMonitoring{
			Spa: types.BoolValue(ws.Rum.Spa),
		}

		if ws.Rum.ApdexTimeInSeconds != nil {
			var apdexTimeInSeconds int32
			apdexTimeInSeconds, resp.Diagnostics = safeIntToInt32(*ws.Rum.ApdexTimeInSeconds, resp.Diagnostics)
			if resp.Diagnostics.HasError() {
				return
			}

			tfState.Monitoring.Rum.ApdexTimeInSeconds = types.Int32Value(apdexTimeInSeconds)
		}

		if ws.Rum.Snippet != nil {
			tfState.Monitoring.Rum.Snippet = types.StringValue(*ws.Rum.Snippet)
		}
	} else {
		tfState.Monitoring.Rum = &rumMonitoring{}
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

	var checkForString *components.CheckForString
	if tfPlan.Monitoring.Availability.CheckForString != nil {
		checkForString = &components.CheckForString{
			Operator: components.CheckForStringOperator(tfPlan.Monitoring.Availability.CheckForString.Operator.ValueString()),
			Value:    tfPlan.Monitoring.Availability.CheckForString.Value.ValueString(),
		}
	}
	var ssl *components.Ssl
	if tfPlan.Monitoring.Availability.SSL != nil {
		ssl = &components.Ssl{
			Enabled:                        tfPlan.Monitoring.Availability.SSL.Enabled.ValueBoolPointer(),
			DaysPriorToExpiration:          swoClient.Ptr(int(tfPlan.Monitoring.Availability.SSL.DaysPriorToExpiration.ValueInt64())),
			IgnoreIntermediateCertificates: tfPlan.Monitoring.Availability.SSL.IgnoreIntermediateCertificates.ValueBoolPointer(),
		}
	}

	rumApdexTimeInSeconds := int(tfPlan.Monitoring.Rum.ApdexTimeInSeconds.ValueInt32())

	var protocols []components.WebsiteProtocol
	tfPlan.Monitoring.Availability.Protocols.ElementsAs(ctx, &protocols, false)

	var probePlatforms []components.ProbePlatform
	tfPlan.Monitoring.Availability.PlatformOptions.Platforms.ElementsAs(ctx, &protocols, false)

	updateWebsiteReq := operations.UpdateWebsiteRequest{
		EntityID: tfState.Id.ValueString(),
		Website: components.Website{
			Name: tfPlan.Name.ValueString(),
			URL:  tfPlan.Url.ValueString(),
			AvailabilityCheckSettings: &components.AvailabilityCheckSettings{
				CheckForString:        checkForString,
				TestIntervalInSeconds: int(tfPlan.Monitoring.Availability.TestIntervalInSeconds.ValueInt64()),
				Protocols:             protocols,
				PlatformOptions: &components.ProbePlatformOptions{
					TestFromAll:    swoClient.Ptr(tfPlan.Monitoring.Availability.PlatformOptions.TestFromAll.ValueBool()),
					ProbePlatforms: probePlatforms,
				},
				TestFrom: components.TestFrom{
					Type: components.ProbeLocationType(tfPlan.Monitoring.Availability.TestFromLocation.ValueString()),
					Values: convertArray(tfPlan.Monitoring.Availability.LocationOptions, func(p probeLocation) string {
						return p.Value.ValueString()
					}),
				},
				Ssl: ssl,
				CustomHeaders: convertArray(tfPlan.Monitoring.CustomHeaders, func(h customHeader) components.CustomHeaders {
					return components.CustomHeaders{
						Name:  h.Name.ValueString(),
						Value: h.Value.ValueString(),
					}
				}),
			},
			Rum: &components.Rum{
				ApdexTimeInSeconds: &rumApdexTimeInSeconds,
				Spa:                *tfPlan.Monitoring.Rum.Spa.ValueBoolPointer(),
			},
		},
	}

	// Update the Website...
	res, err := r.client.Dem.UpdateWebsite(ctx, updateWebsiteReq)

	if err != nil {
		resp.Diagnostics.AddError(clientErrSummary,
			fmt.Sprintf("error updating website %s. err: %s", tfState.Id, err))
		return
	}

	if res.EntityID != nil {
		resp.Diagnostics.AddError(clientErrSummary,
			fmt.Sprintf("error updating website %s. entity id not returned", tfState.Id))
		return
	}

	website := updateWebsiteReq.Website
	availability := website.AvailabilityCheckSettings

	websiteToCompare := operations.GetWebsiteResponseBody{
		ID:   updateWebsiteReq.EntityID,
		Name: website.Name,
		URL:  website.URL,
		AvailabilityCheckSettings: &operations.AvailabilityCheckSettings{
			CheckForString:             (*operations.CheckForString)(availability.CheckForString),
			TestIntervalInSeconds:      availability.TestIntervalInSeconds,
			Protocols:                  availability.Protocols,
			PlatformOptions:            availability.PlatformOptions,
			TestFrom:                   availability.TestFrom,
			Ssl:                        (*operations.Ssl)(availability.Ssl),
			CustomHeaders:              availability.CustomHeaders,
			AllowInsecureRenegotiation: availability.AllowInsecureRenegotiation,
			PostData:                   availability.PostData,
		},
	}

	// Updates are eventually consistent. Retry until the Website we read and the Website we are updating match.
	_, err = BackoffRetry(func() (*operations.GetWebsiteResponse, error) {

		websiteReq := operations.GetWebsiteRequest{
			EntityID: res.EntityID.ID,
		}

		res, err := r.client.Dem.GetWebsite(ctx, websiteReq)
		if err != nil {
			return nil, backoff.Permanent(err)
		}

		if res.Object != nil {
			resBody := res.Object
			websiteToCompare.Type = resBody.Type
			websiteToCompare.Status = resBody.Status

			match := reflect.DeepEqual(&websiteToCompare, resBody)

			// Updated entity properties don't match, retry
			if !match {
				return nil, ErrNonMatchingEntites
			}
		}

		return res, nil
	})

	if err != nil {
		resp.Diagnostics.AddError(clientErrSummary,
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

	ws := operations.DeleteWebsiteRequest{
		EntityID: tfState.Id.ValueString(),
	}

	// Delete the Website...
	if _, err := r.client.Dem.DeleteWebsite(ctx, ws); err != nil {
		resp.Diagnostics.AddError(clientErrSummary,
			fmt.Sprintf("error deleting website %s - %s", tfState.Id, err))
	}
}

func (r *websiteResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// RetryGetWebsite Website Creates and Updates are eventually consistent. Retry until the entity id is returned.
func (r *websiteResource) RetryGetWebsite(ctx context.Context, entityId string) (operations.GetWebsiteResponse, error) {
	website := operations.GetWebsiteRequest{
		EntityID: entityId,
	}

	res, err := BackoffRetry(func() (operations.GetWebsiteResponse, error) {
		res, err := r.client.Dem.GetWebsite(ctx, website)
		if err != nil {
			// The entity is still being created, retry
			if res.Object == nil {
				return *res, ErrWebsiteNotFound
			}

			return *res, backoff.Permanent(err)
		}

		return *res, nil
	})

	if err != nil {
		return res, err
	}

	return res, nil
}
