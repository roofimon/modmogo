package httpadapter

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/samber/mo"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"modmono/internal/order/application"
	"modmono/internal/order/domain"
	"modmono/internal/order/port"
)

// --- mock UseCase ---

type mockUseCase struct {
	placeOrderFn          func(context.Context, domain.CreateInput) mo.Result[*domain.Order]
	viewOrderDetailFn     func(context.Context, string) mo.Result[mo.Option[domain.OrderView]]
	listFn                func(context.Context, int64) mo.Result[[]domain.Order]
	listInactiveFn        func(context.Context, int64) mo.Result[[]domain.Order]
	listPaymentCompletedFn func(context.Context, int64) mo.Result[[]domain.Order]
	deactivateFn          func(context.Context, string) mo.Result[*domain.Order]
	completePaymentFn     func(context.Context, string) mo.Result[*domain.Order]
}

func (m *mockUseCase) PlaceOrder(ctx context.Context, in domain.CreateInput) mo.Result[*domain.Order] {
	if m.placeOrderFn == nil {
		panic("unexpected call to PlaceOrder")
	}
	return m.placeOrderFn(ctx, in)
}

func (m *mockUseCase) ViewOrderDetail(ctx context.Context, id string) mo.Result[mo.Option[domain.OrderView]] {
	if m.viewOrderDetailFn == nil {
		panic("unexpected call to ViewOrderDetail")
	}
	return m.viewOrderDetailFn(ctx, id)
}

func (m *mockUseCase) List(ctx context.Context, limit int64) mo.Result[[]domain.Order] {
	if m.listFn == nil {
		panic("unexpected call to List")
	}
	return m.listFn(ctx, limit)
}

func (m *mockUseCase) ListInactive(ctx context.Context, limit int64) mo.Result[[]domain.Order] {
	if m.listInactiveFn == nil {
		panic("unexpected call to ListInactive")
	}
	return m.listInactiveFn(ctx, limit)
}

func (m *mockUseCase) ListPaymentCompleted(ctx context.Context, limit int64) mo.Result[[]domain.Order] {
	if m.listPaymentCompletedFn == nil {
		panic("unexpected call to ListPaymentCompleted")
	}
	return m.listPaymentCompletedFn(ctx, limit)
}

func (m *mockUseCase) Deactivate(ctx context.Context, id string) mo.Result[*domain.Order] {
	if m.deactivateFn == nil {
		panic("unexpected call to Deactivate")
	}
	return m.deactivateFn(ctx, id)
}

func (m *mockUseCase) CompletePayment(ctx context.Context, id string) mo.Result[*domain.Order] {
	if m.completePaymentFn == nil {
		panic("unexpected call to CompletePayment")
	}
	return m.completePaymentFn(ctx, id)
}

var _ port.UseCase = (*mockUseCase)(nil)

// --- mock catalogs ---

type mockProductCatalog struct {
	listFn func(context.Context, int64) mo.Result[[]port.CatalogProduct]
}

func (m *mockProductCatalog) ListActiveProducts(ctx context.Context, limit int64) mo.Result[[]port.CatalogProduct] {
	if m.listFn == nil {
		return mo.Ok([]port.CatalogProduct{})
	}
	return m.listFn(ctx, limit)
}

func (m *mockProductCatalog) ResolveProductName(_ context.Context, _ string) string { return "" }

type mockCustomerCatalog struct {
	listFn func(context.Context, int64) mo.Result[[]port.CatalogCustomer]
}

func (m *mockCustomerCatalog) ListActiveCustomers(ctx context.Context, limit int64) mo.Result[[]port.CatalogCustomer] {
	if m.listFn == nil {
		return mo.Ok([]port.CatalogCustomer{})
	}
	return m.listFn(ctx, limit)
}

func (m *mockCustomerCatalog) ResolveCustomerName(_ context.Context, _ string) string { return "" }

// --- helpers ---

