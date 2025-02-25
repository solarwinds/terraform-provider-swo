package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var (
	_ resource.ResourceWithConfigure        = &resourceWrapper{}
	_ resource.ResourceWithImportState      = &resourceWrapper{}
	_ resource.ResourceWithConfigValidators = &resourceWrapper{}
	_ resource.ResourceWithModifyPlan       = &resourceWrapper{}
	_ resource.ResourceWithUpgradeState     = &resourceWrapper{}
	_ resource.ResourceWithValidateConfig   = &resourceWrapper{}
)

func newResourceWrapper(i *resource.Resource) resource.Resource {
	return &resourceWrapper{
		innerResource: i,
	}
}

type resourceWrapper struct {
	innerResource *resource.Resource
}

func (r *resourceWrapper) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	rCasted, ok := (*r.innerResource).(resource.ResourceWithConfigure)
	if ok {
		if req.ProviderData == nil {
			return
		}
		_, ok := req.ProviderData.(providerClients)

		if !ok {
			resp.Diagnostics.AddError("Unexpected Resource Configure Type", "")
			return
		}

		rCasted.Configure(ctx, req, resp)
	}
}

func (r *resourceWrapper) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	(*r.innerResource).Metadata(ctx, req, resp)
	resp.TypeName = fmt.Sprintf("%s_%s", req.ProviderTypeName, resp.TypeName)
}

func (r *resourceWrapper) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	(*r.innerResource).Schema(ctx, req, resp)
	enrichSchema(&resp.Schema)
}

func (r *resourceWrapper) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	(*r.innerResource).Create(ctx, req, resp)
}

func (r *resourceWrapper) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	(*r.innerResource).Read(ctx, req, resp)
}

func (r *resourceWrapper) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	(*r.innerResource).Update(ctx, req, resp)
}

func (r *resourceWrapper) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	(*r.innerResource).Delete(ctx, req, resp)
}

func (r *resourceWrapper) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if rCasted, ok := (*r.innerResource).(resource.ResourceWithImportState); ok {
		rCasted.ImportState(ctx, req, resp)
		return
	}

	resp.Diagnostics.AddError(
		"Resource Import Not Implemented",
		"This resource does not support import. Please contact the provider developer for additional information.",
	)
}

func (r *resourceWrapper) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	if rCasted, ok := (*r.innerResource).(resource.ResourceWithConfigValidators); ok {
		return rCasted.ConfigValidators(ctx)
	}
	return nil
}

func (r *resourceWrapper) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if v, ok := (*r.innerResource).(resource.ResourceWithModifyPlan); ok {
		v.ModifyPlan(ctx, req, resp)
	}
}

func (r *resourceWrapper) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	if v, ok := (*r.innerResource).(resource.ResourceWithUpgradeState); ok {
		return v.UpgradeState(ctx)
	}
	return nil
}

func (r *resourceWrapper) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	if v, ok := (*r.innerResource).(resource.ResourceWithValidateConfig); ok {
		v.ValidateConfig(ctx, req, resp)
	}
}
