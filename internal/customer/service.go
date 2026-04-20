package customer

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
	ErrInvalidName  = errors.New("customer: name is required")
	ErrInvalidEmail = errors.New("customer: email is required")
)

// Service coordinates customer use cases.
type Service struct {
	repo Repository
}

// NewService constructs a customer service.
func NewService(r Repository) *Service {
	return &Service{repo: r}
}

// Create validates input and persists a new customer.
func (s *Service) Create(ctx context.Context, in CreateInput) mo.Result[*Customer] {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return mo.Err[*Customer](ErrInvalidName)
	}
	email := strings.TrimSpace(in.Email)
	if email == "" {
		return mo.Err[*Customer](ErrInvalidEmail)
	}
	phone := strings.TrimSpace(in.Phone)
	now := time.Now().UTC()
	c := &Customer{
		Name:      name,
		Email:     email,
		Phone:     phone,
		CreatedAt: now,
	}
	return s.repo.Create(ctx, c)
}

// ErrInvalidObjectID is returned when the id string is not a valid ObjectID hex.
var ErrInvalidObjectID = errors.New("customer: invalid object id")

// GetByID loads a customer by identifier.
func (s *Service) GetByID(ctx context.Context, id string) (mo.Option[Customer], error) {
	oid, err := parseObjectID(id)
	if err != nil {
		return mo.None[Customer](), err
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

// List returns active customers ordered by creation time descending.
func (s *Service) List(ctx context.Context, limit int64) mo.Result[[]Customer] {
	return s.repo.List(ctx, limit)
}

// ListInactive returns deactivated customers ordered by deactivated_at descending.
func (s *Service) ListInactive(ctx context.Context, limit int64) mo.Result[[]Customer] {
	return s.repo.ListInactive(ctx, limit)
}

// Deactivate soft-deactivates a customer. Idempotent: returns the updated document each time.
func (s *Service) Deactivate(ctx context.Context, id string) mo.Result[*Customer] {
	oid, err := parseObjectID(id)
	if err != nil {
		return mo.Err[*Customer](err)
	}
	return s.repo.Deactivate(ctx, oid, time.Now().UTC())
}
