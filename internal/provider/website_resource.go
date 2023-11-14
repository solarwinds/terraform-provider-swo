package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	swoClient "github.com/solarwindscloud/swo-client-go/pkg/client"
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

	var plan WebsiteResourceModel

	// Read the Terraform plan.
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create our input request.
	createInput := swoClient.CreateWebsiteInput{
		Name: plan.Name.ValueString(),
		Url:  plan.Url.ValueString(),
		AvailabilityCheckSettings: &swoClient.AvailabilityCheckSettingsInput{
			CheckForString: &swoClient.CheckForStringInput{
				Operator: swoClient.CheckStringOperator(plan.Monitoring.Availability.CheckForString.Operator.ValueString()),
				Value:    plan.Monitoring.Availability.CheckForString.Value.ValueString(),
			},
			TestIntervalInSeconds: int(plan.Monitoring.Availability.TestIntervalInSeconds.ValueInt64()),
			Protocols:             convertProtocols(plan.Monitoring.Availability.Protocols),
			PlatformOptions: &swoClient.ProbePlatformOptionsInput{
				TestFromAll:    swoClient.Ptr(plan.Monitoring.Availability.PlatformOptions.TestFromAll.ValueBool()),
				ProbePlatforms: convertProbePlatforms(plan.Monitoring.Availability.PlatformOptions.Platforms),
			},
			TestFrom: swoClient.ProbeLocationInput{
				Type:   swoClient.ProbeLocationType(plan.Monitoring.Availability.TestFromLocation.ValueString()),
				Values: convertProbeLocations(plan.Monitoring.Availability.LocationOptions),
			},
			Ssl: &swoClient.SslMonitoringInput{
				Enabled:                        swoClient.Ptr(plan.Monitoring.Availability.SSL.Enabled.ValueBool()),
				DaysPriorToExpiration:          swoClient.Ptr(int(plan.Monitoring.Availability.SSL.DaysPriorToExpiration.ValueInt64())),
				IgnoreIntermediateCertificates: swoClient.Ptr(plan.Monitoring.Availability.SSL.IgnoreIntermediateCertificates.ValueBool()),
			},
			CustomHeaders: convertCustomHeaders(plan.Monitoring.CustomHeaders),
		},
		Rum: &swoClient.RumMonitoringInput{
			ApdexTimeInSeconds: swoClient.Ptr(int(plan.Monitoring.Rum.ApdexTimeInSeconds.ValueInt64())),
			Spa:                swoClient.Ptr(plan.Monitoring.Rum.Spa.ValueBool()),
		},
	}

	// Create the Website...
	newWebsite, err := r.client.WebsiteService().Create(ctx, createInput)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("error creating website '%s' - error: %s",
			plan.Name,
			err))
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("website %s created successfully - id=%s", plan.Name, newWebsite.Id))
	plan.Id = types.StringValue(newWebsite.Id)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *WebsiteResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Trace(ctx, "WebsiteResource: Read")

	var state WebsiteResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read the Website...
	tflog.Trace(ctx, fmt.Sprintf("read website with id: %s", state.Id))
	website, err := r.client.WebsiteService().Read(ctx, state.Id.ValueString())

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("error reading Website %s. error: %s",
			state.Id, err))
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("read Website success: %s", *website.Name))
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
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
			Protocols:             convertProtocols(tfPlan.Monitoring.Availability.Protocols),
			PlatformOptions: &swoClient.ProbePlatformOptionsInput{
				TestFromAll:    swoClient.Ptr(tfPlan.Monitoring.Availability.PlatformOptions.TestFromAll.ValueBool()),
				ProbePlatforms: convertProbePlatforms(tfPlan.Monitoring.Availability.PlatformOptions.Platforms),
			},
			TestFrom: swoClient.ProbeLocationInput{
				Type:   swoClient.ProbeLocationType(tfPlan.Monitoring.Availability.TestFromLocation.ValueString()),
				Values: convertProbeLocations(tfPlan.Monitoring.Availability.LocationOptions),
			},
			Ssl: &swoClient.SslMonitoringInput{
				Enabled:                        swoClient.Ptr(tfPlan.Monitoring.Availability.SSL.Enabled.ValueBool()),
				DaysPriorToExpiration:          swoClient.Ptr(int(tfPlan.Monitoring.Availability.SSL.DaysPriorToExpiration.ValueInt64())),
				IgnoreIntermediateCertificates: swoClient.Ptr(tfPlan.Monitoring.Availability.SSL.IgnoreIntermediateCertificates.ValueBool()),
			},
			CustomHeaders: convertCustomHeaders(tfPlan.Monitoring.CustomHeaders),
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
	var state WebsiteResourceModel

	// Read existing Terraform state.
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// // Delete the Website...
	tflog.Trace(ctx, fmt.Sprintf("deleting website - id=%s", state.Id))
	if err := r.client.WebsiteService().
		Delete(ctx, state.Id.ValueString()); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("error deleting website %s - %s", state.Id, err))
		return
	}
	tflog.Trace(ctx, fmt.Sprintf("website deleted: id=%s", state.Id))
}

func (r *WebsiteResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func convertProbePlatforms(in []string) []swoClient.ProbePlatform {
	var out []swoClient.ProbePlatform
	for _, p := range in {
		out = append(out, swoClient.ProbePlatform(p))
	}
	return out
}

func convertCustomHeaders(in []CustomHeader) []swoClient.CustomHeaderInput {
	var out []swoClient.CustomHeaderInput
	for _, h := range in {
		out = append(out, swoClient.CustomHeaderInput{
			Name:  h.Name.ValueString(),
			Value: h.Value.ValueString(),
		})
	}
	return out
}

func convertProbeLocations(in []ProbeLocation) []string {
	var out []string
	for _, p := range in {
		out = append(out, p.Value.ValueString())
	}
	return out
}

func convertProtocols(in []string) []swoClient.WebsiteProtocol {
	var out []swoClient.WebsiteProtocol
	for _, p := range in {
		out = append(out, swoClient.WebsiteProtocol(p))
	}
	return out
}
