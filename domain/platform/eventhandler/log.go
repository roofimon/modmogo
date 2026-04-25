package eventhandler

import (
	"context"
	"encoding/json"
	"log/slog"

	"modmono/domain/platform/event"
)

// LogHandler logs every event it receives using structured slog output.
// New side effects (email, inventory, CRM) follow this same pattern.
type LogHandler struct{}

func (LogHandler) Handle(_ context.Context, e event.Event) error {
	b, _ := json.Marshal(e.Payload)
	slog.Info("event", "type", e.Type, "occurred_at", e.OccurredAt, "payload", string(b))
	return nil
}
