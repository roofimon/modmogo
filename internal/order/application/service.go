package application

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/samber/mo"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"modmono/internal/order/domain"
	"modmono/internal/order/port"
)

// Domain errors — map to HTTP 400.
var (
	ErrNoItems           = errors.New("order: at least one line item is required")
	ErrInvalidSKU        = errors.New("order: each line item must have a non-empty sku")
	ErrInvalidQuantity   = errors.New("order: each line item quantity must be >= 1")
	ErrInvalidUnitPrice  = errors.New("order: each line item unit_price must be >= 0")
	ErrInvalidObjectID   = errors.New("order: invalid object id")
	ErrInvalidCustomerID = errors.New("order: customer_id is not a valid object id")
	ErrAlreadyCompleted  = errors.New("order: payment already completed")
)

// --- Orchestration ---

// Service coordinates order use cases.
type Service struct {
	repo      port.Repository
	products  port.ProductCatalog
	customers port.CustomerCatalog
}

var _ port.UseCase = (*Service)(nil)

// NewService constructs an order service.
func NewService(r port.Repository, products port.ProductCatalog, customers port.CustomerCatalog) *Service {
	return &Service{repo: r, products: products, customers: customers}
}

// PlaceOrder validates input and persists a new order.
func (s *Service) PlaceOrder(ctx context.Context, in domain.CreateInput) mo.Result[*domain.Order] {
	if len(in.Items) == 0 {
		return mo.Err[*domain.Order](ErrNoItems)
	}
	items := make([]domain.LineItem, 0, len(in.Items))
	for _, raw := range in.Items {
		sku := strings.TrimSpace(raw.SKU)
		if sku == "" {
			return mo.Err[*domain.Order](ErrInvalidSKU)
		}
		if raw.Quantity < 1 {
			return mo.Err[*domain.Order](ErrInvalidQuantity)
		}
		if raw.UnitPrice < 0 {
			return mo.Err[*domain.Order](ErrInvalidUnitPrice)
		}
		items = append(items, domain.LineItem{SKU: sku, Quantity: raw.Quantity, UnitPrice: raw.UnitPrice})
	}

	var customerOID *primitive.ObjectID
	if in.CustomerID != nil {
		oid, err := parseObjectID(*in.CustomerID)
		if err != nil {
			return mo.Err[*domain.Order](ErrInvalidCustomerID)
		}
		customerOID = &oid
	}

	o := &domain.Order{
		CustomerID: customerOID,
		Items:      items,
		CreatedAt:  time.Now().UTC(),
	}
	return s.repo.Create(ctx, o)
}

// ViewOrderDetail loads an order by its hex ID and enriches it with customer and product names.
func (s *Service) ViewOrderDetail(ctx context.Context, id string) mo.Result[mo.Option[domain.OrderView]] {
	oid, err := parseObjectID(id)
	if err != nil {
		return mo.Err[mo.Option[domain.OrderView]](err)
	}
	res := s.repo.GetByID(ctx, oid)
	if res.IsError() {
		return mo.Err[mo.Option[domain.OrderView]](res.Error())
	}
	if res.MustGet().IsAbsent() {
		return mo.Ok(mo.None[domain.OrderView]())
	}
	o, _ := res.MustGet().Get()
	return mo.Ok(mo.Some(s.enrichOrder(ctx, o)))
}

// enrichOrder resolves customer and product names and builds an OrderView.
func (s *Service) enrichOrder(ctx context.Context, o domain.Order) domain.OrderView {
	customerName := ""
	if o.CustomerID != nil {
		customerName = s.customers.ResolveCustomerName(ctx, o.CustomerID.Hex())
	}
	productNames := make(map[string]string)
	for _, item := range o.Items {
		if _, seen := productNames[item.SKU]; !seen {
			productNames[item.SKU] = s.products.ResolveProductName(ctx, item.SKU)
		}
	}
	return domain.ToOrderView(o, customerName, productNames)
}

// List returns pending orders ordered by creation time descending.
func (s *Service) List(ctx context.Context, limit int64) mo.Result[[]domain.Order] {
	return s.repo.List(ctx, limit)
}

// ListInactive returns deactivated orders.
func (s *Service) ListInactive(ctx context.Context, limit int64) mo.Result[[]domain.Order] {
	return s.repo.ListInactive(ctx, limit)
}

// ListPaymentCompleted returns payment-completed orders.
func (s *Service) ListPaymentCompleted(ctx context.Context, limit int64) mo.Result[[]domain.Order] {
	return s.repo.ListPaymentCompleted(ctx, limit)
}

// Deactivate soft-deactivates an order.
func (s *Service) Deactivate(ctx context.Context, id string) mo.Result[*domain.Order] {
	oid, err := parseObjectID(id)
	if err != nil {
		return mo.Err[*domain.Order](err)
	}
	return s.repo.Deactivate(ctx, oid, time.Now().UTC())
}

// CompletePayment creates a new payment-completed order derived from the given order ID.
func (s *Service) CompletePayment(ctx context.Context, id string) mo.Result[*domain.Order] {
	oid, err := parseObjectID(id)
	if err != nil {
		return mo.Err[*domain.Order](err)
	}
	res := s.repo.GetByID(ctx, oid)
	if res.IsError() {
		return mo.Err[*domain.Order](res.Error())
	}
	opt := res.MustGet()
	if opt.IsAbsent() {
		return mo.Err[*domain.Order](port.ErrNotFound)
	}
	orig, _ := opt.Get()
	if orig.Status == domain.StatusPaymentCompleted {
		return mo.Err[*domain.Order](ErrAlreadyCompleted)
	}
	newOrder := &domain.Order{
		CustomerID:      orig.CustomerID,
		Items:           orig.Items,
		Status:          domain.StatusPaymentCompleted,
		OriginalOrderID: &orig.ID,
		CreatedAt:       time.Now().UTC(),
	}
	return s.repo.Create(ctx, newOrder)
}
