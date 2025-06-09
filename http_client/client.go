package http_client

import (
	"net/http"
	"time"
)

type config struct {
	retryCount   int
	retryBackoff time.Duration
	timeout      time.Duration
}

// NewClient returns a *http.Client with retry and context header injection configured.
func NewClient(opts ...Option) *http.Client {
	config := &config{
		timeout:      60 * time.Second,
		retryCount:   3,
		retryBackoff: 500 * time.Millisecond,
	}

	for _, o := range opts {
		o(config)
	}

	ctxTransport := NewContextHeaderTransport(http.DefaultTransport)
	retryTransport := NewRetryTransport(ctxTransport, config.retryCount, config.retryBackoff)

	return &http.Client{
		Transport: retryTransport,
		Timeout:   config.timeout,
	}
}
