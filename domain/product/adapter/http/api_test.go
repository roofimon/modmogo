package httpadapter

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/samber/mo"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"modmono/domain/product/application"
	"modmono/domain/product/domain"
	"modmono/domain/product/port"
)

// --- mock UseCase ---

type mockUseCase struct {
	createFn           func(context.Context, domain.CreateInput) mo.Result[*domain.Product]
	viewProductDetailFn func(context.Context, string) mo.Result[mo.Option[domain.Product]]
	findProductBySKUFn func(context.Context, string) mo.Result[mo.Option[domain.Product]]
	listFn             func(context.Context, int64) mo.Result[[]domain.Product]
	listInactiveFn     func(context.Context, int64) mo.Result[[]domain.Product]
	activateFn         func(context.Context, string) mo.Result[*domain.Product]
	deactivateFn       func(context.Context, string) mo.Result[*domain.Product]
}

func (m *mockUseCase) Create(ctx context.Context, in domain.CreateInput) mo.Result[*domain.Product] {
	if m.createFn == nil {
		panic("unexpected call to Create")
	}
	return m.createFn(ctx, in)
}

func (m *mockUseCase) ViewProductDetail(ctx context.Context, id string) mo.Result[mo.Option[domain.Product]] {
	if m.viewProductDetailFn == nil {
		panic("unexpected call to ViewProductDetail")
	}
	return m.viewProductDetailFn(ctx, id)
}

func (m *mockUseCase) FindProductBySKU(ctx context.Context, sku string) mo.Result[mo.Option[domain.Product]] {
	if m.findProductBySKUFn == nil {
		panic("unexpected call to FindProductBySKU")
	}
	return m.findProductBySKUFn(ctx, sku)
}

func (m *mockUseCase) List(ctx context.Context, limit int64) mo.Result[[]domain.Product] {
	if m.listFn == nil {
		panic("unexpected call to List")
	}
	return m.listFn(ctx, limit)
}

func (m *mockUseCase) ListInactive(ctx context.Context, limit int64) mo.Result[[]domain.Product] {
	if m.listInactiveFn == nil {
		panic("unexpected call to ListInactive")
	}
	return m.listInactiveFn(ctx, limit)
}

func (m *mockUseCase) Activate(ctx context.Context, id string) mo.Result[*domain.Product] {
	if m.activateFn == nil {
		panic("unexpected call to Activate")
	}
	return m.activateFn(ctx, id)
}

func (m *mockUseCase) Deactivate(ctx context.Context, id string) mo.Result[*domain.Product] {
	if m.deactivateFn == nil {
		panic("unexpected call to Deactivate")
	}
	return m.deactivateFn(ctx, id)
}

var _ port.UseCase = (*mockUseCase)(nil)

// --- helpers ---

func newApp(svc port.UseCase) *fiber.App {
	app := fiber.New()
	RegisterRoutes(app, svc)
	return app
}

func mustJSON(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	return b
}

func decodeJSON[T any](t *testing.T, body []byte) T {
	t.Helper()
	var out T
	if err := json.Unmarshal(body, &out); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	return out
}

// --- GET /products ---

