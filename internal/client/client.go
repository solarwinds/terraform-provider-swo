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
)

const (
	defaultBaseURL        = "https://api.dc-01.cloud.solarwinds.com/graphql"
	defaultMediaType      = "application/json"
	defaultRequestTimeout = 30 * time.Second
	clientIdentifier      = "swo-api-go"
	requestIdentifier     = "x-request-id"
)

// ServiceAccessor defines an interface for talking to via domain-specific service constructs
type ServiceAccessor interface {
	AlertsService() AlertsCommunicator
	NotificationsService() NotificationsCommunicator
}

// Client implements ServiceAccessor
type Client struct {
	// Option settings
	baseURL        *url.URL
	debugMode      bool
	requestTimeout time.Duration
	userAgent      string
	transport      http.RoundTripper

	// GraphQL client
	gql graphql.Client

	// Service accessors
	alertsService        AlertsCommunicator
	notificationsService NotificationsCommunicator
}

// apiTokentAuthTransport is an http.RoundTripper that authenticates all requests
// using Bearer token authentication with a SWO api key.
type apiTokenAuthTransport struct {
	client   *Client
	apiToken string
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
		baseURL:        baseURL,
		requestTimeout: defaultRequestTimeout,
	}

	// Set any user options that were provided.
	for _, opt := range opts {
		err = opt(swoClient)
		if err != nil {
			log.Error(fmt.Errorf("client option error. fallback to default value: %s", err))
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
			apiToken: apiToken,
			client:   swoClient,
		},
	})

	swoClient.alertsService = NewAlertsService(swoClient)
	swoClient.notificationsService = NewNotificationsService(swoClient)

	return swoClient
}

// RoundTrip implements the http.RoundTrip interface.
func (t *apiTokenAuthTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	// When using the RoundTipper interface it is important that the original request is not modified.
	// See https://pkg.go.dev/net/http#RoundTripper for more details.
	clone := request.Clone(context.Background())

	clone.Header.Set("Authorization", "Bearer "+t.apiToken)
	clone.Header.Set("User-Agent", t.client.completeUserAgentString())
	clone.Header.Set(requestIdentifier, uuid.NewString())

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

// A subset of the API that deals with Notifications.
func (c *Client) NotificationsService() NotificationsCommunicator {
	return c.notificationsService
}

// Returns the string that will be placed in the User-Agent header. It ensures
// that any caller-set string has the client name and version appended to it.
func (c *Client) completeUserAgentString() string {
	if c.userAgent == "" {
		return clientIdentifier
	}
	return fmt.Sprintf("%s:%s", c.userAgent, clientIdentifier)
}
