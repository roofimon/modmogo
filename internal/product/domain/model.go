package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	StatusActive      = "active"
	StatusDeactivated = "deactivated"
)

// Product is the product aggregate root persisted in MongoDB.
type Product struct {
	ID            primitive.ObjectID  `bson:"_id,omitempty"            json:"id"`
	SKU           string              `bson:"sku"                      json:"sku"`
	Name          string              `bson:"name"                     json:"name"`
	Price         float64             `bson:"price"                    json:"price"`
	Status        string              `bson:"status"                   json:"status"`
	OriginalID    *primitive.ObjectID `bson:"original_id,omitempty"    json:"original_id,omitempty"`
	CreatedAt     time.Time           `bson:"created_at"               json:"created_at"`
	DeactivatedAt *time.Time          `bson:"deactivated_at,omitempty" json:"deactivated_at,omitempty"`
}

// CreateInput is the payload for creating a product.
type CreateInput struct {
	SKU   string  `json:"sku"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}
