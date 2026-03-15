package emitter

import (
	"context"
	"sync"
	"sync/atomic"
)

// Subscriber represents a function that handles emitted events.
// It receives a context for cancellation and the event value.
type Subscriber[T any] func(ctx context.Context, value T)

// UnsubscribeFunc is a function returned by Subscribe() to remove a subscriber.
// Calling it unsubscribes the handler from future events.
type UnsubscribeFunc func()

// Emitter defines an interface for event-driven communication.
// It allows subscribing to events and emitting values to all subscribers.
type Emitter[T any] interface {
	// Emit sends a value to all active subscribers.
	// Each subscriber runs in its own goroutine with a derived context from closer.
	Emit(ctx context.Context, value T)

	// Subscribe adds a new subscriber to the emitter.
	// Returns an UnsubscribeFunc that can be used to remove the subscriber.
	Subscribe(subscriber Subscriber[T]) UnsubscribeFunc

	Wait()
}

// emitter is the internal implementation of Emitter interface.
// It uses atomic counter for unique IDs and RWMutex for thread safety.
type emitter[T any] struct {
	subscribers map[uint64]Subscriber[T]
	mutex       sync.RWMutex
	nextID      atomic.Uint64
	wg          sync.WaitGroup
}

// NewEmitter creates a new Emitter instance.
// The closer is used to derive contexts for subscriber goroutines,
// enabling graceful shutdown of all event handlers.
func NewEmitter[T any]() Emitter[T] {
	return &emitter[T]{
		subscribers: make(map[uint64]Subscriber[T]),
	}
}

// Emit sends a value to all subscribers.
// Each subscriber is executed in its own goroutine with a context
// that is cancelled when the function returns or closer signals shutdown.
func (emitter *emitter[T]) Emit(ctx context.Context, value T) {
	emitter.mutex.RLock()
	defer emitter.mutex.RUnlock()

	emitter.wg.Add(len(emitter.subscribers))
	for _, subscriber := range emitter.subscribers {
		go func(ctx context.Context, subscriber Subscriber[T]) {
			ctx, cancelFunc := context.WithCancel(ctx)
			defer cancelFunc()
			defer emitter.wg.Done()

			subscriber(ctx, value)
		}(ctx, subscriber)
	}
}

// Subscribe adds a new subscriber and returns an unsubscribe function.
// The subscriber will receive all future emitted values until unsubscribed.
// IDs are generated atomically to ensure uniqueness.
func (emitter *emitter[T]) Subscribe(callback Subscriber[T]) UnsubscribeFunc {
	emitter.mutex.Lock()
	defer emitter.mutex.Unlock()

	id := emitter.nextID.Add(1)

	emitter.subscribers[id] = callback

	return UnsubscribeFunc(func() {
		emitter.unsubscribe(id)
	})
}

// unsubscribe removes a subscriber by its ID.
// It's called internally by the returned UnsubscribeFunc.
func (emitter *emitter[T]) unsubscribe(id uint64) {
	emitter.mutex.Lock()
	defer emitter.mutex.Unlock()

	delete(emitter.subscribers, id)
}

func (emitter *emitter[T]) Wait() {
	emitter.wg.Wait()
}
