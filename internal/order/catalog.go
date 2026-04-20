package order

import "context"

// CatalogProduct is a minimal product view used by the order domain for SKU lookup.
type CatalogProduct struct {
	SKU   string  `json:"sku"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

// ProductCatalog is the port through which the order domain fetches and resolves products.
// It is implemented outside this package to avoid import cycles.
type ProductCatalog interface {
	ListActiveProducts(ctx context.Context, limit int64) ([]CatalogProduct, error)
	// ResolveProductName returns the product's name for the given SKU, or "" if not found.
	ResolveProductName(ctx context.Context, sku string) string
}
