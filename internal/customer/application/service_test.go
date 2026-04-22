package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/samber/mo"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"modmono/internal/customer/domain"
	"modmono/internal/customer/port"
)

// --- mock repository ---

type mockRepo struct {
	createFn       func(context.Context, *domain.Customer) mo.Result[*domain.Customer]
	getByIDFn      func(context.Context, primitive.ObjectID) mo.Result[mo.Option[domain.Customer]]
	listFn         func(context.Context, int64) mo.Result[[]domain.Customer]
	listInactiveFn func(context.Context, int64) mo.Result[[]domain.Customer]
	deactivateFn   func(context.Context, primitive.ObjectID, time.Time) mo.Result[*domain.Customer]
}

func (m *mockRepo) Create(ctx context.Context, c *domain.Customer) mo.Result[*domain.Customer] {
	if m.createFn == nil {
		panic("unexpected call to Create")
	}
	return m.createFn(ctx, c)
}

func (m *mockRepo) GetByID(ctx context.Context, id primitive.ObjectID) mo.Result[mo.Option[domain.Customer]] {
	if m.getByIDFn == nil {
		panic("unexpected call to GetByID")
	}
	return m.getByIDFn(ctx, id)
}

func (m *mockRepo) List(ctx context.Context, limit int64) mo.Result[[]domain.Customer] {
	if m.listFn == nil {
		panic("unexpected call to List")
	}
	return m.listFn(ctx, limit)
}

func (m *mockRepo) ListInactive(ctx context.Context, limit int64) mo.Result[[]domain.Customer] {
	if m.listInactiveFn == nil {
		panic("unexpected call to ListInactive")
	}
	return m.listInactiveFn(ctx, limit)
}

func (m *mockRepo) Deactivate(ctx context.Context, id primitive.ObjectID, at time.Time) mo.Result[*domain.Customer] {
	if m.deactivateFn == nil {
		panic("unexpected call to Deactivate")
	}
	return m.deactivateFn(ctx, id, at)
}

var _ port.Repository = (*mockRepo)(nil)

// --- validateCreateInput ---

func TestValidateCreateInput_emptyName(t *testing.T) {
	_, _, _, err := validateCreateInput(domain.CreateInput{Name: "", Email: "a@b.com"})
	if !errors.Is(err, ErrInvalidName) {
		t.Errorf("expected ErrInvalidName, got %v", err)
	}
}

func TestValidateCreateInput_whitespaceName(t *testing.T) {
	_, _, _, err := validateCreateInput(domain.CreateInput{Name: "   ", Email: "a@b.com"})
	if !errors.Is(err, ErrInvalidName) {
		t.Errorf("expected ErrInvalidName, got %v", err)
	}
}

func TestValidateCreateInput_emptyEmail(t *testing.T) {
	_, _, _, err := validateCreateInput(domain.CreateInput{Name: "Alice", Email: ""})
	if !errors.Is(err, ErrInvalidEmail) {
		t.Errorf("expected ErrInvalidEmail, got %v", err)
	}
}

func TestValidateCreateInput_whitespaceEmail(t *testing.T) {
	_, _, _, err := validateCreateInput(domain.CreateInput{Name: "Alice", Email: "   "})
	if !errors.Is(err, ErrInvalidEmail) {
		t.Errorf("expected ErrInvalidEmail, got %v", err)
	}
}

