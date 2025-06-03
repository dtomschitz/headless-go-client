package event

import (
	"context"
	"sync"
)

type (
	BufferedEmitter struct {
		queue        chan Event
		bufferMu     sync.Mutex
		buffer       []Event
		dropCallback func(Event)
		closed       chan struct{}
	}

	BufferedEmitterConfig struct {
		BufferSize   int
		DropCallback func(evt Event)
	}
)

func NewBufferedEmitter(config BufferedEmitterConfig) Emitter {
	if config.BufferSize <= 0 {
		config.BufferSize = 1024
	}

	e := &BufferedEmitter{
		queue:        make(chan Event, config.BufferSize),
		buffer:       make([]Event, 0),
		dropCallback: config.DropCallback,
		closed:       make(chan struct{}),
	}

	go func() {
		for {
			select {
			case evt := <-e.queue:
				e.bufferMu.Lock()
				e.buffer = append(e.buffer, evt)
				e.bufferMu.Unlock()
			case <-e.closed:
				return
			}
		}
	}()

	return e
}

// Push adds an event to the emitter queue. If the queue is full, it will drop the event and call the drop callback if provided.
func (e *BufferedEmitter) Push(evt Event) {
	select {
	case <-e.closed:
		if e.dropCallback != nil {
			e.dropCallback(evt)
		}
	default:
		select {
		case e.queue <- evt:
		default:
			if e.dropCallback != nil {
				e.dropCallback(evt)
			}
		}
	}
}

// PollEvents returns and clears current event buffer
func (e *BufferedEmitter) PollEvents() []Event {
	e.bufferMu.Lock()
	defer e.bufferMu.Unlock()
	events := e.buffer
	e.buffer = nil
	return events
}

// Close stops the internal collector goroutine
func (e *BufferedEmitter) Close(ctx context.Context) error {
	close(e.closed)
	close(e.queue)
	return nil
}
