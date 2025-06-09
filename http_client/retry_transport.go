package http_client

import (
	"net/http"
	"time"
)

type RetryTransport struct {
	base         http.RoundTripper
	retryCount   int
	retryBackoff time.Duration
}

func NewRetryTransport(base http.RoundTripper, retryCount int, retryBackoff time.Duration) http.RoundTripper {
	if base == nil {
		base = http.DefaultTransport
	}
	if retryCount < 0 {
		retryCount = 3
	}
	if retryBackoff <= 0 {
		retryBackoff = 500 * time.Millisecond
	}

	return &RetryTransport{
		base:         base,
		retryCount:   retryCount,
		retryBackoff: retryBackoff,
	}
}

func (t *RetryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	clonedReq := req.Clone(req.Context())

	var lastErr error
	for attempt := 0; attempt <= t.retryCount; attempt++ {
		resp, err := t.base.RoundTrip(clonedReq)
		if err == nil {
			return resp, nil
		}
		lastErr = err

		select {
		case <-clonedReq.Context().Done():
			return nil, clonedReq.Context().Err()
		case <-time.After(t.retryBackoff):
		}
	}

	return nil, lastErr
}
