package client

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

// apiTokentAuthTransport is an http.RoundTripper that authenticates all requests
// using Bearer token authentication with a SWO api key.
type apiTokenAuthTransport struct {
	client   *Client
	apiToken string
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
		DumpRequest(clone)
	}

	response, err := http.DefaultTransport.RoundTrip(clone)

	if t.client.debugMode {
		DumpResponse(response)
	}

	return response, err
}
