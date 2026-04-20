package order

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// LineItem is a single product line within an order.
type LineItem struct {
	SKU       string  `bson:"sku"        json:"sku"`
	Quantity  int     `bson:"quantity"   json:"quantity"`
	UnitPrice float64 `bson:"unit_price" json:"unit_price"`
}

// Status constants for Order.
const (
	StatusPending          = ""
	StatusPaymentCompleted = "payment_completed"
)

// Order is the order aggregate root persisted in MongoDB.
type Order struct {
	ID              primitive.ObjectID  `bson:"_id,omitempty"               json:"id"`
	CustomerID      *primitive.ObjectID `bson:"customer_id,omitempty"       json:"customer_id,omitempty"`
	Items           []LineItem          `bson:"items"                       json:"items"`
	Total           float64             `bson:"-"                           json:"total"`
	Status          string              `bson:"status,omitempty"            json:"status,omitempty"`
	OriginalOrderID *primitive.ObjectID `bson:"original_order_id,omitempty" json:"original_order_id,omitempty"`
	CreatedAt       time.Time           `bson:"created_at"                  json:"created_at"`
	DeactivatedAt   *time.Time          `bson:"deactivated_at,omitempty"    json:"deactivated_at,omitempty"`
}

// ComputeTotal sets Total = sum(quantity * unit_price) for all items.
// Must be called after every MongoDB read before serialising to JSON.
func (o *Order) ComputeTotal() {
	var sum float64
	for _, item := range o.Items {
		sum += float64(item.Quantity) * item.UnitPrice
	}
	o.Total = sum
}

// LineItemInput is one line item in a CreateInput payload.
type LineItemInput struct {
	SKU       string  `json:"sku"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
}

// CreateInput is the JSON payload for creating an order.
type CreateInput struct {
	CustomerID *string         `json:"customer_id"` // optional 24-hex ObjectID string
	Items      []LineItemInput `json:"items"`
}
