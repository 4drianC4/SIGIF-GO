package events

import (
	"context"
	"sync"
	"time"
)

type Event interface {
	Name() string
	Payload() any
	OccurredAt() int64
}

type BaseEvent struct {
	name       string
	payload    any
	occurredAt int64
}

func NewEvent(name string, payload any) *BaseEvent {
	return &BaseEvent{
		name:       name,
		payload:    payload,
		occurredAt: time.Now().UnixMilli(),
	}
}

func (e *BaseEvent) Name() string       { return e.name }
func (e *BaseEvent) Payload() any       { return e.payload }
func (e *BaseEvent) OccurredAt() int64  { return e.occurredAt }

type Handler func(ctx context.Context, event Event) error

type Bus struct {
	mu       sync.RWMutex
	handlers map[string][]Handler
}

func NewBus() *Bus {
	return &Bus{
		handlers: make(map[string][]Handler),
	}
}

func (b *Bus) Subscribe(eventName string, handler Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[eventName] = append(b.handlers[eventName], handler)
}

func (b *Bus) Publish(ctx context.Context, event Event) error {
	b.mu.RLock()
	handlers := b.handlers[event.Name()]
	b.mu.RUnlock()

	for _, handler := range handlers {
		if err := handler(ctx, event); err != nil {
			return err
		}
	}
	return nil
}

func (b *Bus) PublishAsync(ctx context.Context, event Event) {
	go func() {
		_ = b.Publish(ctx, event)
	}()
}