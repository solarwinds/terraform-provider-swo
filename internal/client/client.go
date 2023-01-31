package client

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/Khan/genqlient/graphql"
	log "github.com/sirupsen/logrus"
)

const (
	defaultBaseURL        = "https://api.dc-01.cloud.solarwinds.com/graphql"
	defaultMediaType      = "application/json"
	defaultRequestTimeout = 30 * time.Second
	clientIdentifier      = "swo-api-go"
)

// ServiceAccessor defines an interface for talking to via domain-specific service constructs
type ServiceAccessor interface {
	AlertsService() AlertsCommunicator
}

// Client implements ServiceAccessor
type Client struct {
	// SWO api key used for making remote requests to the SWO platform.
	apiToken string

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

// ClientOption provides functional option-setting behavior.
type ClientOption func(*Client) error

// Returns a new SWO API client. Functional option-settings.
// * BaseUrlOption
// * DebugOption
// * TransportOption
// * UserAgentOption
func NewClient(apiToken string, opts ...ClientOption) *Client {
	baseURL, err := url.Parse(defaultBaseURL)

	if err != nil {
		log.Error(err)
		return nil
	}

	swoClient := &Client{
		baseURL: baseURL,
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
	request = request.Clone(context.Background())

	request.Header.Set("Authorization", "Bearer "+t.client.apiToken)
	request.Header.Set("User-Agent", t.client.completeUserAgentString())
	// request.Header.Set("Accept", defaultMediaType)
	// request.Header.Set("Content-Type", defaultMediaType)

	if t.client.debugMode {
		dumpRequest(request)
	}

	response, err := t.client.transport.RoundTrip(request)

	if t.client.debugMode {
		dumpResponse(response)
	}

	return response, err
}

// UserAgentOption is a config function allowing setting of the User-Agent header in requests.
func UserAgentOption(userAgent string) ClientOption {
	return func(c *Client) error {
		c.userAgent = userAgent
		return nil
	}
}

// Configuation function that allows setting of the http request timeout.
func RequestTimeoutOption(duration time.Duration) ClientOption {
	return func(c *Client) error {
		c.requestTimeout = duration
		return nil
	}
}

// TransportOption is a config function allowing setting of the http.Transport.
func TransportOption(transport http.RoundTripper) ClientOption {
	return func(c *Client) error {
		c.transport = transport
		return nil
	}
}

// BaseUrlOption is a config function allowing setting of the base URL the API is targeted towards.
func BaseUrlOption(urlString string) ClientOption {
	return func(c *Client) error {
		urlObj, err := url.Parse(urlString)

		if err != nil {
			return err
		}

		c.baseURL = urlObj

		return nil
	}
}

// Sets the debug mode to on or off. Debug 'on' produces verbose logging to stdout.
func DebugOption(on bool) ClientOption {
	return func(c *Client) error {
		c.debugMode = on
		return nil
	}
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

// A debugging function which dumps the HTTP response to stdout.
func dumpResponse(resp *http.Response) {
	fmt.Printf("response status: %s\n", resp.Status)
	dump, err := httputil.DumpResponse(resp, true)

	if err != nil {
		log.Printf("error dumping response: %s", err)
		return
	}

	log.Printf("response body: %s\n\n", string(dump))
}

// A debugging function which dumps the HTTP request to stdout.
func dumpRequest(req *http.Request) {
	if req.Body == nil {
		return
	}

	dump, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		log.Printf("error dumping request: %s", err)
		return
	}

	log.Printf("request body: %s\n\n", string(dump))
}
