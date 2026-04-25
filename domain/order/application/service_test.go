package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/samber/mo"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"modmono/domain/order/domain"
	"modmono/domain/order/port"
	"modmono/domain/platform/event"
)

type noopPublisher struct{}

func (noopPublisher) Publish(_ context.Context, _ event.Event) {}

// --- mock repository ---

type mockRepo struct {
	createFn              func(context.Context, *domain.Order) mo.Result[*domain.Order]
	getByIDFn             func(context.Context, primitive.ObjectID) mo.Result[mo.Option[domain.Order]]
	listFn                func(context.Context, int64) mo.Result[[]domain.Order]
	listInactiveFn        func(context.Context, int64) mo.Result[[]domain.Order]
	listPaymentCompletedFn func(context.Context, int64) mo.Result[[]domain.Order]
	deactivateFn          func(context.Context, primitive.ObjectID, time.Time) mo.Result[*domain.Order]
}

func (m *mockRepo) Create(ctx context.Context, o *domain.Order) mo.Result[*domain.Order] {
	if m.createFn == nil {
		panic("unexpected call to Create")
	}
	return m.createFn(ctx, o)
}

func (m *mockRepo) GetByID(ctx context.Context, id primitive.ObjectID) mo.Result[mo.Option[domain.Order]] {
	if m.getByIDFn == nil {
		panic("unexpected call to GetByID")
	}
	return m.getByIDFn(ctx, id)
}

func (m *mockRepo) List(ctx context.Context, limit int64) mo.Result[[]domain.Order] {
	if m.listFn == nil {
		panic("unexpected call to List")
	}
	return m.listFn(ctx, limit)
}

func (m *mockRepo) ListInactive(ctx context.Context, limit int64) mo.Result[[]domain.Order] {
	if m.listInactiveFn == nil {
		panic("unexpected call to ListInactive")
	}
	return m.listInactiveFn(ctx, limit)
}

func (m *mockRepo) ListPaymentCompleted(ctx context.Context, limit int64) mo.Result[[]domain.Order] {
	if m.listPaymentCompletedFn == nil {
		panic("unexpected call to ListPaymentCompleted")
	}
	return m.listPaymentCompletedFn(ctx, limit)
}

func (m *mockRepo) Deactivate(ctx context.Context, id primitive.ObjectID, at time.Time) mo.Result[*domain.Order] {
	if m.deactivateFn == nil {
		panic("unexpected call to Deactivate")
	}
	return m.deactivateFn(ctx, id, at)
}

var _ port.Repository = (*mockRepo)(nil)

// --- mock catalogs ---

type mockProductCatalog struct {
	listFn    func(context.Context, int64) mo.Result[[]port.CatalogProduct]
	resolveFn func(context.Context, string) string
}

func (m *mockProductCatalog) ListActiveProducts(ctx context.Context, limit int64) mo.Result[[]port.CatalogProduct] {
	if m.listFn == nil {
		return mo.Ok([]port.CatalogProduct{})
	}
	return m.listFn(ctx, limit)
}

func (m *mockProductCatalog) ResolveProductName(_ context.Context, sku string) string {
	if m.resolveFn == nil {
		return ""
	}
	return m.resolveFn(context.Background(), sku)
}

type mockCustomerCatalog struct {
	listFn    func(context.Context, int64) mo.Result[[]port.CatalogCustomer]
	resolveFn func(context.Context, string) string
}

func (m *mockCustomerCatalog) ListActiveCustomers(ctx context.Context, limit int64) mo.Result[[]port.CatalogCustomer] {
	if m.listFn == nil {
		return mo.Ok([]port.CatalogCustomer{})
	}
	return m.listFn(ctx, limit)
}

func (m *mockCustomerCatalog) ResolveCustomerName(_ context.Context, hexID string) string {
	if m.resolveFn == nil {
		return ""
	}
	return m.resolveFn(context.Background(), hexID)
}

func newSvc(repo port.Repository) *Service {
	return NewService(repo, &mockProductCatalog{}, &mockCustomerCatalog{}, noopPublisher{})
}

// --- PlaceOrder validation ---

func TestService_PlaceOrder_noItems(t *testing.T) {
	svc := newSvc(&mockRepo{})
	res := svc.PlaceOrder(context.Background(), domain.CreateInput{Items: nil})
	if !res.IsError() || !errors.Is(res.Error(), ErrNoItems) {
		t.Errorf("expected ErrNoItems, got %v", res.Error())
	}
}

func TestService_PlaceOrder_emptySKU(t *testing.T) {
	svc := newSvc(&mockRepo{})
	res := svc.PlaceOrder(context.Background(), domain.CreateInput{
		Items: []domain.LineItemInput{{SKU: "  ", Quantity: 1, UnitPrice: 5.0}},
	})
	if !res.IsError() || !errors.Is(res.Error(), ErrInvalidSKU) {
		t.Errorf("expected ErrInvalidSKU, got %v", res.Error())
	}
}

