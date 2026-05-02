package httpclient

import (
	"net"
	"net/http"
	"net/url"
	"os"
	"time"
)

// Client is a production-ready HTTP client with sensible defaults.
type Client struct {
	*http.Client
}

// New creates a new HTTP client with timeouts optimized for API calls.
func New() *Client {
	return &Client{
		Client: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				DialContext: (&net.Dialer{
					Timeout:   10 * time.Second,
					KeepAlive: 30 * time.Second,
				}).DialContext,
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
				TLSHandshakeTimeout: 10 * time.Second,
			},
		},
	}
}

// Do overrides the default Do method to rewrite URLs through a public proxy if enabled.
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	if os.Getenv("USE_PUBLIC_PROXY") == "true" {
		proxyURL := "https://api.allorigins.win/raw?url=" + url.QueryEscape(req.URL.String())
		parsedProxyURL, err := url.Parse(proxyURL)
		if err == nil {
			req.URL = parsedProxyURL
			req.Host = parsedProxyURL.Host
		}
	}
	return c.Client.Do(req)
}
