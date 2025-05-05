package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	swoClient "github.com/solarwinds/swo-client-go/pkg/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &logFilterResource{}
	_ resource.ResourceWithConfigure   = &logFilterResource{}
	_ resource.ResourceWithImportState = &logFilterResource{}
)

func NewLogFilterResource() resource.Resource {
	return &logFilterResource{}
}

// Defines the resource implementation.
type logFilterResource struct {
	client *swoClient.Client
}

func (r *logFilterResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "logfilter"
}

func (r *logFilterResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, _ := req.ProviderData.(providerClients)
	r.client = client.SwoClient
}

func (r *logFilterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var tfPlan logFilterResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &tfPlan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var expressions []logFilterExpression
	tfPlan.Expressions.ElementsAs(ctx, &expressions, false)

	// Create our input request.
	createInput := swoClient.CreateExclusionFilterInput{
		Name:           tfPlan.Name.ValueString(),
		Description:    tfPlan.Description.ValueString(),
		TokenSignature: tfPlan.TokenSignature.ValueStringPointer(),
		Expressions: convertArray(expressions, func(e logFilterExpression) swoClient.CreateExclusionFilterExpressionInput {
			return swoClient.CreateExclusionFilterExpressionInput{
				Kind:       swoClient.ExclusionFilterExpressionKind(e.Kind.ValueString()),
				Expression: e.Expression.ValueString(),
			}
		}),
	}

	// Create the LogFilter
	result, err := r.client.LogFilterService().Create(ctx, createInput)
	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error creating logFilter '%s' - error: %s", tfPlan.Name, err))
		return
	}

	tfPlan.Id = types.StringValue(result.Id)
	resp.Diagnostics.Append(resp.State.Set(ctx, tfPlan)...)
}

func (r *logFilterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var tfState logFilterResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &tfState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read the LogFilter...
	logFilter, err := r.client.LogFilterService().Read(ctx, tfState.Id.ValueString())

	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error reading logFilter %s. error: %s", tfState.Id, err))
		return
	} else if logFilter == nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("logFilter not found. id=%s", tfState.Id))
		return
	}

	// Update the Terraform state with latest values from the server.
	tfState.Name = types.StringValue(logFilter.Name)
	tfState.Description = types.StringValue(*logFilter.Description)
	tfState.TokenSignature = types.StringValue(*logFilter.TokenSignature)

	var diags diag.Diagnostics
	var elements []attr.Value
	var attributeTypes = ExpressionAttributeTypes()
	for _, p := range logFilter.Expressions {
		objectValue, objectDiags := types.ObjectValueFrom(
			ctx,
			attributeTypes,
			logFilterExpression{
				Kind:       types.StringValue(string(p.Kind)),
				Expression: types.StringValue(p.Expression),
			},
		)
		elements = append(elements, objectValue)
		diags = append(diags, objectDiags...)
	}
	expressions, setDiags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: attributeTypes}, elements)
	diags = append(diags, setDiags...)
	tfState.Expressions = expressions

	// Save to Terraform state.
	resp.Diagnostics.Append(resp.State.Set(ctx, tfState)...)
}

func (r *logFilterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var tfPlan, tfState *logFilterResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &tfPlan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &tfState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var expressions []logFilterExpression
	tfPlan.Expressions.ElementsAs(ctx, &expressions, false)

	err := r.client.LogFilterService().Update(ctx, swoClient.UpdateExclusionFilterInput{
		Id:          tfState.Id.ValueString(),
		Name:        tfPlan.Name.ValueString(),
		Description: tfPlan.Description.ValueString(),
		Expressions: convertArray(expressions, func(e logFilterExpression) swoClient.UpdateExclusionFilterExpressionInput {
			return swoClient.UpdateExclusionFilterExpressionInput{
				Kind:       swoClient.ExclusionFilterExpressionKind(e.Kind.ValueString()),
				Expression: e.Expression.ValueString(),
			}
		}),
	})

	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error updating logFilter %s. err: %s", tfState.Id, err))
		return
	}

	// Save to Terraform state.
	tfPlan.Id = tfState.Id
	resp.Diagnostics.Append(resp.State.Set(ctx, &tfPlan)...)
}

func (r *logFilterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var tfState logFilterResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &tfState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the LogFilter...
	if err := r.client.LogFilterService().Delete(ctx, tfState.Id.ValueString()); err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error deleting logFilter %s - %s", tfState.Id, err))
	}
}

func (r *logFilterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
