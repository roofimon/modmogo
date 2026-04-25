package event

import (
	"context"
	"log/slog"
	"sync"
)

// Handler is a side effect that reacts to a domain event.
type Handler interface {
	Handle(ctx context.Context, e Event) error
}

// HandlerFunc is a function adapter for Handler.
type HandlerFunc func(ctx context.Context, e Event) error

func (f HandlerFunc) Handle(ctx context.Context, e Event) error { return f(ctx, e) }

// Wildcard subscribes a handler to every event type.
const Wildcard = "*"

// EventBus routes published events to registered handlers.
// It satisfies the Publisher interface so it can replace any Publisher.
type EventBus struct {
	mu       sync.RWMutex
	handlers map[string][]Handler
}

// NewEventBus returns a ready-to-use EventBus with no subscribers.
func NewEventBus() *EventBus {
	return &EventBus{handlers: make(map[string][]Handler)}
}

// Subscribe registers h to receive events of the given type.
// Use Wildcard to receive every event.
func (b *EventBus) Subscribe(eventType string, h Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[eventType] = append(b.handlers[eventType], h)
}

// SubscribeFunc is a convenience wrapper for plain functions.
func (b *EventBus) SubscribeFunc(eventType string, fn func(context.Context, Event) error) {
	b.Subscribe(eventType, HandlerFunc(fn))
}

// Publish delivers e to typed handlers then wildcard handlers.
// Handler errors are logged but never surfaced — a failing side effect
// must not block the primary business flow.
func (b *EventBus) Publish(ctx context.Context, e Event) {
	b.mu.RLock()
	typed     := b.handlers[e.Type]
	wildcards := b.handlers[Wildcard]
	b.mu.RUnlock()

	for _, h := range typed {
		if err := h.Handle(ctx, e); err != nil {
			slog.Error("eventbus: handler error", "event_type", e.Type, "err", err)
		}
	}
	for _, h := range wildcards {
		if err := h.Handle(ctx, e); err != nil {
			slog.Error("eventbus: wildcard handler error", "event_type", e.Type, "err", err)
		}
	}
}

var _ Publisher = (*EventBus)(nil)
