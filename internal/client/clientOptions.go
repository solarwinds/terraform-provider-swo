package client

import (
	"net/http"
	"net/url"
	"time"
)

// ClientOption provides functional option-setting behavior.
type ClientOption func(*Client) error

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
		if duration > 0 {
			c.requestTimeout = duration
		}
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
		if urlString != "" {
			urlObj, err := url.Parse(urlString)

			if err != nil {
				return err
			}

			c.baseURL = urlObj
		}

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
