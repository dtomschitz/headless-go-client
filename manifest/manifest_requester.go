package manifest

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type DefaultManifestRequester struct {
	client *http.Client
}

var _ ManifestRequester = &DefaultManifestRequester{}

// NewDefaultManifestRequester returns a DefaultManifestRequester with a default HTTP client if none provided.
func NewDefaultManifestRequester(client *http.Client) *DefaultManifestRequester {
	if client == nil {
		client = http.DefaultClient
	}
	return &DefaultManifestRequester{client: client}
}

func (r *DefaultManifestRequester) Fetch(ctx context.Context, url string) (*Manifest, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get manifest: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var m Manifest
	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}

	return &m, nil
}
