package updater

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/dtomschitz/headless-go-client/manifest"
)

type (
	UpdateRequester interface {
		Fetch(ctx context.Context, manifest *manifest.Manifest) (io.ReadCloser, error)
	}
)

var _ UpdateRequester = &DefaultUpdateRequester{}

type (
	DefaultUpdateRequester struct {
		Client *http.Client
	}
)

func (r *DefaultUpdateRequester) Fetch(ctx context.Context, manifest *manifest.Manifest) (io.ReadCloser, error) {
	if r.Client == nil {
		return nil, errors.New("http client can not be nil")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, manifest.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := r.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch  URL %s: %w", manifest.URL, err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return resp.Body, nil
}
