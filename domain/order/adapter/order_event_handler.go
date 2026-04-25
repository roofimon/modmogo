package adapter

import (
	"context"
	"errors"

	"modmono/domain/order/domain"
	"modmono/domain/order/port"
	"modmono/domain/platform/event"
)

// OrderSaveHandler persists an order when an OrderPlaced event is received.
type OrderSaveHandler struct {
	repo port.Repository
}

func NewOrderSaveHandler(r port.Repository) *OrderSaveHandler {
	return &OrderSaveHandler{repo: r}
}

func (h *OrderSaveHandler) Handle(ctx context.Context, e event.Event) error {
	payload, ok := e.Payload.(domain.OrderPlaced)
	if !ok {
		return errors.New("order save handler: unexpected payload type")
	}
	result := h.repo.Create(ctx, &payload.Order)
	if result.IsError() {
		return result.Error()
	}
	return nil
}