func TestAPI_List_returnsProducts(t *testing.T) {
	products := []domain.Product{
		{ID: primitive.NewObjectID(), SKU: "SKU1", Name: "Widget", Price: 9.99, Status: domain.StatusActive},
	}
	app := newApp(&mockUseCase{
		listFn: func(_ context.Context, _ int64) mo.Result[[]domain.Product] {
			return mo.Ok(products)
		},
	})

	req := httptest.NewRequest("GET", "/products", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestAPI_List_emptyReturns200(t *testing.T) {
	app := newApp(&mockUseCase{
		listFn: func(_ context.Context, _ int64) mo.Result[[]domain.Product] {
			return mo.Ok([]domain.Product{})
		},
	})

	req := httptest.NewRequest("GET", "/products", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestAPI_List_invalidLimit(t *testing.T) {
	app := newApp(&mockUseCase{})

	req := httptest.NewRequest("GET", "/products?limit=abc", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

// --- GET /products/inactive ---

func TestAPI_ListInactive_returnsProducts(t *testing.T) {
	at := time.Now()
	products := []domain.Product{
		{ID: primitive.NewObjectID(), SKU: "SKU1", Name: "Widget", Status: domain.StatusDeactivated, DeactivatedAt: &at},
	}
	app := newApp(&mockUseCase{
		listInactiveFn: func(_ context.Context, _ int64) mo.Result[[]domain.Product] {
			return mo.Ok(products)
		},
	})

	req := httptest.NewRequest("GET", "/products/inactive", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

// --- GET /products/:id ---

func TestAPI_ViewProductDetail_found(t *testing.T) {
	id := primitive.NewObjectID()
	product := domain.Product{ID: id, SKU: "SKU1", Name: "Widget", Price: 9.99}
	app := newApp(&mockUseCase{
		viewProductDetailFn: func(_ context.Context, _ string) mo.Result[mo.Option[domain.Product]] {
			return mo.Ok(mo.Some(product))
		},
	})

	req := httptest.NewRequest("GET", "/products/"+id.Hex(), nil)
	resp, _ := app.Test(req)
	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestAPI_ViewProductDetail_notFound(t *testing.T) {
	id := primitive.NewObjectID()
	app := newApp(&mockUseCase{
		viewProductDetailFn: func(_ context.Context, _ string) mo.Result[mo.Option[domain.Product]] {
			return mo.Ok(mo.None[domain.Product]())
		},
	})

	req := httptest.NewRequest("GET", "/products/"+id.Hex(), nil)
	resp, _ := app.Test(req)
	if resp.StatusCode != fiber.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

func TestAPI_ViewProductDetail_invalidID(t *testing.T) {
	app := newApp(&mockUseCase{
		viewProductDetailFn: func(_ context.Context, _ string) mo.Result[mo.Option[domain.Product]] {
			return mo.Err[mo.Option[domain.Product]](application.ErrInvalidObjectID)
		},
	})

	req := httptest.NewRequest("GET", "/products/not-an-id", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

// --- POST /products ---

func TestAPI_Create_success(t *testing.T) {
	product := &domain.Product{ID: primitive.NewObjectID(), SKU: "SKU1", Name: "Widget", Price: 9.99, Status: domain.StatusActive}
	app := newApp(&mockUseCase{
		createFn: func(_ context.Context, _ domain.CreateInput) mo.Result[*domain.Product] {
			return mo.Ok(product)
		},
	})

	body := mustJSON(t, domain.CreateInput{SKU: "SKU1", Name: "Widget", Price: 9.99})
	req := httptest.NewRequest("POST", "/products", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	if resp.StatusCode != fiber.StatusCreated {
		t.Errorf("expected 201, got %d", resp.StatusCode)
	}
}

func TestAPI_Create_invalidName(t *testing.T) {
	app := newApp(&mockUseCase{
		createFn: func(_ context.Context, _ domain.CreateInput) mo.Result[*domain.Product] {
			return mo.Err[*domain.Product](application.ErrInvalidName)
		},
	})

	body := mustJSON(t, domain.CreateInput{SKU: "SKU1", Name: "", Price: 9.99})
	req := httptest.NewRequest("POST", "/products", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func TestAPI_Create_invalidJSON(t *testing.T) {
	app := newApp(&mockUseCase{})

	req := httptest.NewRequest("POST", "/products", bytes.NewReader([]byte("not-json")))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

// --- POST /products/:id/deactivate ---

func TestAPI_Deactivate_success(t *testing.T) {
	id := primitive.NewObjectID()
	product := &domain.Product{ID: primitive.NewObjectID(), Status: domain.StatusDeactivated}
	app := newApp(&mockUseCase{
		deactivateFn: func(_ context.Context, _ string) mo.Result[*domain.Product] {
			return mo.Ok(product)
		},
	})

	req := httptest.NewRequest("POST", "/products/"+id.Hex()+"/deactivate", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestAPI_Deactivate_notFound(t *testing.T) {
	id := primitive.NewObjectID()
	app := newApp(&mockUseCase{
		deactivateFn: func(_ context.Context, _ string) mo.Result[*domain.Product] {
			return mo.Err[*domain.Product](port.ErrNotFound)
		},
	})

	req := httptest.NewRequest("POST", "/products/"+id.Hex()+"/deactivate", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode != fiber.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

// --- POST /products/:id/activate ---

func TestAPI_Activate_success(t *testing.T) {
	id := primitive.NewObjectID()
	product := &domain.Product{ID: primitive.NewObjectID(), Status: domain.StatusActive}
	app := newApp(&mockUseCase{
		activateFn: func(_ context.Context, _ string) mo.Result[*domain.Product] {
			return mo.Ok(product)
		},
	})

	req := httptest.NewRequest("POST", "/products/"+id.Hex()+"/activate", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestAPI_Activate_notFound(t *testing.T) {
	id := primitive.NewObjectID()
	app := newApp(&mockUseCase{
		activateFn: func(_ context.Context, _ string) mo.Result[*domain.Product] {
			return mo.Err[*domain.Product](port.ErrNotFound)
		},
	})

	req := httptest.NewRequest("POST", "/products/"+id.Hex()+"/activate", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode != fiber.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}
