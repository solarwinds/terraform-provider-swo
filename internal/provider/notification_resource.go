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

	// Read the Terraform plan data into the model and log the results.
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create the notification from the provided Terraform model...
	newNotification, err := r.client.
		NotificationsService().
		Create(&swoClient.CreateNotificationInput{
			Title:       plan.Title,
			Description: plan.Description,
			Type:        plan.Type,
			Settings:    plan.GetSettings(),
		})

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("error creating notification '%s'. error: %s",
			plan.Title,
			err))
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("notification '%s' created successfully. id: %s", newNotification.Title, newNotification.Id))

	// Update the computed values from the response.
	plan.Id = types.StringValue(newNotification.Id)
	plan.CreatedAt = types.StringValue(newNotification.CreatedAt.String())
	plan.CreatedBy = types.StringValue(newNotification.CreatedBy)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *NotificationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Trace(ctx, "NotificationResource: Read")

	var plan NotificationResourceModel

	// Read any existing Terraform state into the model.
	resp.Diagnostics.Append(req.State.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("Getting Notification with ID: %s", plan.Id))
	notification, err := r.client.NotificationsService().Read(plan.Id.ValueString(), plan.Type)

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("error reading notification %s. error: %s",
			plan.Id,
			err))
		return
	}

	plan.Id = types.StringValue(notification.Id)
	plan.Title = notification.Title
	plan.Type = notification.Type
	plan.Description = notification.Description
	plan.CreatedAt = types.StringValue(notification.CreatedAt.String())
	plan.CreatedBy = types.StringValue(notification.CreatedBy)

	err = plan.SetSettings(notification.Settings)
	if err != nil {
		resp.Diagnostics.AddError("Settings Error", err.Error())
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("notification received: %s", notification.Title))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *NotificationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Trace(ctx, "NotificationResource: Update")

	var plan, state NotificationResourceModel

	// Read the Terraform plan data into the plan.
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// The plan doesn't seem to capture the existing Id. We can take it from state here.
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.Id = state.Id
	plan.CreatedAt = state.CreatedAt
	plan.CreatedBy = state.CreatedBy

	settings := plan.GetSettings()

	update := &swoClient.UpdateNotificationInput{
		Id:          plan.Id.ValueString(),
		Title:       &plan.Title,
		Description: plan.Description,
		Settings:    &settings,
	}

	tflog.Trace(ctx, fmt.Sprintf("updating notification with id: %s", update.Id))

	// Update the notification...
	err := r.client.NotificationsService().Update(update)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("error updating notification %s. err: %s", plan.Id, err))
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("notification '%s' updated successfully.", plan.Title))

	// Save and log the model into Terraform state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *NotificationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Trace(ctx, "NotificationResource: Delete")
	var plan NotificationResourceModel

	// Read Terraform prior state data into the plan.
	resp.Diagnostics.Append(req.State.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	planId := plan.Id.ValueString()
	tflog.Trace(ctx, fmt.Sprintf("deleting notification. id: %s", planId))

	// Delete the notification...
	err := r.client.NotificationsService().Delete(planId)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("error deleting notification %s. err: %s", planId, err))
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("notification deleted. id: %s", planId))
}

func (r *NotificationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tflog.Trace(ctx, "NotificationResource: ImportState")
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
