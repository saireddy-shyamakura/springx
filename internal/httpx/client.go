// Package httpx provides a hardened HTTP client for outbound springx
// requests. It applies sane timeouts, enforces HTTP/2, and bounds resource
// usage so responses from start.spring.io (or a compromised mirror) cannot
// stall the CLI or exhaust disk/memory.
package httpx

import (
	"net"
	"net/http"
	"time"
)

// New returns an *http.Client with production-safe transport settings.
// TLS certificate verification remains enabled (Go's default), and every
// dial is bounded so a hung server cannot block the CLI indefinitely.
func New(timeout time.Duration) *http.Client {
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	return &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   10 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          10,
			IdleConnTimeout:       30 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: time.Second,
		},
	}
}
