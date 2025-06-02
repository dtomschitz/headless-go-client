package event

import (
	"context"
)

type (
	Emitter interface {
		Push(evt Event)
		PollEvents() []Event
		Close(ctx context.Context) error
	}

	Producer interface {
		PollEvents() []Event
		Close(ctx context.Context) error
	}

	NoopEmitter struct{}
)

func (n NoopEmitter) Push(_ Event)                    {}
func (n NoopEmitter) PollEvents() []Event             { return nil }
func (n NoopEmitter) Close(ctx context.Context) error { return nil }
