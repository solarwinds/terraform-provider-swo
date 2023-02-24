package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	swoClient "github.com/solarwindscloud/terraform-provider-swo/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &NotificationResource{}
var _ resource.ResourceWithConfigure = &NotificationResource{}
var _ resource.ResourceWithImportState = &NotificationResource{}

func NewNotificationResource() resource.Resource {
	return &NotificationResource{}
}

// Defines the resource implementation.
type NotificationResource struct {
	client *swoClient.Client
}

func (r *NotificationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_notification"
}

func (r *NotificationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	tflog.Trace(ctx, "NotificationResource: Configure")

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

func (r *NotificationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Trace(ctx, "NotificationResource: Create")

	var plan NotificationResourceModel

	// Read the Terraform plan.
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	desc := plan.Description.ValueString()

	// Create the notification...
	newNotification, err := r.client.
		NotificationsService().
		Create(ctx, swoClient.CreateNotificationInput{
			Title:       plan.Title.ValueString(),
			Description: &desc,
			Type:        plan.Type.ValueString(),
			Settings:    plan.GetSettings(),
		})

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("error creating notification '%s'. error: %s",
			plan.Title,
			err))
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("notification %s created successfully. id: %s", newNotification.Title, newNotification.Id))

	// Update the computed values from the response. We need to set the Id to a combination of the
	// notification Id and the notification Type because the server requires both values when asking
	// for the data. See link for more details on why this is necessary in Terraform.
	// https://developer.hashicorp.com/terraform/plugin/framework/resources/import#multiple-attributes
	plan.Id = types.StringValue(fmt.Sprintf("%s:%s", newNotification.Id, newNotification.Type))
	plan.CreatedAt = types.StringValue(newNotification.CreatedAt.String())
	plan.CreatedBy = types.StringValue(newNotification.CreatedBy)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *NotificationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Trace(ctx, "NotificationResource: Read")

	var state NotificationResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	nId, nType, err := state.ParseId()
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("error parsing notification id. got: %s. error: %s",
			state.Id, err))
		return
	}

	// Read the notification...
	tflog.Trace(ctx, fmt.Sprintf("read notification with id: %s", nId))
	notification, err := r.client.
		NotificationsService().
		Read(ctx, nId, nType)

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("error reading notification %s. error: %s",
			nId, err))
		return
	}

	state.Id = types.StringValue(fmt.Sprintf("%s:%s", notification.Id, notification.Type))
	state.Title = types.StringValue(notification.Title)
	state.Type = types.StringValue(notification.Type)
	state.Description = types.StringValue(*notification.Description)
	state.CreatedAt = types.StringValue(notification.CreatedAt.String())
	state.CreatedBy = types.StringValue(notification.CreatedBy)

	err = state.SetSettings(notification.Settings)
	if err != nil {
		resp.Diagnostics.AddError("Settings Error", err.Error())
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("read notification success: %s", notification.Id))

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *NotificationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state NotificationResourceModel

	// Read the Terraform plan.
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	// The plan doesn't capture any existing computed values so we can take it from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	nId, _, err := state.ParseId()
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("error parsing notification id. got %s. err: %s", state.Id, err))
		return
	}

	title := plan.Title.ValueString()
	desc := plan.Description.ValueString()
	settings := plan.GetSettings()

	// Update the notification...
	tflog.Trace(ctx, fmt.Sprintf("updating notification with id: %s", nId))
	err = r.client.
		NotificationsService().
		Update(ctx, swoClient.UpdateNotificationInput{
			Id:          nId,
			Title:       &title,
			Description: &desc,
			Settings:    &settings,
		})

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("error updating notification %s. err: %s", nId, err))
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("notification '%s' updated successfully.", nId))

	plan.Id = state.Id
	plan.CreatedAt = state.CreatedAt
	plan.CreatedBy = state.CreatedBy

	// Save to Terraform state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *NotificationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state NotificationResourceModel

	// Read existing Terraform state.
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	nId, _, err := state.ParseId()
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("error deleting notification %s. err: %s", nId, err))
		return
	}

	// Delete the notification...
	tflog.Trace(ctx, fmt.Sprintf("deleting notification. id: %s", nId))
	err = r.client.
		NotificationsService().
		Delete(ctx, nId)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("error deleting notification %s. err: %s", nId, err))
		return
	}
	tflog.Trace(ctx, fmt.Sprintf("notification deleted. id: %s", nId))
}

func (r *NotificationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
