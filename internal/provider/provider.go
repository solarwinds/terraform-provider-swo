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
	swoClient "github.com/solarwinds/swo-client-go/pkg/client"
	"github.com/solarwinds/terraform-provider-swo/internal/envvar"
)

// Ensure SwoProvider satisfies various provider interfaces.
var (
	_ provider.Provider = &swoProvider{}
)

var resources = []func() resource.Resource{
	NewAlertResource,
	NewApiTokenResource,
	NewLogFilterResource,
	NewDashboardResource,
	NewNotificationResource,
	NewWebsiteResource,
	NewUriResource,
}

var dataSources = []func() datasource.DataSource{}

// swoProvider defines the provider implementation.
type swoProvider struct {
	// Version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version   string
	transport http.RoundTripper
}

// swoProviderModel describes the provider data model.
type swoProviderModel struct {
	ApiToken       types.String `tfsdk:"api_token"`
	RequestTimeout types.Int64  `tfsdk:"request_timeout"`
	BaseURL        types.String `tfsdk:"base_url"`
	DebugMode      types.Bool   `tfsdk:"debug_mode"`
}

func (p *swoProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "swo"
	resp.Version = p.version
}

func (p *swoProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_token": schema.StringAttribute{
				Description: fmt.Sprintf("The api token for the %s account.", envvar.AppName),
				Required:    true,
				Sensitive:   true,
			},
			"base_url": schema.StringAttribute{
				Description: "The base url to use for requests to the server.",
				Required:    true,
			},
			"request_timeout": schema.Int64Attribute{
				Description: "The request timeout period in seconds. Default is 30 seconds.",
				Optional:    true,
			},
			"debug_mode": schema.BoolAttribute{
				Description: "Setting to true will provide additional logging details.",
				Optional:    true,
			},
		},
	}
}

func (p *swoProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config swoProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if config.ApiToken.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_token"),
			"api_token required",
			"The api_token configuration parameter was not provided and is required. Please provide a public API token for the SolarWinds Observability API.",
		)
		return
	}

	if config.BaseURL.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("base_url"),
			"base_url required",
			"The base_url configuration parameter was not provided and is required. Please provide the URL of the SolarWinds Observability API.",
		)
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

func (p *swoProvider) Resources(ctx context.Context) []func() resource.Resource {
	var wrappedResources []func() resource.Resource
	for _, f := range resources {
		r := f()
		wrappedResources = append(wrappedResources, func() resource.Resource { return newResourceWrapper(&r) })
	}

	return wrappedResources
}

func (p *swoProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return dataSources
}

func New(version string, transport http.RoundTripper) func() provider.Provider {
	return func() provider.Provider {
		return &swoProvider{
			version:   version,
			transport: transport,
		}
	}
}
