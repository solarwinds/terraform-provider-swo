package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	defaultBaseURL   = "https://public-api.dc-01.dev-ssp.solarwinds.com/public/schema"
	defaultMediaType = "application/json"
	clientIdentifier = "swo-api-go"
)

var (
	client = &http.Client{
		Timeout: 30 * time.Second,
	}
)

// ServiceAccessor defines an interface for talking to via domain-specific service constructs
type ServiceAccessor interface {
	AlertsService() AlertsCommunicator
}

// Client implements ServiceAccessor
type Client struct {
	baseURL                 *url.URL
	httpClient              httpClient
	apiToken                string
	alertsService           AlertsCommunicator
	callerUserAgentFragment string
	debugMode               bool
}

// httpClient defines the http.Client method used by Client.
type httpClient interface {
	Do(*http.Request) (*http.Response, error)
}

// ClientOption provides functional option-setting behavior
type ClientOption func(*Client) error

// New returns a new SWO API client. Optional arguments UserAgentClientOption and BaseURLClientOption can be provided.
func NewClient(apiToken string, opts ...func(*Client) error) *Client {
	baseURL, err := url.Parse(defaultBaseURL)

	if err != nil {
		log.Println(err)
		return nil
	}

	c := &Client{
		apiToken:   apiToken,
		baseURL:    baseURL,
		httpClient: client,
	}

	c.alertsService = NewAlertsService(c)

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// NewRequest standardizes the request being sent
func (c *Client) NewRequest(method, path string, body interface{}) (*http.Request, error) {
	rel, err := url.Parse(path)
	if err != nil {
		return nil, err
	}

	requestURL := c.baseURL.ResolveReference(rel)

	var buffer io.ReadWriter

	if body != nil {
		buffer = &bytes.Buffer{}
		encodeErr := json.NewEncoder(buffer).Encode(body)
		if encodeErr != nil {
			log.Println(encodeErr)
		}
	}

	req, err := http.NewRequest(method, requestURL.String(), buffer)

	if err != nil {
		return nil, err
	}

	req.SetBasicAuth("token", c.apiToken)
	req.Header.Set("Accept", defaultMediaType)
	req.Header.Set("Content-Type", defaultMediaType)
	req.Header.Set("User-Agent", c.completeUserAgentString())

	return req, nil
}

// UserAgentClientOption is a config function allowing setting of the User-Agent header in requests
func UserAgentClientOption(userAgentString string) ClientOption {
	return func(c *Client) error {
		c.callerUserAgentFragment = userAgentString
		return nil
	}
}

// BaseURLClientOption is a config function allowing setting of the base URL the API is on
func BaseURLClientOption(urlString string) ClientOption {
	return func(c *Client) error {
		var altURL *url.URL
		var err error
		if altURL, err = url.Parse(urlString); err != nil {
			return err
		}
		c.baseURL = altURL
		return nil
	}
}

// SetDebugMode sets the debugMode struct member to true
func SetDebugMode() ClientOption {
	return func(c *Client) error {
		c.debugMode = true
		return nil
	}
}

// SetHTTPClient allows the user to provide a custom http.Client configuration
func SetHTTPClient(client *http.Client) ClientOption {
	return func(c *Client) error {
		c.httpClient = client
		return nil
	}
}

// AlertsService represents the subset of the API that deals with Alerts
func (c *Client) AlertsService() AlertsCommunicator {
	return c.alertsService
}

// Do performs the HTTP request on the wire, taking an optional second parameter for containing a response.
func (c *Client) Do(req *http.Request, respData interface{}) (*http.Response, error) {
	if c.debugMode {
		dumpRequest(req)
	}

	resp, err := c.httpClient.Do(req)

	if err != nil {
		return resp, err
	}

	if c.debugMode {
		dumpResponse(resp)
	}

	if err = checkError(resp); err != nil {
		return resp, err
	}

	defer resp.Body.Close()
	if respData != nil {
		err = json.NewDecoder(resp.Body).Decode(respData)
	}

	return resp, err
}

// completeUserAgentString returns the string that will be placed in the User-Agent header.
// It ensures that any caller-set string has the client name and version appended to it.
func (c *Client) completeUserAgentString() string {
	if c.callerUserAgentFragment == "" {
		return clientVersionString()
	}
	return fmt.Sprintf("%s:%s", c.callerUserAgentFragment, clientVersionString())
}

// clientVersionString returns the canonical name-and-version string
func clientVersionString() string {
	return clientIdentifier
}

// checkError creates an ErrorResponse from the http.Response.Body, if there is one
func checkError(resp *http.Response) error {
	errResponse := &ErrorResponse{}
	if resp.StatusCode >= 400 {
		errResponse.Status = resp.Status
		errResponse.Response = resp
		if resp.ContentLength != 0 {
			body, _ := ioutil.ReadAll(resp.Body)
			err := json.Unmarshal(body, errResponse)
			if err != nil {
				errResponse.Errors = strconv.Quote(string(body))
			}
			log.Debugf("error: %+v\n", errResponse)
			return errResponse
		}
		return errResponse
	}
	return nil
}

// dumpResponse is a debugging function which dumps the HTTP response to stdout
func dumpResponse(resp *http.Response) {
	fmt.Printf("response status: %s\n", resp.Status)
	dump, err := httputil.DumpResponse(resp, true)

	if err != nil {
		log.Printf("error dumping response: %s", err)
		return
	}
	log.Printf("response body: %s\n\n", string(dump))
}

// dumpRequest is a debugging function which dumps the HTTP request to stdout
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
