package application

import "go.mongodb.org/mongo-driver/bson/primitive"

// parseObjectID converts a 24-char hex string to a MongoDB ObjectID.
func parseObjectID(s string) (primitive.ObjectID, error) {
	if len(s) != 24 {
		return primitive.NilObjectID, ErrInvalidObjectID
	}
	id, err := primitive.ObjectIDFromHex(s)
	if err != nil {
		return primitive.NilObjectID, ErrInvalidObjectID
	}
	return id, nil
}
