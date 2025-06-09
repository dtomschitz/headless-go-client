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
	cfg := &config{
		retryCount:   3,
		retryBackoff: 500 * time.Millisecond,
	}

	for _, o := range opts {
		o(cfg)
	}

	ctxTransport := NewContextHeaderTransport(http.DefaultTransport)
	retryTransport := NewRetryTransport(ctxTransport, cfg.retryCount, cfg.retryBackoff)

	return &http.Client{
		Transport: retryTransport,
		Timeout:   cfg.timeout,
	}
}
