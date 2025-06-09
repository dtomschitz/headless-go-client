package hash_verifier

import (
	"crypto/md5"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"io"
	"strings"
)

type Verifier interface {
	Verify(reader io.Reader) error
}

func NewVerifierFromHashString(hash string) (Verifier, string, error) {
	parts := strings.SplitN(hash, ":", 2)
	if len(parts) != 2 {
		return nil, "", errors.New("invalid hash format: must be algo:hex")
	}

	algo, expected := parts[0], parts[1]
	verifier, err := NewVerifier(algo, expected)
	if err != nil {
		return nil, "", err
	}

	return verifier, expected, nil
}

func NewVerifier(algo string, expected string) (Verifier, error) {
	switch strings.ToLower(algo) {
	case "sha256":
		return &sha256Verifier{expected: strings.ToLower(expected)}, nil
	case "sha512":
		return &sha512Verifier{expected: strings.ToLower(expected)}, nil
	case "md5":
		return &md5Verifier{expected: strings.ToLower(expected)}, nil
	default:
		return nil, fmt.Errorf("unsupported algorithm: %s", algo)
	}
}

type sha256Verifier struct{ expected string }
type sha512Verifier struct{ expected string }
type md5Verifier struct{ expected string }

func (v *sha256Verifier) Verify(r io.Reader) error {
	return verifyHash(r, v.expected, sha256.New())
}

func (v *sha512Verifier) Verify(r io.Reader) error {
	return verifyHash(r, v.expected, sha512.New())
}

func (v *md5Verifier) Verify(r io.Reader) error {
	return verifyHash(r, v.expected, md5.New())
}

func verifyHash(r io.Reader, expected string, h hash.Hash) error {
	if _, err := io.Copy(h, r); err != nil {
		return fmt.Errorf("hashing failed: %w", err)
	}
	actual := hex.EncodeToString(h.Sum(nil))
	if actual != expected {
		return fmt.Errorf("hash mismatch: expected %s, got %s", expected, actual)
	}
	return nil
}
