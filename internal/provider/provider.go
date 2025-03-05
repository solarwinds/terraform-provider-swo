package provider

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/cenkalti/backoff/v5"
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
	_                     provider.Provider = &swoProvider{}
	ErrNonMatchingEntites                   = errors.New("updated entity properties don't match")
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

const (
	expBackoffMaxInterval = 30 * time.Second
	expBackoffMaxElapsed  = 2 * time.Minute
)

type ReadOperation[T any] func(context.Context, string) (T, error)

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
	ApiToken        types.String `tfsdk:"api_token"`
	RequestTimeout  types.Int64  `tfsdk:"request_timeout"`
	BaseURL         types.String `tfsdk:"base_url"`
	DebugMode       types.Bool   `tfsdk:"debug_mode"`
	ApiTokenEnvName types.String `tfsdk:"api_token_env_name"`
	BaseURLEnvName  types.String `tfsdk:"base_url_env_name"`
}

func (p *swoProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "swo"
	resp.Version = p.version
}

func (p *swoProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_token_env_name": schema.StringAttribute{
				Optional:    true,
				Description: "Name of the environment variable containing the API token",
			},
			"base_url_env_name": schema.StringAttribute{
				Optional:    true,
				Description: "Name of the environment variable containing the URL",
			},
			"api_token": schema.StringAttribute{
				Description: fmt.Sprintf("The api token for the %s account.", envvar.AppName),
				Optional:    true,
				Sensitive:   true,
			},
			"base_url": schema.StringAttribute{
				Description: "The base url to use for requests to the server.",
				Optional:    true,
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
	var conf swoProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &conf)...)

	if resp.Diagnostics.HasError() {
		return
	}

	config, configErr := p.ConfigureClientVars(conf, resp)

	if configErr {
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

func (p *swoProvider) ConfigureClientVars(config swoProviderModel, resp *provider.ConfigureResponse) (*swoProviderModel, bool) {
	// favor using api_token/base_url over using api_token_env_name/base_url_env_name.
	if config.ApiToken.ValueString() == "" {
		if config.ApiTokenEnvName.ValueString() != "" {
			apiTokenEnvName := config.ApiTokenEnvName.ValueString()
			apiToken, exists := os.LookupEnv(apiTokenEnvName)

			if !exists {
				resp.Diagnostics.AddAttributeError(
					path.Root("api_token_env_name"),
					"api_token_env_name doesn't exist",
					fmt.Sprintf("The api_token_env_name URL environment variable '%s' does not exist. Please provide a vilid environment variable.", apiTokenEnvName),
				)
				return nil, true
			}
			config.ApiToken = types.StringValue(apiToken)
		} else {
			resp.Diagnostics.AddAttributeError(
				path.Root("api_token"),
				"api_token required",
				"The api_token configuration parameter was not provided and is required. Please provide a public API token for the SolarWinds Observability API.",
			)
			return nil, true
		}
	}

	if config.BaseURL.ValueString() == "" {
		if config.BaseURLEnvName.ValueString() != "" {
			baseUrlEnvName := config.BaseURLEnvName.ValueString()
			baseUrl, exists := os.LookupEnv(baseUrlEnvName)

			if !exists {
				resp.Diagnostics.AddAttributeError(
					path.Root("base_url_env_name"),
					"base_url_env_name doesn't exist",
					fmt.Sprintf("The base_url_env_name environment variable %s does not exist. Please provide a vilid environment variable.", baseUrlEnvName),
				)
				return nil, true
			}
			config.BaseURL = types.StringValue(baseUrl)
		} else {
			resp.Diagnostics.AddAttributeError(
				path.Root("base_url"),
				"base_url required",
				"The base_url configuration parameter was not provided and is required. Please provide the URL of the SolarWinds Observability API.",
			)
			return nil, true
		}
	}

	return &config, false
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

func BackoffRetry[T any](operation backoff.Operation[T]) (T, error) {
	expBackoff := backoff.NewExponentialBackOff()
	expBackoff.MaxInterval = expBackoffMaxInterval

	return backoff.Retry(context.Background(), operation, backoff.WithBackOff(expBackoff), backoff.WithMaxElapsedTime(expBackoffMaxElapsed))
}

func ReadRetry[T any](ctx context.Context, id string, operation ReadOperation[T]) (T, error) {
	var entity T
	// Uri, and Website Creates and Updates are eventually consistant. Retry until the entity id is returned.
	entity, err := BackoffRetry(func() (T, error) {
		entity, err := operation(ctx, id)
		if err != nil {
			// The entity is still being created, retry
			if errors.Is(err, swoClient.ErrEntityIdNil) {
				return entity, swoClient.ErrEntityIdNil
			}

			return entity, backoff.Permanent(err)
		}

		return entity, nil
	})

	if err != nil {
		return entity, err
	}

	return entity, nil
}