func newApp(svc port.UseCase, products port.ProductCatalog, customers port.CustomerCatalog) *fiber.App {
	app := fiber.New()
	RegisterRoutes(app, svc, products, customers)
	return app
}

func defaultApp(svc port.UseCase) *fiber.App {
	return newApp(svc, &mockProductCatalog{}, &mockCustomerCatalog{})
}

func mustJSON(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	return b
}

// --- GET /orders ---

func TestAPI_List_returnsOrders(t *testing.T) {
	orders := []domain.Order{{ID: primitive.NewObjectID()}}
	app := defaultApp(&mockUseCase{
		listFn: func(_ context.Context, _ int64) mo.Result[[]domain.Order] {
			return mo.Ok(orders)
		},
	})
	req := httptest.NewRequest("GET", "/orders", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestAPI_List_invalidLimit(t *testing.T) {
	app := defaultApp(&mockUseCase{})
	req := httptest.NewRequest("GET", "/orders?limit=bad", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

// --- GET /orders/inactive ---

func TestAPI_ListInactive_returnsOrders(t *testing.T) {
	app := defaultApp(&mockUseCase{
		listInactiveFn: func(_ context.Context, _ int64) mo.Result[[]domain.Order] {
			return mo.Ok([]domain.Order{})
		},
	})
	req := httptest.NewRequest("GET", "/orders/inactive", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

// --- GET /orders/payment-completed ---

func TestAPI_ListPaymentCompleted_returnsOrders(t *testing.T) {
	app := defaultApp(&mockUseCase{
		listPaymentCompletedFn: func(_ context.Context, _ int64) mo.Result[[]domain.Order] {
			return mo.Ok([]domain.Order{})
		},
	})
	req := httptest.NewRequest("GET", "/orders/payment-completed", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

// --- GET /orders/products ---

func TestAPI_ListProducts_returnsProducts(t *testing.T) {
	products := &mockProductCatalog{
		listFn: func(_ context.Context, _ int64) mo.Result[[]port.CatalogProduct] {
			return mo.Ok([]port.CatalogProduct{{SKU: "SKU1", Name: "Widget", Price: 9.99}})
		},
	}
	app := newApp(&mockUseCase{}, products, &mockCustomerCatalog{})
	req := httptest.NewRequest("GET", "/orders/products", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

// --- GET /orders/customers ---

func TestAPI_ListCustomers_returnsCustomers(t *testing.T) {
	customers := &mockCustomerCatalog{
		listFn: func(_ context.Context, _ int64) mo.Result[[]port.CatalogCustomer] {
			return mo.Ok([]port.CatalogCustomer{{ID: primitive.NewObjectID().Hex(), Name: "Alice"}})
		},
	}
	app := newApp(&mockUseCase{}, &mockProductCatalog{}, customers)
	req := httptest.NewRequest("GET", "/orders/customers", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

// --- GET /orders/:id ---

func TestAPI_ViewOrderDetail_found(t *testing.T) {
	id := primitive.NewObjectID()
	view := domain.OrderView{ID: id}
	app := defaultApp(&mockUseCase{
		viewOrderDetailFn: func(_ context.Context, _ string) mo.Result[mo.Option[domain.OrderView]] {
			return mo.Ok(mo.Some(view))
		},
	})
	req := httptest.NewRequest("GET", "/orders/"+id.Hex(), nil)
	resp, _ := app.Test(req)
	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestAPI_ViewOrderDetail_notFound(t *testing.T) {
	id := primitive.NewObjectID()
	app := defaultApp(&mockUseCase{
		viewOrderDetailFn: func(_ context.Context, _ string) mo.Result[mo.Option[domain.OrderView]] {
			return mo.Ok(mo.None[domain.OrderView]())
		},
	})
	req := httptest.NewRequest("GET", "/orders/"+id.Hex(), nil)
	resp, _ := app.Test(req)
	if resp.StatusCode != fiber.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

func TestAPI_ViewOrderDetail_invalidID(t *testing.T) {
	app := defaultApp(&mockUseCase{
		viewOrderDetailFn: func(_ context.Context, _ string) mo.Result[mo.Option[domain.OrderView]] {
			return mo.Err[mo.Option[domain.OrderView]](application.ErrInvalidObjectID)
		},
	})
	req := httptest.NewRequest("GET", "/orders/not-an-id", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

// --- POST /orders ---

func TestAPI_PlaceOrder_success(t *testing.T) {
	order := &domain.Order{ID: primitive.NewObjectID()}
	app := defaultApp(&mockUseCase{
		placeOrderFn: func(_ context.Context, _ domain.CreateInput) mo.Result[*domain.Order] {
			return mo.Ok(order)
		},
	})
	body := mustJSON(t, domain.CreateInput{
		Items: []domain.LineItemInput{{SKU: "SKU1", Quantity: 1, UnitPrice: 9.99}},
	})
	req := httptest.NewRequest("POST", "/orders", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	if resp.StatusCode != fiber.StatusCreated {
		t.Errorf("expected 201, got %d", resp.StatusCode)
	}
}

func TestAPI_PlaceOrder_noItems(t *testing.T) {
	app := defaultApp(&mockUseCase{
		placeOrderFn: func(_ context.Context, _ domain.CreateInput) mo.Result[*domain.Order] {
			return mo.Err[*domain.Order](application.ErrNoItems)
		},
	})
	body := mustJSON(t, domain.CreateInput{Items: []domain.LineItemInput{}})
	req := httptest.NewRequest("POST", "/orders", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func TestAPI_PlaceOrder_invalidJSON(t *testing.T) {
	app := defaultApp(&mockUseCase{})
	req := httptest.NewRequest("POST", "/orders", bytes.NewReader([]byte("not-json")))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

// --- POST /orders/:id/deactivate ---

func TestAPI_Deactivate_success(t *testing.T) {
	id := primitive.NewObjectID()
	app := defaultApp(&mockUseCase{
		deactivateFn: func(_ context.Context, _ string) mo.Result[*domain.Order] {
			return mo.Ok(&domain.Order{Status: domain.StatusDeactivated})
		},
	})
	req := httptest.NewRequest("POST", "/orders/"+id.Hex()+"/deactivate", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestAPI_Deactivate_notFound(t *testing.T) {
	id := primitive.NewObjectID()
	app := defaultApp(&mockUseCase{
		deactivateFn: func(_ context.Context, _ string) mo.Result[*domain.Order] {
			return mo.Err[*domain.Order](port.ErrNotFound)
		},
	})
	req := httptest.NewRequest("POST", "/orders/"+id.Hex()+"/deactivate", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode != fiber.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

// --- POST /orders/:id/complete-payment ---

func TestAPI_CompletePayment_success(t *testing.T) {
	id := primitive.NewObjectID()
	app := defaultApp(&mockUseCase{
		completePaymentFn: func(_ context.Context, _ string) mo.Result[*domain.Order] {
			return mo.Ok(&domain.Order{Status: domain.StatusPaymentCompleted})
		},
	})
	req := httptest.NewRequest("POST", "/orders/"+id.Hex()+"/complete-payment", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode != fiber.StatusCreated {
		t.Errorf("expected 201, got %d", resp.StatusCode)
	}
}

func TestAPI_CompletePayment_alreadyCompleted(t *testing.T) {
	id := primitive.NewObjectID()
	app := defaultApp(&mockUseCase{
		completePaymentFn: func(_ context.Context, _ string) mo.Result[*domain.Order] {
			return mo.Err[*domain.Order](application.ErrAlreadyCompleted)
		},
	})
	req := httptest.NewRequest("POST", "/orders/"+id.Hex()+"/complete-payment", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode != fiber.StatusConflict {
		t.Errorf("expected 409, got %d", resp.StatusCode)
	}
}
