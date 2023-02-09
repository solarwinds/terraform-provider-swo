package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/Khan/genqlient/graphql"
	log "github.com/sirupsen/logrus"
	"github.com/solarwindscloud/swo-session-creator-go/session"
)

const (
	defaultGatewayUrl = "https://my.na-01.cloud.solarwinds.com/common/graphql"
)

type userSessionAuthTransport struct {
	client *Client // SWO client object.
}

//GQL Client for the non-public Gateway.
func NewPriaveteClient(username string, password string, opts ...ClientOption) (*Client, error) {
	baseURL, err := url.Parse(defaultGatewayUrl)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	sc := session.NewSessionCreator()
	userSession, err := sc.Create(username, password)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	swoClient := &Client{
		userSession: userSession,
		baseURL:     baseURL,
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
		Transport: &userSessionAuthTransport{
			client: swoClient,
		},
	})

	swoClient.alertsService = NewAlertsService(swoClient)

	return swoClient, nil
}

// RoundTrip implements the http.RoundTrip interface.
func (t *userSessionAuthTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	// When using the RoundTipper interface it is important that the original request is not modified.
	// See https://pkg.go.dev/net/http#RoundTripper for more details.
	clone := request.Clone(context.Background())

	clone.Header.Set("Cookie", fmt.Sprintf("swi-settings=%s", t.client.userSession.SessionId))
	clone.Header.Set("x-csrf-token", t.client.userSession.CSRFToken)

	if t.client.debugMode {
		dumpRequest(clone)
	}

	response, err := t.client.transport.RoundTrip(clone)

	if t.client.debugMode {
		dumpResponse(response)
	}

	return response, err
}
