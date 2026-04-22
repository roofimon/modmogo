package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/samber/mo"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"modmono/internal/product/domain"
	"modmono/internal/product/port"
)

// --- mock repository ---

type mockRepo struct {
	createFn      func(context.Context, *domain.Product) mo.Result[*domain.Product]
	getByIDFn     func(context.Context, primitive.ObjectID) mo.Result[mo.Option[domain.Product]]
	getBySKUFn    func(context.Context, string) mo.Result[mo.Option[domain.Product]]
	listFn        func(context.Context, int64) mo.Result[[]domain.Product]
	listInactiveFn func(context.Context, int64) mo.Result[[]domain.Product]
	deactivateFn  func(context.Context, primitive.ObjectID, time.Time) mo.Result[*domain.Product]
	activateFn    func(context.Context, primitive.ObjectID) mo.Result[*domain.Product]
}

func (m *mockRepo) Create(ctx context.Context, p *domain.Product) mo.Result[*domain.Product] {
	if m.createFn == nil {
		panic("unexpected call to Create")
	}
	return m.createFn(ctx, p)
}

func (m *mockRepo) GetByID(ctx context.Context, id primitive.ObjectID) mo.Result[mo.Option[domain.Product]] {
	if m.getByIDFn == nil {
		panic("unexpected call to GetByID")
	}
	return m.getByIDFn(ctx, id)
}

func (m *mockRepo) GetBySKU(ctx context.Context, sku string) mo.Result[mo.Option[domain.Product]] {
	if m.getBySKUFn == nil {
		panic("unexpected call to GetBySKU")
	}
	return m.getBySKUFn(ctx, sku)
}

func (m *mockRepo) List(ctx context.Context, limit int64) mo.Result[[]domain.Product] {
	if m.listFn == nil {
		panic("unexpected call to List")
	}
	return m.listFn(ctx, limit)
}

func (m *mockRepo) ListInactive(ctx context.Context, limit int64) mo.Result[[]domain.Product] {
	if m.listInactiveFn == nil {
		panic("unexpected call to ListInactive")
	}
	return m.listInactiveFn(ctx, limit)
}

func (m *mockRepo) Deactivate(ctx context.Context, id primitive.ObjectID, at time.Time) mo.Result[*domain.Product] {
	if m.deactivateFn == nil {
		panic("unexpected call to Deactivate")
	}
	return m.deactivateFn(ctx, id, at)
}

func (m *mockRepo) Activate(ctx context.Context, id primitive.ObjectID) mo.Result[*domain.Product] {
	if m.activateFn == nil {
		panic("unexpected call to Activate")
	}
	return m.activateFn(ctx, id)
}

var _ port.Repository = (*mockRepo)(nil)

// --- validateCreateInput ---

func TestValidateCreateInput_emptyName(t *testing.T) {
	_, _, _, err := validateCreateInput(domain.CreateInput{SKU: "SKU1", Name: "", Price: 10})
	if !errors.Is(err, ErrInvalidName) {
		t.Errorf("expected ErrInvalidName, got %v", err)
	}
}

func TestValidateCreateInput_whitespaceName(t *testing.T) {
	_, _, _, err := validateCreateInput(domain.CreateInput{SKU: "SKU1", Name: "   ", Price: 10})
	if !errors.Is(err, ErrInvalidName) {
		t.Errorf("expected ErrInvalidName, got %v", err)
	}
}

func TestValidateCreateInput_negativePrice(t *testing.T) {
	_, _, _, err := validateCreateInput(domain.CreateInput{SKU: "SKU1", Name: "Widget", Price: -1})
	if !errors.Is(err, ErrInvalidPrice) {
		t.Errorf("expected ErrInvalidPrice, got %v", err)
	}
}

func TestValidateCreateInput_valid(t *testing.T) {
	sku, name, price, err := validateCreateInput(domain.CreateInput{SKU: "  SKU1  ", Name: "  Widget  ", Price: 9.99})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sku != "SKU1" {
		t.Errorf("sku: expected %q, got %q", "SKU1", sku)
	}
	if name != "Widget" {
		t.Errorf("name: expected %q, got %q", "Widget", name)
	}
	if price != 9.99 {
		t.Errorf("price: expected %v, got %v", 9.99, price)
	}
}

func TestValidateCreateInput_zeroPriceAllowed(t *testing.T) {
	_, _, price, err := validateCreateInput(domain.CreateInput{SKU: "FREE", Name: "Free Item", Price: 0})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if price != 0 {
		t.Errorf("expected price 0, got %v", price)
	}
}

// --- buildProduct ---

