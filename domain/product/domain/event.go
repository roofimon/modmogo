package domain

import "time"

const (
	EventProductCreated    = "product.created"
	EventProductActivated  = "product.activated"
	EventProductDeactivated = "product.deactivated"
)

type ProductCreated struct {
	ProductID string
	SKU       string
	Name      string
	Price     float64
}

type ProductActivated struct {
	ProductID string
}

type ProductDeactivated struct {
	ProductID     string
	DeactivatedAt time.Time
}
