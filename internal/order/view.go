package order

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// LineItemView is a line item enriched with a resolved product name.
type LineItemView struct {
	SKU         string  `json:"sku"`
	ProductName string  `json:"product_name,omitempty"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
}

// OrderView is the read model returned by GetByID, enriched with display names.
type OrderView struct {
	ID              primitive.ObjectID  `json:"id"`
	CustomerID      *primitive.ObjectID `json:"customer_id,omitempty"`
	CustomerName    string              `json:"customer_name,omitempty"`
	Items           []LineItemView      `json:"items"`
	Total           float64             `json:"total"`
	Status          string              `json:"status,omitempty"`
	OriginalOrderID *primitive.ObjectID `json:"original_order_id,omitempty"`
	CreatedAt       time.Time           `json:"created_at"`
	DeactivatedAt   *time.Time          `json:"deactivated_at,omitempty"`
}

// toOrderView converts a domain Order into an enriched OrderView.
func toOrderView(o Order, customerName string, productNames map[string]string) OrderView {
	items := make([]LineItemView, len(o.Items))
	for i, item := range o.Items {
		items[i] = LineItemView{
			SKU:         item.SKU,
			ProductName: productNames[item.SKU],
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
		}
	}
	return OrderView{
		ID:              o.ID,
		CustomerID:      o.CustomerID,
		CustomerName:    customerName,
		Items:           items,
		Total:           o.Total,
		Status:          o.Status,
		OriginalOrderID: o.OriginalOrderID,
		CreatedAt:       o.CreatedAt,
		DeactivatedAt:   o.DeactivatedAt,
	}
}
