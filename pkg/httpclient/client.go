package httpclient

import (
	"net"
	"net/http"
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
