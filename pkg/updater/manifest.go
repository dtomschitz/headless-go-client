package updater

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type (
	Manifest struct {
		Version string `json:"version"`
		SHA256  string `json:"sha256"`
		URL     string `json:"url"`
	}

	ManifestRequester interface {
		Fetch(ctx context.Context, url string) (*Manifest, error)
	}
)

type DefaultManifestRequester struct {
	client *http.Client
}

var _ ManifestRequester = &DefaultManifestRequester{}

func (requester *DefaultManifestRequester) Fetch(ctx context.Context, url string) (*Manifest, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := requester.client.Do(req)
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
