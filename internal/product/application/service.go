package application

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/samber/mo"

	"modmono/internal/product/domain"
	"modmono/internal/product/port"
)

// Domain errors for mapping to HTTP 400.
var (
	ErrInvalidName     = errors.New("product: name is required")
	ErrInvalidPrice    = errors.New("product: price must be non-negative")
	ErrInvalidObjectID = errors.New("product: invalid object id")
)

// --- Pure Logic ---

// validateCreateInput trims and validates a CreateInput, returning sanitised fields.
func validateCreateInput(in domain.CreateInput) (sku, name string, price float64, err error) {
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
func buildProduct(sku, name string, price float64, now time.Time) *domain.Product {
	return &domain.Product{
		SKU:       sku,
		Name:      name,
		Price:     price,
		CreatedAt: now,
	}
}

// --- Orchestration ---

// Service coordinates product use cases.
type Service struct {
	repo port.Repository
}

var _ port.UseCase = (*Service)(nil)

// NewService constructs a product service.
func NewService(r port.Repository) *Service {
	return &Service{repo: r}
}

// Create validates input, builds the domain object, and persists it.
func (s *Service) Create(ctx context.Context, in domain.CreateInput) mo.Result[*domain.Product] {
	sku, name, price, err := validateCreateInput(in)
	if err != nil {
		return mo.Err[*domain.Product](err)
	}
	p := buildProduct(sku, name, price, time.Now().UTC())
	return s.repo.Create(ctx, p)
}

// ViewProductDetail loads a product by its hex ID string.
func (s *Service) ViewProductDetail(ctx context.Context, id string) mo.Result[mo.Option[domain.Product]] {
	oid, err := parseObjectID(id)
	if err != nil {
		return mo.Err[mo.Option[domain.Product]](err)
	}
	return s.repo.GetByID(ctx, oid)
}

// FindProductBySKU loads a product by its SKU string.
func (s *Service) FindProductBySKU(ctx context.Context, sku string) mo.Result[mo.Option[domain.Product]] {
	return s.repo.GetBySKU(ctx, sku)
}

// List returns active products ordered by creation time descending.
func (s *Service) List(ctx context.Context, limit int64) mo.Result[[]domain.Product] {
	return s.repo.List(ctx, limit)
}

// ListInactive returns deactivated products ordered by deactivated_at descending.
func (s *Service) ListInactive(ctx context.Context, limit int64) mo.Result[[]domain.Product] {
	return s.repo.ListInactive(ctx, limit)
}

// Activate re-activates a product.
func (s *Service) Activate(ctx context.Context, id string) mo.Result[*domain.Product] {
	oid, err := parseObjectID(id)
	if err != nil {
		return mo.Err[*domain.Product](err)
	}
	return s.repo.Activate(ctx, oid)
}

// Deactivate soft-deactivates a product.
func (s *Service) Deactivate(ctx context.Context, id string) mo.Result[*domain.Product] {
	oid, err := parseObjectID(id)
	if err != nil {
		return mo.Err[*domain.Product](err)
	}
	return s.repo.Deactivate(ctx, oid, time.Now().UTC())
}
