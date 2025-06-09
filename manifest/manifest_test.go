package manifest_test

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/dtomschitz/headless-go-client/manifest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func MustHashSHA256(data []byte) string {
	h := sha256.New()
	_, err := h.Write(data)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func TestManifest_Verify(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// given
		content := []byte("hello world")
		hash := MustHashSHA256(content)

		m := manifest.Manifest{
			Hash: "sha256:" + hash,
		}

		// when
		err := m.Verify(content)

		// then
		assert.NoError(t, err)
	})

	t.Run("no hash provided", func(t *testing.T) {
		// given
		m := manifest.Manifest{}

		// when
		err := m.Verify([]byte("anything"))

		// then
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no hash provided")
	})

	t.Run("hash mismatch", func(t *testing.T) {
		// given
		content := []byte("hello world")
		m := manifest.Manifest{
			Hash: "sha256:abcdef",
		}

		// when
		err := m.Verify(content)

		// then
		require.Error(t, err)
		assert.Contains(t, err.Error(), "hash mismatch")
	})
}
