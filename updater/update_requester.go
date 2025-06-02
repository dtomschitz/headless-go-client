package updater

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

type (
	UpdateRequester interface {
		Fetch(ctx context.Context, manifest *Manifest) (io.ReadCloser, error)
	}
)

var _ UpdateRequester = &DefaultUpdateRequester{}

type (
	DefaultUpdateRequester struct {
		Client *http.Client
	}

	DefaultRangeUpdateRequester struct {
		Client      *http.Client
		TempDir     string
		ChunkSize   int64
		TargetPerms os.FileMode
	}
)

func (r *DefaultUpdateRequester) Fetch(ctx context.Context, manifest *Manifest) (io.ReadCloser, error) {
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

	return resp.Body, nil
}

func (r *DefaultRangeUpdateRequester) Fetch(ctx context.Context, manifest *Manifest) (io.ReadCloser, error) {
	if r.Client == nil {
		return nil, errors.New("http client can not be nil")
	}
	if r.ChunkSize == 0 {
		r.ChunkSize = 2 * 1024 * 1024
	}

	tmpPath := filepath.Join(r.TempDir, fmt.Sprintf("update-%s.tmp", manifest.Version))
	out, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY, r.TargetPerms)
	if err != nil {
		return nil, fmt.Errorf("failed to open temp file: %w", err)
	}
	defer out.Close()

	// Determine file size if partial download already exists
	var start int64
	if info, err := os.Stat(tmpPath); err == nil {
		start = info.Size()
	}

	// Determine total size from a HEAD request
	headReq, _ := http.NewRequestWithContext(ctx, http.MethodHead, manifest.URL, nil)
	resp, err := r.Client.Do(headReq)
	if err != nil {
		return nil, fmt.Errorf("failed HEAD request: %w", err)
	}
	resp.Body.Close()

	totalSize, err := strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid Content-Length: %w", err)
	}

	for start < totalSize {
		end := start + r.ChunkSize - 1
		if end >= totalSize {
			end = totalSize - 1
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, manifest.URL, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, end))

		resp, err := r.Client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed range request: %w", err)
		}

		if resp.StatusCode != http.StatusPartialContent && resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}

		n, err := io.Copy(out, resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("error writing chunk: %w", err)
		}
		start += n
	}

	final, err := os.Open(tmpPath)
	if err != nil {
		return nil, fmt.Errorf("failed to reopen file: %w", err)
	}

	return &autoDeleteReadCloser{
		File: final,
		path: tmpPath,
	}, nil
}
