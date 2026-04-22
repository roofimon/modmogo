package domain

import "testing"

func TestComputeTotal_singleItem(t *testing.T) {
	o := &Order{Items: []LineItem{{SKU: "A", Quantity: 2, UnitPrice: 5.0}}}
	o.ComputeTotal()
	if o.Total != 10.0 {
		t.Errorf("expected 10.0, got %v", o.Total)
	}
}

func TestComputeTotal_multipleItems(t *testing.T) {
	o := &Order{Items: []LineItem{
		{SKU: "A", Quantity: 2, UnitPrice: 5.0},
		{SKU: "B", Quantity: 3, UnitPrice: 4.0},
	}}
	o.ComputeTotal()
	if o.Total != 22.0 {
		t.Errorf("expected 22.0, got %v", o.Total)
	}
}

func TestComputeTotal_noItems(t *testing.T) {
	o := &Order{}
	o.ComputeTotal()
	if o.Total != 0 {
		t.Errorf("expected 0, got %v", o.Total)
	}
}
