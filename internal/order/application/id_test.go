package application

import (
	"testing"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestParseObjectID_emptyString(t *testing.T) {
	_, err := parseObjectID("")
	if err != ErrInvalidObjectID {
		t.Errorf("expected ErrInvalidObjectID, got %v", err)
	}
}

func TestParseObjectID_tooShort(t *testing.T) {
	_, err := parseObjectID("abc123")
	if err != ErrInvalidObjectID {
		t.Errorf("expected ErrInvalidObjectID, got %v", err)
	}
}

func TestParseObjectID_invalidHex(t *testing.T) {
	_, err := parseObjectID("zzzzzzzzzzzzzzzzzzzzzzzz")
	if err != ErrInvalidObjectID {
		t.Errorf("expected ErrInvalidObjectID, got %v", err)
	}
}

func TestParseObjectID_valid(t *testing.T) {
	want := primitive.NewObjectID()
	got, err := parseObjectID(want.Hex())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != want {
		t.Errorf("expected %v, got %v", want, got)
	}
}