func TestService_PlaceOrder_zeroQuantity(t *testing.T) {
	svc := newSvc(&mockRepo{})
	res := svc.PlaceOrder(context.Background(), domain.CreateInput{
		Items: []domain.LineItemInput{{SKU: "SKU1", Quantity: 0, UnitPrice: 5.0}},
	})
	if !res.IsError() || !errors.Is(res.Error(), ErrInvalidQuantity) {
		t.Errorf("expected ErrInvalidQuantity, got %v", res.Error())
	}
}

func TestService_PlaceOrder_negativeUnitPrice(t *testing.T) {
	svc := newSvc(&mockRepo{})
	res := svc.PlaceOrder(context.Background(), domain.CreateInput{
		Items: []domain.LineItemInput{{SKU: "SKU1", Quantity: 1, UnitPrice: -1}},
	})
	if !res.IsError() || !errors.Is(res.Error(), ErrInvalidUnitPrice) {
		t.Errorf("expected ErrInvalidUnitPrice, got %v", res.Error())
	}
}

func TestService_PlaceOrder_invalidCustomerID(t *testing.T) {
	svc := newSvc(&mockRepo{})
	bad := "not-an-id"
	res := svc.PlaceOrder(context.Background(), domain.CreateInput{
		CustomerID: &bad,
		Items:      []domain.LineItemInput{{SKU: "SKU1", Quantity: 1, UnitPrice: 5.0}},
	})
	if !res.IsError() || !errors.Is(res.Error(), ErrInvalidCustomerID) {
		t.Errorf("expected ErrInvalidCustomerID, got %v", res.Error())
	}
}

func TestService_PlaceOrder_valid(t *testing.T) {
	want := &domain.Order{ID: primitive.NewObjectID()}
	repo := &mockRepo{
		createFn: func(_ context.Context, _ *domain.Order) mo.Result[*domain.Order] {
			return mo.Ok(want)
		},
	}
	svc := newSvc(repo)
	res := svc.PlaceOrder(context.Background(), domain.CreateInput{
		Items: []domain.LineItemInput{{SKU: "SKU1", Quantity: 2, UnitPrice: 9.99}},
	})
	if res.IsError() {
		t.Fatalf("unexpected error: %v", res.Error())
	}
	if res.MustGet() != want {
		t.Error("expected repo result to be returned")
	}
}

func TestService_PlaceOrder_validWithCustomerID(t *testing.T) {
	customerID := primitive.NewObjectID()
	hexID := customerID.Hex()
	var capturedOrder *domain.Order
	repo := &mockRepo{
		createFn: func(_ context.Context, o *domain.Order) mo.Result[*domain.Order] {
			capturedOrder = o
			return mo.Ok(o)
		},
	}
	svc := newSvc(repo)
	svc.PlaceOrder(context.Background(), domain.CreateInput{
		CustomerID: &hexID,
		Items:      []domain.LineItemInput{{SKU: "SKU1", Quantity: 1, UnitPrice: 5.0}},
	})
	if capturedOrder == nil {
		t.Fatal("repo.Create was not called")
	}
	if capturedOrder.CustomerID == nil || *capturedOrder.CustomerID != customerID {
		t.Errorf("expected CustomerID %v, got %v", customerID, capturedOrder.CustomerID)
	}
}

// --- ViewOrderDetail ---

func TestService_ViewOrderDetail_invalidID(t *testing.T) {
	svc := newSvc(&mockRepo{})
	res := svc.ViewOrderDetail(context.Background(), "bad")
	if !res.IsError() || !errors.Is(res.Error(), ErrInvalidObjectID) {
		t.Errorf("expected ErrInvalidObjectID, got %v", res.Error())
	}
}

func TestService_ViewOrderDetail_notFound(t *testing.T) {
	id := primitive.NewObjectID()
	repo := &mockRepo{
		getByIDFn: func(_ context.Context, _ primitive.ObjectID) mo.Result[mo.Option[domain.Order]] {
			return mo.Ok(mo.None[domain.Order]())
		},
	}
	svc := newSvc(repo)
	res := svc.ViewOrderDetail(context.Background(), id.Hex())
	if res.IsError() {
		t.Fatalf("unexpected error: %v", res.Error())
	}
	if !res.MustGet().IsAbsent() {
		t.Error("expected absent result")
	}
}

