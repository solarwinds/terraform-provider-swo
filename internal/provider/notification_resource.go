package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	swoClient "github.com/solarwinds/swo-client-go/pkg/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &notificationResource{}
	_ resource.ResourceWithConfigure   = &notificationResource{}
	_ resource.ResourceWithImportState = &notificationResource{}
)

func NewNotificationResource() resource.Resource {
	return &notificationResource{}
}

// Defines the resource implementation.
type notificationResource struct {
	client *swoClient.Client
}

func (r *notificationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "notification"
}

func (r *notificationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, _ := req.ProviderData.(providerClients)
	r.client = client.SwoClient
}

func (r *notificationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var tfPlan notificationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &tfPlan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create the notification...
	tfSettings := tfPlan.GetSettings(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	input := swoClient.CreateNotificationInput{
		Title:       tfPlan.Title.ValueString(),
		Description: tfPlan.Description.ValueStringPointer(),
		Type:        tfPlan.Type.ValueString(),
		Settings:    tfSettings,
	}
	newNotification, err := r.client.NotificationsService().Create(ctx, input)

	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error creating notification '%s'. error: %s", tfPlan.Title, err))
		return
	}

	// Update the computed values from the response. We need to set the Id to a combination of the
	// notification Id and the notification Type because the server requires both values when asking
	// for the data. See link for more details on why this is necessary in Terraform.
	// https://developer.hashicorp.com/terraform/plugin/framework/resources/import#multiple-attributes
	tfPlan.Id = types.StringValue(fmt.Sprintf("%s:%s", newNotification.Id, newNotification.Type))
	resp.Diagnostics.Append(resp.State.Set(ctx, tfPlan)...)
}

func (r *notificationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var tfState notificationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &tfState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	nId, nType, err := ParseNotificationId(tfState.Id)
	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error parsing notification id. got: %s. error: %s", tfState.Id, err))
		return
	}

	// Read the notification...
	notification, err := r.client.NotificationsService().Read(ctx, nId, nType)

	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error reading notification %s. error: %s", nId, err))
		return
	}

	tfState.Id = types.StringValue(fmt.Sprintf("%s:%s", notification.Id, notification.Type))
	tfState.Title = types.StringValue(notification.Title)
	tfState.Type = types.StringValue(notification.Type)
	tfState.Description = types.StringPointerValue(notification.Description)
	err = tfState.SetSettings(notification.Settings, ctx)
	if err != nil {
		resp.Diagnostics.AddError("Settings Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, tfState)...)
}

func (r *notificationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var tfPlan, tfState notificationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &tfPlan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &tfState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	nId, _, err := ParseNotificationId(tfState.Id)
	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error parsing notification id. got %s. err: %s", tfState.Id, err))
		return
	}

	tfSettings := tfPlan.GetSettings(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	// Update the notification...
	_, err = r.client.NotificationsService().Update(ctx,
		swoClient.UpdateNotificationInput{
			Id:          nId,
			Title:       tfPlan.Title.ValueStringPointer(),
			Description: tfPlan.Description.ValueStringPointer(),
			Settings:    &tfSettings,
		})

	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error updating notification %s. err: %s", nId, err))
		return
	}

	// Save to Terraform state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &tfPlan)...)
}

func (r *notificationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var tfState notificationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &tfState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	nId, _, err := ParseNotificationId(tfState.Id)
	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error deleting notification %s. err: %s", nId, err))
		return
	}

	// Delete the notification...
	err = r.client.NotificationsService().Delete(ctx, nId)
	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error deleting notification %s. err: %s", nId, err))
	}
}

func (r *notificationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
