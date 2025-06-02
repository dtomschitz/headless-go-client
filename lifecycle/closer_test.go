package lifecycle_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/dtomschitz/headless-go-client/lifecycle"
	"github.com/stretchr/testify/assert"
)

type MockCloser struct {
	mu       sync.Mutex
	closed   bool
	closeErr error
}

func (m *MockCloser) Close(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	return m.closeErr
}

func TestManager_RegisterAndClose(t *testing.T) {
	t.Run("CloseAllWithoutErrors", func(t *testing.T) {
		//given
		manager := lifecycle.NewCloser()
		closer1 := &MockCloser{}
		closer2 := &MockCloser{}

		manager.Register(closer1)
		manager.Register(closer2)

		//when
		err := manager.CloseAll(context.Background())

		//then
		assert.NoError(t, err, "expected no error when closing all closers")
		assert.True(t, closer1.closed, "expected closer1 to be closed")
		assert.True(t, closer2.closed, "expected closer2 to be closed")
	})

	t.Run("CloseAllWithErrors", func(t *testing.T) {
		//given
		manager := lifecycle.NewCloser()
		closer1 := &MockCloser{closeErr: errors.New("error closing closer1")}
		closer2 := &MockCloser{closeErr: errors.New("error closing closer2")}

		manager.Register(closer1)
		manager.Register(closer2)

		//when
		err := manager.CloseAll(context.Background())

		//then
		assert.Error(t, err, "expected error when closing all closers")
		assert.Contains(t, err.Error(), "error closing closer1")
		assert.Contains(t, err.Error(), "error closing closer2")
		assert.True(t, closer1.closed, "expected closer1 to be closed")
		assert.True(t, closer2.closed, "expected closer2 to be closed")
	})

	t.Run("CloseEmptyManager", func(t *testing.T) {
		manager := lifecycle.NewCloser()
		err := manager.CloseAll(context.Background())
		assert.NoError(t, err, "expected no error when closing an empty manager")
	})
}
