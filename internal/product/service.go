package product

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/samber/mo"
)

// Domain errors for mapping to HTTP 400.
var (
	ErrInvalidName     = errors.New("product: name is required")
	ErrInvalidPrice    = errors.New("product: price must be non-negative")
	ErrInvalidObjectID = errors.New("product: invalid object id")
)

// --- Pure Logic ---

// validateCreateInput trims and validates a CreateInput, returning sanitised fields.
func validateCreateInput(in CreateInput) (sku, name string, price float64, err error) {
	name = strings.TrimSpace(in.Name)
	if name == "" {
		return "", "", 0, ErrInvalidName
	}
	if in.Price < 0 {
		return "", "", 0, ErrInvalidPrice
	}
	return strings.TrimSpace(in.SKU), name, in.Price, nil
}

// buildProduct constructs a Product value from validated fields.
func buildProduct(sku, name string, price float64, now time.Time) *Product {
	return &Product{
		SKU:       sku,
		Name:      name,
		Price:     price,
		CreatedAt: now,
	}
}

// --- Orchestration ---

// Service coordinates product use cases.
type Service struct {
	repo Repository
}

// NewService constructs a product service.
func NewService(r Repository) *Service {
	return &Service{repo: r}
}

// Create validates input, builds the domain object, and persists it.
func (s *Service) Create(ctx context.Context, in CreateInput) mo.Result[*Product] {
	sku, name, price, err := validateCreateInput(in)
	if err != nil {
		return mo.Err[*Product](err)
	}
	p := buildProduct(sku, name, price, time.Now().UTC())
	return s.repo.Create(ctx, p)
}

// GetByID loads a product by its hex ID string.
func (s *Service) GetByID(ctx context.Context, id string) mo.Result[mo.Option[Product]] {
	oid, err := parseObjectID(id)
	if err != nil {
		return mo.Err[mo.Option[Product]](err)
	}
	return s.repo.GetByID(ctx, oid)
}

// GetBySKU loads a product by its SKU string.
func (s *Service) GetBySKU(ctx context.Context, sku string) mo.Result[mo.Option[Product]] {
	return s.repo.GetBySKU(ctx, sku)
}

// List returns active products ordered by creation time descending.
func (s *Service) List(ctx context.Context, limit int64) mo.Result[[]Product] {
	return s.repo.List(ctx, limit)
}

// ListInactive returns deactivated products ordered by deactivated_at descending.
func (s *Service) ListInactive(ctx context.Context, limit int64) mo.Result[[]Product] {
	return s.repo.ListInactive(ctx, limit)
}

// Activate re-activates a product.
func (s *Service) Activate(ctx context.Context, id string) mo.Result[*Product] {
	oid, err := parseObjectID(id)
	if err != nil {
		return mo.Err[*Product](err)
	}
	return s.repo.Activate(ctx, oid)
}

// Deactivate soft-deactivates a product.
func (s *Service) Deactivate(ctx context.Context, id string) mo.Result[*Product] {
	oid, err := parseObjectID(id)
	if err != nil {
		return mo.Err[*Product](err)
	}
	return s.repo.Deactivate(ctx, oid, time.Now().UTC())
}
