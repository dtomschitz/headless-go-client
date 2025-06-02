package event

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/dtomschitz/headless-go-client/logger"
	"net/http"
	"sync"
	"time"
)

type (
	Service struct {
		endpoint string
		logger   logger.Logger

		requestBuilder RequestBuilder
		producers      []Emitter

		interval time.Duration
		ticker   *time.Ticker
		client   *http.Client

		wg           sync.WaitGroup
		mu           sync.RWMutex
		shutdownOnce sync.Once
	}

	ServiceOption func(context.Context, *Service) (string, error)

	RequestBuilder func(ctx context.Context, events []Event) (*http.Request, error)
)

func defaultRequestBuilder(endpoint string) RequestBuilder {
	return func(ctx context.Context, events []Event) (*http.Request, error) {
		payload, err := json.Marshal(events)
		if err != nil {
			return nil, err
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(payload))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		return req, nil
	}
}

func WithRequestBuilder(builder RequestBuilder) ServiceOption {
	return func(ctx context.Context, s *Service) (string, error) {
		if builder == nil {
			return "", fmt.Errorf("request builder cannot be nil")
		}
		s.requestBuilder = builder
		return "WithRequestBuilder", nil
	}
}

func WithLogger(logger logger.Logger) ServiceOption {
	return func(ctx context.Context, s *Service) (string, error) {
		if logger == nil {
			return "", fmt.Errorf("logger cannot be nil")
		}

		s.logger = logger
		return "WithLogger", nil
	}
}

func NewService(ctx context.Context, endpoint string, interval time.Duration, opts ...ServiceOption) (*Service, error) {
	service := &Service{
		endpoint:       endpoint,
		interval:       interval,
		client:         &http.Client{Timeout: 5 * time.Second},
		requestBuilder: defaultRequestBuilder(endpoint),
	}

	for _, opt := range opts {
		if optName, err := opt(ctx, service); err != nil {
			return nil, fmt.Errorf("failed to apply option %s: %w", optName, err)
		}
	}

	return service, nil
}

func (s *Service) RegisterProducer(e Emitter) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.producers = append(s.producers, e)
}

func (s *Service) Start(ctx context.Context) {
	s.ticker = time.NewTicker(s.interval)
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case <-s.ticker.C:
				err := s.Flush(ctx)
				if err != nil {
					return
				}
			}
		}
	}()
}

func (s *Service) Flush(ctx context.Context) error {
	s.mu.RLock()
	producers := make([]Emitter, len(s.producers))
	copy(producers, s.producers)
	s.mu.RUnlock()

	var batch []Event
	for _, p := range producers {
		batch = append(batch, p.PollEvents()...)
	}

	if len(batch) == 0 {
		return nil
	}

	req, err := s.requestBuilder(ctx, batch)
	if err != nil {
		return fmt.Errorf("failed to build request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("received non-2xx response: %s", resp.Status)
	}

	return nil
}

func (s *Service) Close(ctx context.Context) error {
	s.shutdownOnce.Do(func() {
		s.ticker.Stop()
	})

	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
		return nil
	}
}
