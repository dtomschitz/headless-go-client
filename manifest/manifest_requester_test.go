package manifest_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dtomschitz/headless-go-client/manifest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultManifestRequester_Fetch(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// given
		expectedManifest := manifest.Manifest{
			Version: "1.2.3",
			Hash:    "sha256:abc123",
			URL:     "http://example.com/config.json",
		}
		data, err := json.Marshal(expectedManifest)
		require.NoError(t, err)

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write(data)
		}))
		defer server.Close()

		req := manifest.NewDefaultManifestRequester(nil)
		ctx := context.Background()

		// when
		m, err := req.Fetch(ctx, server.URL)

		// then
		require.NoError(t, err)
		assert.Equal(t, expectedManifest.Version, m.Version)
		assert.Equal(t, expectedManifest.Hash, m.Hash)
		assert.Equal(t, expectedManifest.URL, m.URL)
	})

	t.Run("unexpected status code", func(t *testing.T) {
		// given
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		req := manifest.NewDefaultManifestRequester(nil)
		ctx := context.Background()

		// when
		_, err := req.Fetch(ctx, server.URL)

		// then
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unexpected status code")
	})

	t.Run("invalid JSON response", func(t *testing.T) {
		// given
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("{invalid json"))
		}))
		defer server.Close()

		req := manifest.NewDefaultManifestRequester(nil)
		ctx := context.Background()

		// when
		_, err := req.Fetch(ctx, server.URL)

		// then
		require.Error(t, err)
		assert.Contains(t, err.Error(), "decode")
	})

	t.Run("request creation error", func(t *testing.T) {
		// given
		req := manifest.NewDefaultManifestRequester(nil)
		ctx := context.Background()

		// when
		_, err := req.Fetch(ctx, "http://\x7f")

		// then
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create request")
	})
}
