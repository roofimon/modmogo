package order

import (
	"context"

	"github.com/samber/mo"
)

// CatalogCustomer is a minimal customer view used by the order domain for customer lookup.
type CatalogCustomer struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Phone string `json:"phone,omitempty"`
}

// CustomerCatalog is the port through which the order domain fetches and resolves customers.
// It is implemented outside this package to avoid import cycles.
type CustomerCatalog interface {
	ListActiveCustomers(ctx context.Context, limit int64) mo.Result[[]CatalogCustomer]
	// ResolveCustomerName returns the customer's name for the given hex ObjectID, or "" if not found.
	ResolveCustomerName(ctx context.Context, hexID string) string
}
