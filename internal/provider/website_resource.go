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
			Protocols:             convertWebsiteProtocols(tfPlan.Monitoring.Availability.Protocols),
			PlatformOptions: &swoClient.ProbePlatformOptionsInput{
				TestFromAll:    swoClient.Ptr(tfPlan.Monitoring.Availability.PlatformOptions.TestFromAll.ValueBool()),
				ProbePlatforms: convertWebsiteProbePlatforms(tfPlan.Monitoring.Availability.PlatformOptions.Platforms),
			},
			TestFrom: swoClient.ProbeLocationInput{
				Type:   swoClient.ProbeLocationType(tfPlan.Monitoring.Availability.TestFromLocation.ValueString()),
				Values: convertWebsiteProbeLocations(tfPlan.Monitoring.Availability.LocationOptions),
			},
			Ssl: &swoClient.SslMonitoringInput{
				Enabled:                        swoClient.Ptr(tfPlan.Monitoring.Availability.SSL.Enabled.ValueBool()),
				DaysPriorToExpiration:          swoClient.Ptr(int(tfPlan.Monitoring.Availability.SSL.DaysPriorToExpiration.ValueInt64())),
				IgnoreIntermediateCertificates: swoClient.Ptr(tfPlan.Monitoring.Availability.SSL.IgnoreIntermediateCertificates.ValueBool()),
			},
			CustomHeaders: convertWebsiteCustomHeaders(tfPlan.Monitoring.CustomHeaders),
		},
		Rum: &swoClient.RumMonitoringInput{
			ApdexTimeInSeconds: swoClient.Ptr(int(tfPlan.Monitoring.Rum.ApdexTimeInSeconds.ValueInt64())),
			Spa:                swoClient.Ptr(tfPlan.Monitoring.Rum.Spa.ValueBool()),
		},
	}

	// Create the Website...
	newWebsite, err := r.client.WebsiteService().Create(ctx, createInput)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("error creating website '%s' - error: %s",
			tfPlan.Name,
			err))
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("website %s created successfully - id=%s", tfPlan.Name, newWebsite.Id))
	tfPlan.Id = types.StringValue(newWebsite.Id)
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
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("error reading Website %s. error: %s",
			tfState.Id, err))
		return
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
			Protocols:             convertWebsiteProtocols(tfPlan.Monitoring.Availability.Protocols),
			PlatformOptions: &swoClient.ProbePlatformOptionsInput{
				TestFromAll:    swoClient.Ptr(tfPlan.Monitoring.Availability.PlatformOptions.TestFromAll.ValueBool()),
				ProbePlatforms: convertWebsiteProbePlatforms(tfPlan.Monitoring.Availability.PlatformOptions.Platforms),
			},
			TestFrom: swoClient.ProbeLocationInput{
				Type:   swoClient.ProbeLocationType(tfPlan.Monitoring.Availability.TestFromLocation.ValueString()),
				Values: convertWebsiteProbeLocations(tfPlan.Monitoring.Availability.LocationOptions),
			},
			Ssl: &swoClient.SslMonitoringInput{
				Enabled:                        swoClient.Ptr(tfPlan.Monitoring.Availability.SSL.Enabled.ValueBool()),
				DaysPriorToExpiration:          swoClient.Ptr(int(tfPlan.Monitoring.Availability.SSL.DaysPriorToExpiration.ValueInt64())),
				IgnoreIntermediateCertificates: swoClient.Ptr(tfPlan.Monitoring.Availability.SSL.IgnoreIntermediateCertificates.ValueBool()),
			},
			CustomHeaders: convertWebsiteCustomHeaders(tfPlan.Monitoring.CustomHeaders),
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
	resp.Diagnostics.Append(resp.State.Set(ctx, &tfPlan)...)
}

func (r *WebsiteResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var tfState WebsiteResourceModel

	// Read existing Terraform state.
	resp.Diagnostics.Append(req.State.Get(ctx, &tfState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// // Delete the Website...
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

func convertWebsiteProbePlatforms(in []string) []swoClient.ProbePlatform {
	var out []swoClient.ProbePlatform
	for _, p := range in {
		out = append(out, swoClient.ProbePlatform(p))
	}
	return out
}

func convertWebsiteCustomHeaders(in []CustomHeader) []swoClient.CustomHeaderInput {
	var out []swoClient.CustomHeaderInput
	for _, h := range in {
		out = append(out, swoClient.CustomHeaderInput{
			Name:  h.Name.ValueString(),
			Value: h.Value.ValueString(),
		})
	}
	return out
}

func convertWebsiteProbeLocations(in []ProbeLocation) []string {
	var out []string
	for _, p := range in {
		out = append(out, p.Value.ValueString())
	}
	return out
}

func convertWebsiteProtocols(in []string) []swoClient.WebsiteProtocol {
	var out []swoClient.WebsiteProtocol
	for _, p := range in {
		out = append(out, swoClient.WebsiteProtocol(p))
	}
	return out
}
