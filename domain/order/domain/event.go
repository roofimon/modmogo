package domain

import "time"

const (
	EventOrderPlaced           = "order.placed"
	EventOrderPaymentCompleted = "order.payment_completed"
	EventOrderCancelled        = "order.cancelled"
)

type OrderPlaced struct {
	OrderID    string
	CustomerID *string
	Total      float64
}

type OrderPaymentCompleted struct {
	NewOrderID      string
	OriginalOrderID string
	Total           float64
}

type OrderCancelled struct {
	OrderID       string
	DeactivatedAt time.Time
}
