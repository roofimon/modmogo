package event

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"
)

type Event struct {
	Type       string
	OccurredAt time.Time
	Payload    any
}

type Publisher interface {
	Publish(ctx context.Context, e Event)
}

type LogPublisher struct{}

func (LogPublisher) Publish(_ context.Context, e Event) {
	b, _ := json.Marshal(e.Payload)
	slog.Info("event", "type", e.Type, "occurred_at", e.OccurredAt, "payload", string(b))
}
