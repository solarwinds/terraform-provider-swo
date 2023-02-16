package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	swoClient "github.com/solarwindscloud/terraform-provider-swo/internal/client"
	"github.com/solarwindscloud/terraform-provider-swo/internal/envvar"
)

// Ensure SwoProvider satisfies various provider interfaces.
var _ provider.Provider = &SwoProvider{}

// SwoProvider defines the provider implementation.
type SwoProvider struct {
	// Version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// SwoProviderModel describes the provider data model.
type SwoProviderModel struct {
	ApiToken       types.String `tfsdk:"api_token"`
	RequestTimeout types.Int64  `tfsdk:"request_timeout"`
	DebugMode      types.Bool   `tfsdk:"debug_mode"`
}

func (p *SwoProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	tflog.Trace(ctx, "SWO Provider: Metadata")

	resp.TypeName = "swo"
	resp.Version = p.version
}

func (p *SwoProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	tflog.Trace(ctx, "SWO Provider: Schema")

	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_token": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: fmt.Sprintf("The authentication token for the %s account.", envvar.AppName),
			},
			"request_timeout": schema.Int64Attribute{
				Optional:    true,
				Description: "The request timeout period in seconds. Default is 30 seconds.",
			},
			"debug_mode": schema.BoolAttribute{
				Optional:    true,
				Description: "Setting to true will provide additional logging details.",
			},
		},
	}
}

func (p *SwoProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "SWO Provider: Configure")

	var config SwoProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if config.ApiToken.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Api Token Required",
			"The api token was not provided.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Client configuration for data sources and resources.
	client := swoClient.NewClient(config.ApiToken.ValueString(),
		swoClient.RequestTimeoutOption(time.Duration(config.RequestTimeout.ValueInt64())*time.Second),
		swoClient.DebugOption(config.DebugMode.ValueBool()),
	)

	if client == nil {
		resp.Diagnostics.AddError("Client Error", "Unable to create an instance of the SWO client.")
		return
	}

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *SwoProvider) Resources(ctx context.Context) []func() resource.Resource {
	tflog.Trace(ctx, "SWO Provider: Resources")

	return []func() resource.Resource{
		NewAlertResource,
		NewNotificationResource,
	}
}

func (p *SwoProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	tflog.Trace(ctx, "SWO Provider: DataSources")

	return []func() datasource.DataSource{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &SwoProvider{
			version: version,
		}
	}
}
