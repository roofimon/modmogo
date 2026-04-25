package domain

import (
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestToOrderView_mapsFields(t *testing.T) {
	id := primitive.NewObjectID()
	now := time.Now().UTC()
	o := Order{
		ID:        id,
		Items:     []LineItem{{SKU: "SKU1", Quantity: 2, UnitPrice: 9.99}},
		Total:     19.98,
		Status:    StatusPending,
		CreatedAt: now,
	}
	productNames := map[string]string{"SKU1": "Widget"}

	view := ToOrderView(o, "Alice", productNames)

	if view.ID != id {
		t.Errorf("ID: expected %v, got %v", id, view.ID)
	}
	if view.CustomerName != "Alice" {
		t.Errorf("CustomerName: expected %q, got %q", "Alice", view.CustomerName)
	}
	if view.Total != 19.98 {
		t.Errorf("Total: expected 19.98, got %v", view.Total)
	}
	if len(view.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(view.Items))
	}
	if view.Items[0].ProductName != "Widget" {
		t.Errorf("ProductName: expected %q, got %q", "Widget", view.Items[0].ProductName)
	}
	if view.Items[0].Quantity != 2 {
		t.Errorf("Quantity: expected 2, got %d", view.Items[0].Quantity)
	}
}

func TestToOrderView_noCustomer(t *testing.T) {
	o := Order{
		ID:        primitive.NewObjectID(),
		Items:     []LineItem{{SKU: "SKU1", Quantity: 1, UnitPrice: 5.0}},
		CreatedAt: time.Now(),
	}
	view := ToOrderView(o, "", map[string]string{"SKU1": "Widget"})
	if view.CustomerName != "" {
		t.Errorf("expected empty CustomerName, got %q", view.CustomerName)
	}
	if view.CustomerID != nil {
		t.Errorf("expected nil CustomerID")
	}
}

func TestToOrderView_unknownSKUGetsEmptyName(t *testing.T) {
	o := Order{
		ID:        primitive.NewObjectID(),
		Items:     []LineItem{{SKU: "UNKNOWN", Quantity: 1, UnitPrice: 1.0}},
		CreatedAt: time.Now(),
	}
	view := ToOrderView(o, "", map[string]string{})
	if view.Items[0].ProductName != "" {
		t.Errorf("expected empty ProductName for unknown SKU, got %q", view.Items[0].ProductName)
	}
}
