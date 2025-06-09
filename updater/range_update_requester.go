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

	"github.com/dtomschitz/headless-go-client/manifest"
)

var _ UpdateRequester = &DefaultRangeUpdateRequester{}

type DefaultRangeUpdateRequester struct {
	Client      *http.Client
	TempDir     string
	ChunkSize   int64
	TargetPerms os.FileMode
}

func (r *DefaultRangeUpdateRequester) Fetch(ctx context.Context, manifest *manifest.Manifest) (io.ReadCloser, error) {
	if r.Client == nil {
		return nil, errors.New("http client cannot be nil")
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

	var start int64
	if info, err := os.Stat(tmpPath); err == nil {
		start = info.Size()
	}

	headReq, _ := http.NewRequestWithContext(ctx, http.MethodHead, manifest.URL, nil)
	resp, err := r.Client.Do(headReq)
	if err != nil {
		return nil, fmt.Errorf("HEAD request failed: %w", err)
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
			return nil, fmt.Errorf("range request failed: %w", err)
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

	file, err := os.Open(tmpPath)
	if err != nil {
		return nil, fmt.Errorf("failed to reopen file: %w", err)
	}

	return &autoDeleteReadCloser{
		File: file,
		path: tmpPath,
	}, nil
}
