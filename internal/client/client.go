package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/Khan/genqlient/graphql"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/solarwindscloud/swo-session-creator-go/session"
)

const (
	defaultBaseURL        = "https://api.dc-01.cloud.solarwinds.com/graphql"
	defaultMediaType      = "application/json"
	defaultRequestTimeout = 30 * time.Second
	clientIdentifier      = "swo-api-go"
	requestIdentifier     = "swo-request-id"
)

// ServiceAccessor defines an interface for talking to via domain-specific service constructs
type ServiceAccessor interface {
	AlertsService() AlertsCommunicator
}

// Client implements ServiceAccessor
type Client struct {
	// SWO api key used for making remote requests to the SWO platform.
	apiToken string
	// UserSession is used to make requests to the non-public GQL Gateway.
	// Used when creating a new private client.
	userSession *session.UserSession

	// Option settings
	baseURL        *url.URL
	debugMode      bool
	requestTimeout time.Duration
	userAgent      string
	transport      http.RoundTripper

	// GraphQL client
	gql graphql.Client

	// Service accessors
	alertsService AlertsCommunicator
}

// apiTokentAuthTransport is an http.RoundTripper that authenticates all requests
// using Bearer token authentication with a SWO api key.
type apiTokenAuthTransport struct {
	client *Client // SWO client object.
}

// Returns a new SWO API client with functional override options.
// * BaseUrlOption
// * DebugOption
// * TransportOption
// * UserAgentOption
// * RequestTimeoutOption
func NewClient(apiToken string, opts ...ClientOption) *Client {
	baseURL, err := url.Parse(defaultBaseURL)

	if err != nil {
		log.Error(err)
		return nil
	}

	swoClient := &Client{
		apiToken: apiToken,
		baseURL:  baseURL,
	}

	// Set any user options that were provided.
	for _, opt := range opts {
		err = opt(swoClient)
		if err != nil {
			log.Error(fmt.Sprintf("Client option error. Fallback to default value: %s", err))
		}
	}

	// Use the default http transport if one wasn't provided.
	if swoClient.transport == nil {
		swoClient.transport = http.DefaultTransport
	}

	if swoClient.debugMode {
		log.SetLevel(log.TraceLevel)
		log.Info("DebugMode set to true.")
	}

	swoClient.gql = graphql.NewClient(swoClient.baseURL.String(), &http.Client{
		Timeout: swoClient.requestTimeout,
		Transport: &apiTokenAuthTransport{
			client: swoClient,
		},
	})

	swoClient.alertsService = NewAlertsService(swoClient)

	return swoClient
}

// RoundTrip implements the http.RoundTrip interface.
func (t *apiTokenAuthTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	// When using the RoundTipper interface it is important that the original request is not modified.
	// See https://pkg.go.dev/net/http#RoundTripper for more details.
	clone := request.Clone(context.Background())

	clone.Header.Set("Authorization", "Bearer "+t.client.apiToken)
	clone.Header.Set("User-Agent", t.client.completeUserAgentString())
	clone.Header.Set(requestIdentifier, uuid.NewString())
	// request.Header.Set("Accept", defaultMediaType)
	// request.Header.Set("Content-Type", defaultMediaType)

	if t.client.debugMode {
		dumpRequest(clone)
	}

	response, err := t.client.transport.RoundTrip(clone)

	if t.client.debugMode {
		dumpResponse(response)
	}

	return response, err
}

// A subset of the API that deals with Alerts.
func (c *Client) AlertsService() AlertsCommunicator {
	return c.alertsService
}

// Returns the string that will be placed in the User-Agent header. It ensures
// that any caller-set string has the client name and version appended to it.
func (c *Client) completeUserAgentString() string {
	if c.userAgent == "" {
		return clientIdentifier
	}
	return fmt.Sprintf("%s:%s", c.userAgent, clientIdentifier)
}
