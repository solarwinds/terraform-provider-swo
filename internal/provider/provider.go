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
	"github.com/solarwinds/swo-sdk-go/swov1"
	"github.com/solarwinds/terraform-provider-swo/internal/envvar"
)

var (
	dataSources            []func() datasource.DataSource
	ErrNonMatchingEntities = errors.New("updated entity properties don't match")
	ErrMarshal             = errors.New("error during marshalling")
)

var resources = []func() resource.Resource{
	NewAlertResource,
	NewApiTokenResource,
	NewCompositeMetricResource,
	NewDashboardResource,
	NewLogFilterResource,
	NewNotificationResource,
	NewUriResource,
	NewWebsiteResource,
}

const (
	expBackoffMaxInterval = 30 * time.Second
	expBackoffMaxElapsed  = 2 * time.Minute

	// #nosec G101: Potential hardcoded credentials
	apiTokenEnv = "SWO_API_TOKEN"
	// #nosec G101: Potential hardcoded credentials
	baseUrlEnv = "SWO_BASE_URL"
)

type ReadOperation[T any] func(context.Context, string) (T, error)

// swoProvider defines the provider implementation.
type swoProvider struct {
	// Version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string

	transport http.RoundTripper
}

// swoProviderModel describes the provider data model.
type swoProviderModel struct {
	ApiToken       types.String `tfsdk:"api_token"`
	RequestTimeout types.Int64  `tfsdk:"request_timeout"`
	BaseURL        types.String `tfsdk:"base_url"`
	DebugMode      types.Bool   `tfsdk:"debug_mode"`
}

type providerClients struct {
	SwoClient   *swoClient.Client
	SwoV1Client *swov1.Swo
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
	var model swoProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	config, hasErr := p.ConfigureClientVars(model, resp)

	if hasErr {
		return
	}

	clientTimeout := time.Duration(config.RequestTimeout.ValueInt64()) * time.Second

	// Client configuration for data sources and resources.
	client, err := swoClient.New(config.ApiToken.ValueString(),
		swoClient.RequestTimeoutOption(clientTimeout),
		swoClient.BaseUrlOption(config.BaseURL.ValueString()),
		swoClient.TransportOption(p.transport),
		swoClient.DebugOption(config.DebugMode.ValueBool()),
	)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("error creating client: %s", err))
		return
	}

	baseUrl, err := stripURLToDomain(config.BaseURL.ValueString())

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("invalid base url: %s err: %s", baseUrl, err))
		return
	}

	swoV1Client := swov1.New(
		swov1.WithServerURL(baseUrl),
		swov1.WithSecurity(config.ApiToken.ValueString()),
		swov1.WithClient(&http.Client{Timeout: clientTimeout}),
	)

	providerClients := providerClients{
		SwoClient:   client,
		SwoV1Client: swoV1Client,
	}

	resp.DataSourceData = providerClients
	resp.ResourceData = providerClients
}

func (p *swoProvider) ConfigureClientVars(config swoProviderModel, resp *provider.ConfigureResponse) (*swoProviderModel, bool) {
	if config.ApiToken.ValueString() == "" {
		apiToken, exists := os.LookupEnv(apiTokenEnv)

		if !exists {
			resp.Diagnostics.AddAttributeError(
				path.Root("api_token"),
				"No API token provided",
				fmt.Sprintf("Please set either the 'api_token' parameter or the '%s' environment variable.", apiTokenEnv),
			)
			return nil, true
		}

		config.ApiToken = types.StringValue(apiToken)
	}

	if config.BaseURL.ValueString() == "" {
		baseUrl, exists := os.LookupEnv(baseUrlEnv)

		if !exists {
			resp.Diagnostics.AddAttributeError(
				path.Root("api_token"),
				"No Base URL provided",
				fmt.Sprintf("Please set either the 'base_url' parameter or the '%s' environment variable.", baseUrlEnv),
			)
			return nil, true
		}

		config.BaseURL = types.StringValue(baseUrl)
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

func BackoffRetry[T any](ctx context.Context, operation backoff.Operation[T]) (T, error) {
	expBackoff := backoff.NewExponentialBackOff()
	expBackoff.MaxInterval = expBackoffMaxInterval

	return backoff.Retry(ctx, operation, backoff.WithBackOff(expBackoff), backoff.WithMaxElapsedTime(expBackoffMaxElapsed))
}

func ReadRetry[T any](ctx context.Context, id string, operation ReadOperation[T]) (T, error) {
	var entity T
	// Uri, and Website Creates and Updates are eventually consistent. Retry until the entity id is returned.
	entity, err := BackoffRetry(ctx, func() (T, error) {
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
