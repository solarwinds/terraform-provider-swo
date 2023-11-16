package provider

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	swoClient "github.com/solarwindscloud/swo-client-go/pkg/client"
	"github.com/solarwindscloud/terraform-provider-swo/internal/envvar"
)

// Ensure SwoProvider satisfies various provider interfaces.
var _ provider.Provider = &SwoProvider{}

// SwoProvider defines the provider implementation.
type SwoProvider struct {
	// Version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version   string
	transport http.RoundTripper
}

// SwoProviderModel describes the provider data model.
type SwoProviderModel struct {
	ApiToken       types.String `tfsdk:"api_token"`
	RequestTimeout types.Int64  `tfsdk:"request_timeout"`
	BaseURL        types.String `tfsdk:"base_url"`
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
			"base_url": schema.StringAttribute{
				Optional:    true,
				Description: "The base url to use for requests to the server.",
			},
			"debug_mode": schema.BoolAttribute{
				Optional:    true,
				Description: "Setting to true will provide additional logging details.",
			},
		},
	}
}

func (p *SwoProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config SwoProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if config.ApiToken.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_token"),
			"Api Token Required",
			"The api token was not provided.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Client configuration for data sources and resources.
	client, err := swoClient.New(config.ApiToken.ValueString(),
		swoClient.RequestTimeoutOption(time.Duration(config.RequestTimeout.ValueInt64())*time.Second),
		swoClient.BaseUrlOption(config.BaseURL.ValueString()),
		swoClient.TransportOption(p.transport),
		swoClient.DebugOption(config.DebugMode.ValueBool()),
	)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("error creating client: %s", err))
		return
	}

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *SwoProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewAlertResource,
		NewDashboardResource,
		NewNotificationResource,
		NewWebsiteResource,
		NewUriResource,
	}
}

func (p *SwoProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func New(version string, transport http.RoundTripper) func() provider.Provider {
	return func() provider.Provider {
		return &SwoProvider{
			version:   version,
			transport: transport,
		}
	}
}
