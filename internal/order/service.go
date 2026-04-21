package order

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/samber/mo"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

// Service coordinates order use cases.
type Service struct {
	repo      Repository
	products  ProductCatalog
	customers CustomerCatalog
}

// NewService constructs an order service.
func NewService(r Repository, products ProductCatalog, customers CustomerCatalog) *Service {
	return &Service{repo: r, products: products, customers: customers}
}

// Create validates input and persists a new order.
func (s *Service) Create(ctx context.Context, in CreateInput) mo.Result[*Order] {
	if len(in.Items) == 0 {
		return mo.Err[*Order](ErrNoItems)
	}
	items := make([]LineItem, 0, len(in.Items))
	for _, raw := range in.Items {
		sku := strings.TrimSpace(raw.SKU)
		if sku == "" {
			return mo.Err[*Order](ErrInvalidSKU)
		}
		if raw.Quantity < 1 {
			return mo.Err[*Order](ErrInvalidQuantity)
		}
		if raw.UnitPrice < 0 {
			return mo.Err[*Order](ErrInvalidUnitPrice)
		}
		items = append(items, LineItem{SKU: sku, Quantity: raw.Quantity, UnitPrice: raw.UnitPrice})
	}

	var customerOID *primitive.ObjectID
	if in.CustomerID != nil {
		oid, err := parseObjectID(*in.CustomerID)
		if err != nil {
			return mo.Err[*Order](ErrInvalidCustomerID)
		}
		customerOID = &oid
	}

	o := &Order{
		CustomerID: customerOID,
		Items:      items,
		CreatedAt:  time.Now().UTC(),
	}
	return s.repo.Create(ctx, o)
}

// GetByID loads an order by its hex ID and enriches it with customer and product names.
func (s *Service) GetByID(ctx context.Context, id string) mo.Result[mo.Option[OrderView]] {
	oid, err := parseObjectID(id)
	if err != nil {
		return mo.Err[mo.Option[OrderView]](err)
	}
	res := s.repo.GetByID(ctx, oid)
	if res.IsError() {
		return mo.Err[mo.Option[OrderView]](res.Error())
	}
	if res.MustGet().IsAbsent() {
		return mo.Ok(mo.None[OrderView]())
	}
	o, _ := res.MustGet().Get()
	return mo.Ok(mo.Some(s.enrichOrder(ctx, o)))
}

// enrichOrder resolves customer and product names and builds an OrderView.
func (s *Service) enrichOrder(ctx context.Context, o Order) OrderView {
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
	return toOrderView(o, customerName, productNames)
}

// List returns active orders ordered by creation time descending.
func (s *Service) List(ctx context.Context, limit int64) mo.Result[[]Order] {
	return s.repo.List(ctx, limit)
}

// ListInactive returns deactivated orders ordered by deactivated_at descending.
func (s *Service) ListInactive(ctx context.Context, limit int64) mo.Result[[]Order] {
	return s.repo.ListInactive(ctx, limit)
}

// ListPaymentCompleted returns payment-completed orders.
func (s *Service) ListPaymentCompleted(ctx context.Context, limit int64) mo.Result[[]Order] {
	return s.repo.ListPaymentCompleted(ctx, limit)
}

// Deactivate soft-deactivates an order.
func (s *Service) Deactivate(ctx context.Context, id string) mo.Result[*Order] {
	oid, err := parseObjectID(id)
	if err != nil {
		return mo.Err[*Order](err)
	}
	return s.repo.Deactivate(ctx, oid, time.Now().UTC())
}

// CompletePayment creates a new payment-completed order derived from the given order ID.
func (s *Service) CompletePayment(ctx context.Context, id string) mo.Result[*Order] {
	oid, err := parseObjectID(id)
	if err != nil {
		return mo.Err[*Order](err)
	}
	res := s.repo.GetByID(ctx, oid)
	if res.IsError() {
		return mo.Err[*Order](res.Error())
	}
	opt := res.MustGet()
	if opt.IsAbsent() {
		return mo.Err[*Order](ErrNotFound)
	}
	orig, _ := opt.Get()
	if orig.Status == StatusPaymentCompleted {
		return mo.Err[*Order](ErrAlreadyCompleted)
	}
	newOrder := &Order{
		CustomerID:      orig.CustomerID,
		Items:           orig.Items,
		Status:          StatusPaymentCompleted,
		OriginalOrderID: &orig.ID,
		CreatedAt:       time.Now().UTC(),
	}
	return s.repo.Create(ctx, newOrder)
}

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
