package event

import (
	"context"
	"time"

	commonCtx "github.com/dtomschitz/headless-go-client/context"
	"github.com/google/uuid"
)

type (
	EventType string

	Event struct {
		Id            string                 `json:"id"`
		DeviceId      string                 `json:"deviceId"`
		ClientVersion string                 `json:"clientVersion"`
		Timestamp     time.Time              `json:"timestamp"`
		Source        string                 `json:"source"`
		Type          EventType              `json:"type"`
		Message       string                 `json:"message"`
		Data          map[string]interface{} `json:"data,omitempty"`
	}

	EventOption func(*Event)
)

// WithMessage creates an EventOption that sets the message of the event.
func WithMessage(message string) EventOption {
	return func(e *Event) {
		e.Message = message
	}
}

// WithData creates an EventOption that sets the data of the event.
func WithData(data map[string]interface{}) EventOption {
	return func(e *Event) {
		e.Data = data
	}
}

// WithDataField creates an EventOption that sets a specific field in the event data.
func WithDataField(key string, value interface{}) EventOption {
	return func(e *Event) {
		if e.Data == nil {
			e.Data = make(map[string]interface{})
		}
		e.Data[key] = value
	}
}

// NewEvent creates a new Event with the given context and event type.
func NewEvent(ctx context.Context, eventType EventType, opts ...EventOption) *Event {
	event := &Event{
		Id:            uuid.New().String(),
		DeviceId:      commonCtx.GetStringValue(ctx, commonCtx.DeviceIdKey),
		ClientVersion: commonCtx.GetStringValue(ctx, commonCtx.ClientVersionKey),
		Type:          eventType,
		Timestamp:     time.Now(),
	}

	for _, opt := range opts {
		opt(event)
	}

	return event
}