func TestService_ViewOrderDetail_found(t *testing.T) {
	id := primitive.NewObjectID()
	order := domain.Order{
		ID:    id,
		Items: []domain.LineItem{{SKU: "SKU1", Quantity: 1, UnitPrice: 5.0}},
	}
	repo := &mockRepo{
		getByIDFn: func(_ context.Context, _ primitive.ObjectID) mo.Result[mo.Option[domain.Order]] {
			return mo.Ok(mo.Some(order))
		},
	}
	products := &mockProductCatalog{
		resolveFn: func(_ context.Context, _ string) string { return "Widget" },
	}
	svc := NewService(repo, products, &mockCustomerCatalog{}, noopPublisher{})
	res := svc.ViewOrderDetail(context.Background(), id.Hex())
	if res.IsError() {
		t.Fatalf("unexpected error: %v", res.Error())
	}
	view, ok := res.MustGet().Get()
	if !ok {
		t.Fatal("expected present result")
	}
	if view.Items[0].ProductName != "Widget" {
		t.Errorf("expected ProductName 'Widget', got %q", view.Items[0].ProductName)
	}
}

// --- Deactivate ---

func TestService_Deactivate_invalidID(t *testing.T) {
	svc := newSvc(&mockRepo{})
	res := svc.Deactivate(context.Background(), "bad")
	if !res.IsError() || !errors.Is(res.Error(), ErrInvalidObjectID) {
		t.Errorf("expected ErrInvalidObjectID, got %v", res.Error())
	}
}

func TestService_Deactivate_validID(t *testing.T) {
	id := primitive.NewObjectID()
	called := false
	repo := &mockRepo{
		deactivateFn: func(_ context.Context, got primitive.ObjectID, _ time.Time) mo.Result[*domain.Order] {
			called = true
			if got != id {
				panic("wrong id")
			}
			return mo.Ok(&domain.Order{})
		},
	}
	svc := newSvc(repo)
	svc.Deactivate(context.Background(), id.Hex())
	if !called {
		t.Error("expected repo.Deactivate to be called")
	}
}

// --- CompletePayment ---

func TestService_CompletePayment_invalidID(t *testing.T) {
	svc := newSvc(&mockRepo{})
	res := svc.CompletePayment(context.Background(), "bad")
	if !res.IsError() || !errors.Is(res.Error(), ErrInvalidObjectID) {
		t.Errorf("expected ErrInvalidObjectID, got %v", res.Error())
	}
}

func TestService_CompletePayment_notFound(t *testing.T) {
	id := primitive.NewObjectID()
	repo := &mockRepo{
		getByIDFn: func(_ context.Context, _ primitive.ObjectID) mo.Result[mo.Option[domain.Order]] {
			return mo.Ok(mo.None[domain.Order]())
		},
	}
	svc := newSvc(repo)
	res := svc.CompletePayment(context.Background(), id.Hex())
	if !res.IsError() || !errors.Is(res.Error(), port.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", res.Error())
	}
}

func TestService_CompletePayment_alreadyCompleted(t *testing.T) {
	id := primitive.NewObjectID()
	repo := &mockRepo{
		getByIDFn: func(_ context.Context, _ primitive.ObjectID) mo.Result[mo.Option[domain.Order]] {
			return mo.Ok(mo.Some(domain.Order{ID: id, Status: domain.StatusPaymentCompleted}))
		},
	}
	svc := newSvc(repo)
	res := svc.CompletePayment(context.Background(), id.Hex())
	if !res.IsError() || !errors.Is(res.Error(), ErrAlreadyCompleted) {
		t.Errorf("expected ErrAlreadyCompleted, got %v", res.Error())
	}
}

func TestService_CompletePayment_success(t *testing.T) {
	id := primitive.NewObjectID()
	orig := domain.Order{ID: id, Items: []domain.LineItem{{SKU: "SKU1", Quantity: 1, UnitPrice: 5.0}}}
	var capturedOrder *domain.Order
	repo := &mockRepo{
		getByIDFn: func(_ context.Context, _ primitive.ObjectID) mo.Result[mo.Option[domain.Order]] {
			return mo.Ok(mo.Some(orig))
		},
		createFn: func(_ context.Context, o *domain.Order) mo.Result[*domain.Order] {
			capturedOrder = o
			return mo.Ok(o)
		},
	}
	svc := newSvc(repo)
	res := svc.CompletePayment(context.Background(), id.Hex())
	if res.IsError() {
		t.Fatalf("unexpected error: %v", res.Error())
	}
	if capturedOrder == nil {
		t.Fatal("expected repo.Create to be called")
	}
	if capturedOrder.Status != domain.StatusPaymentCompleted {
		t.Errorf("expected status %q, got %q", domain.StatusPaymentCompleted, capturedOrder.Status)
	}
	if capturedOrder.OriginalOrderID == nil || *capturedOrder.OriginalOrderID != id {
		t.Errorf("expected OriginalOrderID %v, got %v", id, capturedOrder.OriginalOrderID)
	}
}