func TestValidateCreateInput_valid(t *testing.T) {
	name, email, phone, err := validateCreateInput(domain.CreateInput{
		Name: "  Alice  ", Email: "  alice@example.com  ", Phone: "  0812345678  ",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "Alice" {
		t.Errorf("name: expected %q, got %q", "Alice", name)
	}
	if email != "alice@example.com" {
		t.Errorf("email: expected %q, got %q", "alice@example.com", email)
	}
	if phone != "0812345678" {
		t.Errorf("phone: expected %q, got %q", "0812345678", phone)
	}
}

func TestValidateCreateInput_phoneOptional(t *testing.T) {
	_, _, phone, err := validateCreateInput(domain.CreateInput{Name: "Alice", Email: "a@b.com", Phone: ""})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if phone != "" {
		t.Errorf("expected empty phone, got %q", phone)
	}
}

// --- buildCustomer ---

func TestBuildCustomer_fields(t *testing.T) {
	now := time.Now().UTC()
	c := buildCustomer("Alice", "alice@example.com", "0812345678", now)
	if c.Name != "Alice" {
		t.Errorf("Name: expected %q, got %q", "Alice", c.Name)
	}
	if c.Email != "alice@example.com" {
		t.Errorf("Email: expected %q, got %q", "alice@example.com", c.Email)
	}
	if c.Phone != "0812345678" {
		t.Errorf("Phone: expected %q, got %q", "0812345678", c.Phone)
	}
	if !c.CreatedAt.Equal(now) {
		t.Errorf("CreatedAt: expected %v, got %v", now, c.CreatedAt)
	}
	if !c.ID.IsZero() {
		t.Errorf("ID should be zero, got %v", c.ID)
	}
}

// --- Service.Create ---

func TestService_Create_invalidName(t *testing.T) {
	svc := NewService(&mockRepo{})
	res := svc.Create(context.Background(), domain.CreateInput{Name: "", Email: "a@b.com"})
	if !res.IsError() || !errors.Is(res.Error(), ErrInvalidName) {
		t.Errorf("expected ErrInvalidName, got %v", res.Error())
	}
}

func TestService_Create_invalidEmail(t *testing.T) {
	svc := NewService(&mockRepo{})
	res := svc.Create(context.Background(), domain.CreateInput{Name: "Alice", Email: ""})
	if !res.IsError() || !errors.Is(res.Error(), ErrInvalidEmail) {
		t.Errorf("expected ErrInvalidEmail, got %v", res.Error())
	}
}

func TestService_Create_valid(t *testing.T) {
	want := &domain.Customer{Name: "Alice", Email: "alice@example.com", Status: domain.StatusActive}
	repo := &mockRepo{
		createFn: func(_ context.Context, _ *domain.Customer) mo.Result[*domain.Customer] {
			return mo.Ok(want)
		},
	}
	svc := NewService(repo)
	res := svc.Create(context.Background(), domain.CreateInput{Name: "Alice", Email: "alice@example.com"})
	if res.IsError() {
		t.Fatalf("unexpected error: %v", res.Error())
	}
	if res.MustGet() != want {
		t.Error("expected returned customer to be the repo result")
	}
}

// --- Service.ViewCustomerDetail ---

func TestService_ViewCustomerDetail_invalidID(t *testing.T) {
	svc := NewService(&mockRepo{})
	res := svc.ViewCustomerDetail(context.Background(), "bad-id")
	if !res.IsError() || !errors.Is(res.Error(), ErrInvalidObjectID) {
		t.Errorf("expected ErrInvalidObjectID, got %v", res.Error())
	}
}

func TestService_ViewCustomerDetail_validID(t *testing.T) {
	id := primitive.NewObjectID()
	called := false
	repo := &mockRepo{
		getByIDFn: func(_ context.Context, got primitive.ObjectID) mo.Result[mo.Option[domain.Customer]] {
			called = true
			if got != id {
				panic("wrong id passed to GetByID")
			}
			return mo.Ok(mo.None[domain.Customer]())
		},
	}
	svc := NewService(repo)
	svc.ViewCustomerDetail(context.Background(), id.Hex())
	if !called {
		t.Error("expected repo.GetByID to be called")
	}
}

// --- Service.Deactivate ---

func TestService_Deactivate_invalidID(t *testing.T) {
	svc := NewService(&mockRepo{})
	res := svc.Deactivate(context.Background(), "bad")
	if !res.IsError() || !errors.Is(res.Error(), ErrInvalidObjectID) {
		t.Errorf("expected ErrInvalidObjectID, got %v", res.Error())
	}
}

func TestService_Deactivate_validID(t *testing.T) {
	id := primitive.NewObjectID()
	called := false
	repo := &mockRepo{
		deactivateFn: func(_ context.Context, got primitive.ObjectID, _ time.Time) mo.Result[*domain.Customer] {
			called = true
			if got != id {
				panic("wrong id")
			}
			return mo.Ok(&domain.Customer{})
		},
	}
	svc := NewService(repo)
	svc.Deactivate(context.Background(), id.Hex())
	if !called {
		t.Error("expected repo.Deactivate to be called")
	}
}
