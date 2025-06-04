package event_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/dtomschitz/headless-go-client/event"

	"github.com/stretchr/testify/assert"
)

func TestBufferedEmitter_PushAndPoll(t *testing.T) {
	emitter := event.NewBufferedEmitter(event.BufferedEmitterConfig{
		BufferSize: 10,
	})

	defer emitter.Close(context.Background())

	emitter.Push(&event.Event{Message: "test1"})
	emitter.Push(&event.Event{Message: "test2"})

	// Let goroutine process queue
	time.Sleep(10 * time.Millisecond)

	events := emitter.PollEvents()
	assert.Len(t, events, 2)
	assert.Equal(t, "test1", events[0].Message)
	assert.Equal(t, "test2", events[1].Message)
}

func TestBufferedEmitter_BufferClearsAfterPoll(t *testing.T) {
	emitter := event.NewBufferedEmitter(event.BufferedEmitterConfig{
		BufferSize: 5,
	})

	defer emitter.Close(context.Background())

	emitter.Push(&event.Event{Message: "one"})
	time.Sleep(10 * time.Millisecond)

	events := emitter.PollEvents()
	assert.Len(t, events, 1)

	// Poll again; should be empty
	events = emitter.PollEvents()
	assert.Len(t, events, 0)
}

func TestBufferedEmitter_DropCallbackCalled(t *testing.T) {
	var dropped []*event.Event
	var mu sync.Mutex

	emitter := event.NewBufferedEmitter(event.BufferedEmitterConfig{
		BufferSize: 1,
		DropCallback: func(event *event.Event) {
			mu.Lock()
			defer mu.Unlock()
			dropped = append(dropped, event)
		},
	})

	defer emitter.Close(context.Background())

	// Fill the buffer
	emitter.Push(&event.Event{Message: "ok"})
	emitter.Push(&event.Event{Message: "drop me"})

	time.Sleep(10 * time.Millisecond)

	mu.Lock()
	assert.Len(t, dropped, 1)
	assert.Equal(t, "drop me", dropped[0].Message)
	mu.Unlock()
}

func TestBufferedEmitter_Close(t *testing.T) {
	emitter := event.NewBufferedEmitter(event.BufferedEmitterConfig{
		BufferSize: 5,
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	assert.NoError(t, emitter.Close(ctx))

	// Push after close should not panic or block
	done := make(chan struct{})
	go func() {
		defer close(done)
		emitter.Push(&event.Event{Message: "ignored"})
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("push after close blocked")
	}
}
