package product

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/samber/mo"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Domain errors for mapping to HTTP 400.
var (
	ErrInvalidName  = errors.New("product: name is required")
	ErrInvalidPrice = errors.New("product: price must be non-negative")
)

// Service coordinates product use cases.
type Service struct {
	repo Repository
}

// NewService constructs a product service.
func NewService(r Repository) *Service {
	return &Service{repo: r}
}

// Create validates input and persists a new product.
func (s *Service) Create(ctx context.Context, in CreateInput) mo.Result[*Product] {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return mo.Err[*Product](ErrInvalidName)
	}
	if in.Price < 0 {
		return mo.Err[*Product](ErrInvalidPrice)
	}
	now := time.Now().UTC()
	p := &Product{
		Name:      name,
		Price:     in.Price,
		CreatedAt: now,
	}
	return s.repo.Create(ctx, p)
}

// ErrInvalidObjectID is returned when the id string is not a valid ObjectID hex.
var ErrInvalidObjectID = errors.New("product: invalid object id")

// GetByID loads a product by identifier.
func (s *Service) GetByID(ctx context.Context, id string) (mo.Option[Product], error) {
	oid, err := parseObjectID(id)
	if err != nil {
		return mo.None[Product](), err
	}
	return s.repo.GetByID(ctx, oid)
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

// List returns products ordered by creation time descending.
func (s *Service) List(ctx context.Context, limit int64) mo.Result[[]Product] {
	return s.repo.List(ctx, limit)
}
