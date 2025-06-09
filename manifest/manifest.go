package manifest

import (
	"bytes"
	"context"
	"fmt"

	"github.com/dtomschitz/headless-go-client/hash_verifier"
)

type (
	Manifest struct {
		Version string `json:"version"`
		Hash    string `json:"hash"`
		URL     string `json:"url"`
	}

	ManifestRequester interface {
		Fetch(ctx context.Context, url string) (*Manifest, error)
	}
)

// Verify verifies the content against the hash in the manifest.
func (m *Manifest) Verify(content []byte) error {
	if m.Hash == "" {
		return fmt.Errorf("no hash provided")
	}

	verifier, _, err := hash_verifier.NewVerifierFromHashString(m.Hash)
	if err != nil {
		return fmt.Errorf("failed to create hash verifier: %w", err)
	}

	return verifier.Verify(bytes.NewReader(content))
}
