package client

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/Khan/genqlient/graphql"
	log "github.com/sirupsen/logrus"
)

const (
	defaultBaseURL        = "https://api.dc-01.cloud.solarwinds.com/graphql"
	defaultMediaType      = "application/json"
	defaultRequestTimeout = 30 * time.Second
	clientIdentifier      = "Swo-Api-Go"
	requestIdentifier     = "X-Request-Id"
)

var (
	serviceInitError = func(serviceName string) error {
		return fmt.Errorf("could not instantiate service. name: %s", serviceName)
	}
)

// ServiceAccessor defines an interface for talking to via domain-specific service constructs
type ServiceAccessor interface {
	AlertsService() AlertsCommunicator
	DashboardsService() DashboardsCommunicator
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
	dashboardsService    DashboardsCommunicator
	notificationsService NotificationsCommunicator
}

// Each service derives from the service type.
type service struct {
	client *Client
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

	// Use the api token transport if one wasn't provided.
	if swoClient.transport == nil {
		swoClient.transport = &apiTokenAuthTransport{
			apiToken: apiToken,
			client:   swoClient,
		}
	}

	if swoClient.debugMode {
		log.SetLevel(log.TraceLevel)
		log.Info("swoclient: debugMode set to true.")
	}

	swoClient.gql = graphql.NewClient(swoClient.baseURL.String(), &http.Client{
		Timeout:   swoClient.requestTimeout,
		Transport: swoClient.transport,
	})

	if err = initServices(swoClient); err != nil {
		log.Fatal(err)
	}

	return swoClient
}

func initServices(c *Client) error {
	if c.alertsService = NewAlertsService(c); c.alertsService == nil {
		return serviceInitError("AlertsService")
	}
	if c.dashboardsService = NewDashboardsService(c); c.dashboardsService == nil {
		return serviceInitError("DashboardsService")
	}
	if c.notificationsService = NewNotificationsService(c); c.notificationsService == nil {
		return serviceInitError("NotificationsService")
	}

	return nil
}

// A subset of the API that deals with Alerts.
func (c *Client) AlertsService() AlertsCommunicator {
	return c.alertsService
}

// A subset of the API that deals with Dashboards.
func (c *Client) DashboardsService() DashboardsCommunicator {
	return c.dashboardsService
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
