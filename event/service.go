package event

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	commonCtx "github.com/dtomschitz/headless-go-client/context"
	commonHttp "github.com/dtomschitz/headless-go-client/http"
	"github.com/dtomschitz/headless-go-client/logger"
)

type (
	Service struct {
		endpoint string
		logger   logger.Logger

		requestBuilder RequestBuilder
		producers      []Producer

		interval time.Duration
		client   *http.Client

		internalCtx    context.Context
		internalCancel context.CancelFunc
		wg             sync.WaitGroup
		mu             sync.RWMutex
		shutdownOnce   sync.Once
	}

	RequestBuilder func(ctx context.Context, events []*Event) (*http.Request, error)
)

const (
	ServiceName = "EventService"
)

func defaultRequestBuilder(endpoint string) RequestBuilder {
	return func(ctx context.Context, events []*Event) (*http.Request, error) {
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

func NewService(ctx context.Context, endpoint string, opts ...ServiceOption) (*Service, error) {
	internalCtx, internalCancel := context.WithCancel(context.WithValue(ctx, commonCtx.ServiceKey, ServiceName))

	httpClient := commonHttp.NewClient()

	service := &Service{
		internalCtx:    internalCtx,
		internalCancel: internalCancel,
		endpoint:       endpoint,
		interval:       1 * time.Minute,
		client:         httpClient,
		logger:         &logger.NoOpLogger{},
		requestBuilder: defaultRequestBuilder(endpoint),
	}

	for _, opt := range opts {
		if optName, err := opt(internalCtx, service); err != nil {
			internalCancel()
			return nil, fmt.Errorf("failed to apply option %s: %w", optName, err)
		}
	}

	service.start(internalCtx)
	service.logger.Info("started service successfully", "pushInterval", service.interval)

	return service, nil
}

func (s *Service) start(ctx context.Context) {
	s.wg.Add(1)

	go func() {
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()
		defer s.wg.Done()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := s.Flush(ctx); err != nil {
					s.logger.Error("failed to flush events", "error", err)
				}
			}
		}
	}()
}

func (s *Service) RegisterProducer(e Producer) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.producers = append(s.producers, e)
}

func (s *Service) Flush(ctx context.Context) error {
	s.mu.RLock()
	producers := make([]Producer, len(s.producers))
	copy(producers, s.producers)
	s.mu.RUnlock()

	var batch []*Event
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

func (s *Service) Name() string {
	return ServiceName
}

func (s *Service) Close(ctx context.Context) error {
	s.shutdownOnce.Do(func() {
		if s.internalCancel != nil {
			s.internalCancel()
		}
	})

	s.mu.RLock()
	producers := make([]Producer, len(s.producers))
	copy(producers, s.producers)
	s.mu.RUnlock()

	for _, p := range producers {
		if err := p.Close(ctx); err != nil {
			s.logger.Error("failed to close event producer: %v", err)
		}
	}

	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	<-done

	return nil
}
