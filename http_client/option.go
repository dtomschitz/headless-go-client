package http_client

import (
	"time"
)

type Option func(*config)

func WithRetry(retryCount int, backoff time.Duration) Option {
	return func(c *config) {
		c.retryCount = retryCount
		c.retryBackoff = backoff
	}
}

func WithTimeout(timeout time.Duration) Option {
	return func(c *config) {
		c.timeout = timeout
	}
}
