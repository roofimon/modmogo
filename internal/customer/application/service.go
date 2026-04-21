package application

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/samber/mo"

	"modmono/internal/customer/domain"
	"modmono/internal/customer/port"
)

// Domain errors for mapping to HTTP 400.
var (
	ErrInvalidName     = errors.New("customer: name is required")
	ErrInvalidEmail    = errors.New("customer: email is required")
	ErrInvalidObjectID = errors.New("customer: invalid object id")
)

// --- Pure Logic ---

// validateCreateInput trims and validates a CreateInput, returning sanitised fields.
func validateCreateInput(in domain.CreateInput) (name, email, phone string, err error) {
	name = strings.TrimSpace(in.Name)
	if name == "" {
		return "", "", "", ErrInvalidName
	}
	email = strings.TrimSpace(in.Email)
	if email == "" {
		return "", "", "", ErrInvalidEmail
	}
	phone = strings.TrimSpace(in.Phone)
	return name, email, phone, nil
}

// buildCustomer constructs a Customer value from validated fields.
func buildCustomer(name, email, phone string, now time.Time) *domain.Customer {
	return &domain.Customer{
		Name:      name,
		Email:     email,
		Phone:     phone,
		CreatedAt: now,
	}
}

// --- Orchestration ---

// Service coordinates customer use cases.
type Service struct {
	repo port.Repository
}

// NewService constructs a customer service.
func NewService(r port.Repository) *Service {
	return &Service{repo: r}
}

// Create validates input, builds the domain object, and persists it.
func (s *Service) Create(ctx context.Context, in domain.CreateInput) mo.Result[*domain.Customer] {
	name, email, phone, err := validateCreateInput(in)
	if err != nil {
		return mo.Err[*domain.Customer](err)
	}
	c := buildCustomer(name, email, phone, time.Now().UTC())
	return s.repo.Create(ctx, c)
}

// GetByID loads a customer by its hex ID string.
func (s *Service) GetByID(ctx context.Context, id string) mo.Result[mo.Option[domain.Customer]] {
	oid, err := parseObjectID(id)
	if err != nil {
		return mo.Err[mo.Option[domain.Customer]](err)
	}
	return s.repo.GetByID(ctx, oid)
}

// List returns active customers ordered by creation time descending.
func (s *Service) List(ctx context.Context, limit int64) mo.Result[[]domain.Customer] {
	return s.repo.List(ctx, limit)
}

// ListInactive returns deactivated customers ordered by deactivated_at descending.
func (s *Service) ListInactive(ctx context.Context, limit int64) mo.Result[[]domain.Customer] {
	return s.repo.ListInactive(ctx, limit)
}

// Deactivate soft-deactivates a customer.
func (s *Service) Deactivate(ctx context.Context, id string) mo.Result[*domain.Customer] {
	oid, err := parseObjectID(id)
	if err != nil {
		return mo.Err[*domain.Customer](err)
	}
	return s.repo.Deactivate(ctx, oid, time.Now().UTC())
}
