package product

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Product is the product aggregate root persisted in MongoDB.
type Product struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name      string             `bson:"name" json:"name"`
	Price     float64            `bson:"price" json:"price"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}

// CreateInput is the payload for creating a product.
type CreateInput struct {
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}
