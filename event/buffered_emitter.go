package event

import (
	"context"
	"log/slog"
	"sync"
)

type (
	BufferedEmitter struct {
		queue        chan *Event
		bufferMu     sync.Mutex
		buffer       []*Event
		dropCallback func(*Event)

		closed chan struct{}
		done   chan struct{}
		once   sync.Once
	}

	BufferedEmitterConfig struct {
		BufferSize   int
		DropCallback func(event *Event)
	}
)

func NewBufferedEmitter(config BufferedEmitterConfig) Emitter {
	if config.BufferSize <= 0 {
		config.BufferSize = 1024
	}

	e := &BufferedEmitter{
		queue:        make(chan *Event, config.BufferSize),
		buffer:       make([]*Event, 0, config.BufferSize),
		dropCallback: config.DropCallback,
		closed:       make(chan struct{}),
		done:         make(chan struct{}),
	}

	go e.collector()

	return e
}

func (e *BufferedEmitter) collector() {
	defer close(e.done)

	// Drain any remaining events from the queue upon shutdown signal.
	// This ensures no events are lost in the queue if Push stopped before
	// the collector fully processes everything.
	defer func() {
		for {
			select {
			case evt := <-e.queue:
				e.bufferMu.Lock()
				e.buffer = append(e.buffer, evt)
				e.bufferMu.Unlock()
			default: // Queue is empty, done draining
				return
			}
		}
	}()

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
}

// Push adds an event to the emitter queue. If the emitter is closed or the queue is full,
// it will drop the event and call the drop callback if provided.
func (e *BufferedEmitter) Push(event *Event) {
	select {
	case <-e.done: // If collector goroutine is done, emitter is fully closed
		if e.dropCallback != nil {
			e.dropCallback(event)
		}
		return
	default:
		// Not fully closed yet, proceed
	}

	select {
	case e.queue <- event:
	default: // Queue is full, or it was just closed by the collector before this send
		if e.dropCallback != nil {
			e.dropCallback(event)
		}
	}
}

// PollEvents returns and clears current event buffer
func (e *BufferedEmitter) PollEvents() []*Event {
	e.bufferMu.Lock()
	defer e.bufferMu.Unlock()
	events := e.buffer
	e.buffer = make([]*Event, 0, cap(e.buffer))
	return events
}

// Close stops the internal collector goroutine and waits for it to finish.
// It also ensures any remaining events in the queue are moved to the buffer
// so they can be retrieved by PollEvents.
func (e *BufferedEmitter) Close(ctx context.Context) error {
	var err error
	e.once.Do(func() {
		close(e.closed)

		select {
		case <-e.done:
			// Collector finished successfully
		case <-ctx.Done():
			// Context cancelled or timed out before collector finished
			err = ctx.Err()
			slog.Warn("BufferedEmitter collector did not finish draining in time", "error", err) // [CHANGE 20] Log warning
		}
	})

	return err
}
