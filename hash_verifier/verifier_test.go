package hash_verifier_test

import (
	"bytes"
	"crypto/md5"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"hash"
	"io"
	"strings"
	"testing"

	"github.com/dtomschitz/headless-go-client/hash_verifier"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewVerifierFromHashString(t *testing.T) {
	tests := []struct {
		name      string
		hashStr   string
		shouldErr bool
	}{
		{"valid md5", "md5:d41d8cd98f00b204e9800998ecf8427e", false},
		{"valid sha256", "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", false},
		{"valid sha512", "sha512:cf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce47d0d13c5d85f2b0ff8318d2877eec2f63b931bd47417a81a538327af927da3e", false},
		{"missing algo", "abcdef", true},
		{"unsupported algo", "foo:abcdef", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			verifier, expected, err := hash_verifier.NewVerifierFromHashString(tc.hashStr)
			if tc.shouldErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.NotEmpty(t, expected)

			if strings.HasPrefix(tc.hashStr, "md5:") && expected == "d41d8cd98f00b204e9800998ecf8427e" {
				assert.NoError(t, verifier.Verify(bytes.NewReader([]byte{})))
			}
		})
	}
}

func TestVerifier_Verify(t *testing.T) {
	tests := []struct {
		name    string
		algo    string
		data    []byte
		wantErr bool
	}{
		{"md5 match", "md5", []byte("hello world"), false},
		{"md5 mismatch", "md5", []byte("hello world!"), true},
		{"sha256 match", "sha256", []byte("hello world"), false},
		{"sha256 mismatch", "sha256", []byte("hello world!"), true},
		{"sha512 match", "sha512", []byte("hello world"), false},
		{"sha512 mismatch", "sha512", []byte("hello world!"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			verifier := createVerifierWithData(t, tt.algo, []byte("hello world"))
			err := verifier.Verify(bytes.NewReader(tt.data))
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func createVerifierWithData(t *testing.T, algo string, data []byte) hash_verifier.Verifier {
	var h hash.Hash
	switch algo {
	case "md5":
		h = md5.New()
	case "sha256":
		h = sha256.New()
	case "sha512":
		h = sha512.New()
	default:
		t.Fatalf("unsupported algo: %s", algo)
	}

	_, err := h.Write(data)
	require.NoError(t, err)

	expected := hex.EncodeToString(h.Sum(nil))
	verifier, err := hash_verifier.NewVerifier(algo, expected)
	require.NoError(t, err)
	return verifier
}

func TestVerifyHash_ErrorOnReadFailure(t *testing.T) {
	verifier, err := hash_verifier.NewVerifier("md5", "dummy")
	require.NoError(t, err)

	errReader := &errorReader{}
	err = verifier.Verify(errReader)
	assert.Error(t, err)
}

type errorReader struct{}

func (e *errorReader) Read(p []byte) (int, error) {
	return 0, io.ErrUnexpectedEOF
}
