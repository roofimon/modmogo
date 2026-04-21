package customer

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	StatusActive      = "active"
	StatusDeactivated = "deactivated"
)

// Customer is the customer aggregate root persisted in MongoDB.
type Customer struct {
	ID            primitive.ObjectID  `bson:"_id,omitempty"            json:"id"`
	Name          string              `bson:"name"                     json:"name"`
	Email         string              `bson:"email"                    json:"email"`
	Phone         string              `bson:"phone,omitempty"          json:"phone,omitempty"`
	Status        string              `bson:"status"                   json:"status"`
	OriginalID    *primitive.ObjectID `bson:"original_id,omitempty"    json:"original_id,omitempty"`
	CreatedAt     time.Time           `bson:"created_at"               json:"created_at"`
	DeactivatedAt *time.Time          `bson:"deactivated_at,omitempty" json:"deactivated_at,omitempty"`
}

// CreateInput is the payload for creating a customer.
type CreateInput struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone"`
}
