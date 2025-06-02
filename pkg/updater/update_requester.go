package updater

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

type (
	UpdateRequester interface {
		Fetch(ctx context.Context, url string) (io.ReadCloser, error)
	}
)

type DefaultUpdateRequester struct {
	client *http.Client
}

var _ UpdateRequester = &DefaultUpdateRequester{}

func (requester *DefaultUpdateRequester) Fetch(ctx context.Context, url string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := requester.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL %s: %w", url, err)
	}

	return resp.Body, nil
}