func TestBuildProduct_fields(t *testing.T) {
	now := time.Now().UTC()
	p := buildProduct("SKU1", "Widget", 9.99, now)
	if p.SKU != "SKU1" {
		t.Errorf("SKU: expected %q, got %q", "SKU1", p.SKU)
	}
	if p.Name != "Widget" {
		t.Errorf("Name: expected %q, got %q", "Widget", p.Name)
	}
	if p.Price != 9.99 {
		t.Errorf("Price: expected %v, got %v", 9.99, p.Price)
	}
	if !p.CreatedAt.Equal(now) {
		t.Errorf("CreatedAt: expected %v, got %v", now, p.CreatedAt)
	}
	if !p.ID.IsZero() {
		t.Errorf("ID should be zero, got %v", p.ID)
	}
}

// --- Service.Create ---

func TestService_Create_invalidName(t *testing.T) {
	svc := NewService(&mockRepo{})
	res := svc.Create(context.Background(), domain.CreateInput{Name: "", Price: 10})
	if !res.IsError() {
		t.Fatal("expected error result")
	}
	if !errors.Is(res.Error(), ErrInvalidName) {
		t.Errorf("expected ErrInvalidName, got %v", res.Error())
	}
}

func TestService_Create_invalidPrice(t *testing.T) {
	svc := NewService(&mockRepo{})
	res := svc.Create(context.Background(), domain.CreateInput{Name: "Widget", Price: -5})
	if !res.IsError() {
		t.Fatal("expected error result")
	}
	if !errors.Is(res.Error(), ErrInvalidPrice) {
		t.Errorf("expected ErrInvalidPrice, got %v", res.Error())
	}
}

func TestService_Create_valid(t *testing.T) {
	want := &domain.Product{SKU: "SKU1", Name: "Widget", Price: 9.99, Status: domain.StatusActive}
	repo := &mockRepo{
		createFn: func(_ context.Context, p *domain.Product) mo.Result[*domain.Product] {
			return mo.Ok(want)
		},
	}
	svc := NewService(repo)
	res := svc.Create(context.Background(), domain.CreateInput{SKU: "SKU1", Name: "Widget", Price: 9.99})
	if res.IsError() {
		t.Fatalf("unexpected error: %v", res.Error())
	}
	if res.MustGet() != want {
		t.Errorf("expected returned product to be the repo result")
	}
}

// --- Service.ViewProductDetail ---

func TestService_ViewProductDetail_invalidID(t *testing.T) {
	svc := NewService(&mockRepo{})
	res := svc.ViewProductDetail(context.Background(), "not-an-id")
	if !res.IsError() {
		t.Fatal("expected error result")
	}
	if !errors.Is(res.Error(), ErrInvalidObjectID) {
		t.Errorf("expected ErrInvalidObjectID, got %v", res.Error())
	}
}

func TestService_ViewProductDetail_validID(t *testing.T) {
	id := primitive.NewObjectID()
	called := false
	repo := &mockRepo{
		getByIDFn: func(_ context.Context, got primitive.ObjectID) mo.Result[mo.Option[domain.Product]] {
			called = true
			if got != id {
				panic("wrong id passed to GetByID")
			}
			return mo.Ok(mo.None[domain.Product]())
		},
	}
	svc := NewService(repo)
	svc.ViewProductDetail(context.Background(), id.Hex())
	if !called {
		t.Error("expected repo.GetByID to be called")
	}
}

// --- Service.Activate ---

func TestService_Activate_invalidID(t *testing.T) {
	svc := NewService(&mockRepo{})
	res := svc.Activate(context.Background(), "bad")
	if !res.IsError() || !errors.Is(res.Error(), ErrInvalidObjectID) {
		t.Errorf("expected ErrInvalidObjectID, got %v", res.Error())
	}
}

func TestService_Activate_validID(t *testing.T) {
	id := primitive.NewObjectID()
	called := false
	repo := &mockRepo{
		activateFn: func(_ context.Context, got primitive.ObjectID) mo.Result[*domain.Product] {
			called = true
			if got != id {
				panic("wrong id")
			}
			return mo.Ok(&domain.Product{})
		},
	}
	svc := NewService(repo)
	svc.Activate(context.Background(), id.Hex())
	if !called {
		t.Error("expected repo.Activate to be called")
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
		deactivateFn: func(_ context.Context, got primitive.ObjectID, _ time.Time) mo.Result[*domain.Product] {
			called = true
			if got != id {
				panic("wrong id")
			}
			return mo.Ok(&domain.Product{})
		},
	}
	svc := NewService(repo)
	svc.Deactivate(context.Background(), id.Hex())
	if !called {
		t.Error("expected repo.Deactivate to be called")
	}
}
