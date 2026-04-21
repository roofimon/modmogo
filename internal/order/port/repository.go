package port

import (
	"context"
	"errors"
	"time"

	"github.com/samber/mo"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"modmono/internal/order/domain"
)

// ErrNotFound is returned when no order exists for the given id.
var ErrNotFound = errors.New("order: not found")

// Repository is the PortOut — persistence contract.
type Repository interface {
	Create(ctx context.Context, o *domain.Order) mo.Result[*domain.Order]
	GetByID(ctx context.Context, id primitive.ObjectID) mo.Result[mo.Option[domain.Order]]
	List(ctx context.Context, limit int64) mo.Result[[]domain.Order]
	ListInactive(ctx context.Context, limit int64) mo.Result[[]domain.Order]
	ListPaymentCompleted(ctx context.Context, limit int64) mo.Result[[]domain.Order]
	Deactivate(ctx context.Context, id primitive.ObjectID, at time.Time) mo.Result[*domain.Order]
}

// CatalogProduct is a minimal product view for SKU lookup within the order domain.
type CatalogProduct struct {
	SKU   string  `json:"sku"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

// ProductCatalog is the PortOut for fetching and resolving products.
type ProductCatalog interface {
	ListActiveProducts(ctx context.Context, limit int64) mo.Result[[]CatalogProduct]
	ResolveProductName(ctx context.Context, sku string) string
}

// CatalogCustomer is a minimal customer view for lookup within the order domain.
type CatalogCustomer struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Phone string `json:"phone,omitempty"`
}

// CustomerCatalog is the PortOut for fetching and resolving customers.
type CustomerCatalog interface {
	ListActiveCustomers(ctx context.Context, limit int64) mo.Result[[]CatalogCustomer]
	ResolveCustomerName(ctx context.Context, hexID string) string
}

// UseCase is the PortIn — contract consumed by transport adapters.
type UseCase interface {
	PlaceOrder(ctx context.Context, in domain.CreateInput) mo.Result[*domain.Order]
	ViewOrderDetail(ctx context.Context, id string) mo.Result[mo.Option[domain.OrderView]]
	List(ctx context.Context, limit int64) mo.Result[[]domain.Order]
	ListInactive(ctx context.Context, limit int64) mo.Result[[]domain.Order]
	ListPaymentCompleted(ctx context.Context, limit int64) mo.Result[[]domain.Order]
	Deactivate(ctx context.Context, id string) mo.Result[*domain.Order]
	CompletePayment(ctx context.Context, id string) mo.Result[*domain.Order]
}
