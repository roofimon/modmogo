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

	"modmono/internal/customer/application"
	"modmono/internal/customer/domain"
	"modmono/internal/customer/port"
)

// --- mock UseCase ---

type mockUseCase struct {
	createFn            func(context.Context, domain.CreateInput) mo.Result[*domain.Customer]
	viewCustomerDetailFn func(context.Context, string) mo.Result[mo.Option[domain.Customer]]
	listFn              func(context.Context, int64) mo.Result[[]domain.Customer]
	listInactiveFn      func(context.Context, int64) mo.Result[[]domain.Customer]
	deactivateFn        func(context.Context, string) mo.Result[*domain.Customer]
}

func (m *mockUseCase) Create(ctx context.Context, in domain.CreateInput) mo.Result[*domain.Customer] {
	if m.createFn == nil {
		panic("unexpected call to Create")
	}
	return m.createFn(ctx, in)
}

func (m *mockUseCase) ViewCustomerDetail(ctx context.Context, id string) mo.Result[mo.Option[domain.Customer]] {
	if m.viewCustomerDetailFn == nil {
		panic("unexpected call to ViewCustomerDetail")
	}
	return m.viewCustomerDetailFn(ctx, id)
}

func (m *mockUseCase) List(ctx context.Context, limit int64) mo.Result[[]domain.Customer] {
	if m.listFn == nil {
		panic("unexpected call to List")
	}
	return m.listFn(ctx, limit)
}

func (m *mockUseCase) ListInactive(ctx context.Context, limit int64) mo.Result[[]domain.Customer] {
	if m.listInactiveFn == nil {
		panic("unexpected call to ListInactive")
	}
	return m.listInactiveFn(ctx, limit)
}

func (m *mockUseCase) Deactivate(ctx context.Context, id string) mo.Result[*domain.Customer] {
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

// --- GET /customers ---

func TestAPI_List_returnsCustomers(t *testing.T) {
	customers := []domain.Customer{
		{ID: primitive.NewObjectID(), Name: "Alice", Email: "alice@example.com", Status: domain.StatusActive},
	}
	app := newApp(&mockUseCase{
		listFn: func(_ context.Context, _ int64) mo.Result[[]domain.Customer] {
			return mo.Ok(customers)
		},
	})

	req := httptest.NewRequest("GET", "/customers", nil)
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
		listFn: func(_ context.Context, _ int64) mo.Result[[]domain.Customer] {
			return mo.Ok([]domain.Customer{})
		},
	})

	req := httptest.NewRequest("GET", "/customers", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestAPI_List_invalidLimit(t *testing.T) {
	app := newApp(&mockUseCase{})

	req := httptest.NewRequest("GET", "/customers?limit=abc", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

// --- GET /customers/inactive ---

func TestAPI_ListInactive_returnsCustomers(t *testing.T) {
	at := time.Now()
	customers := []domain.Customer{
		{ID: primitive.NewObjectID(), Name: "Bob", Email: "bob@example.com", Status: domain.StatusDeactivated, DeactivatedAt: &at},
	}
	app := newApp(&mockUseCase{
		listInactiveFn: func(_ context.Context, _ int64) mo.Result[[]domain.Customer] {
			return mo.Ok(customers)
		},
	})

	req := httptest.NewRequest("GET", "/customers/inactive", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

// --- GET /customers/:id ---

func TestAPI_ViewCustomerDetail_found(t *testing.T) {
	id := primitive.NewObjectID()
	customer := domain.Customer{ID: id, Name: "Alice", Email: "alice@example.com"}
	app := newApp(&mockUseCase{
		viewCustomerDetailFn: func(_ context.Context, _ string) mo.Result[mo.Option[domain.Customer]] {
			return mo.Ok(mo.Some(customer))
		},
	})

	req := httptest.NewRequest("GET", "/customers/"+id.Hex(), nil)
	resp, _ := app.Test(req)
	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestAPI_ViewCustomerDetail_notFound(t *testing.T) {
	id := primitive.NewObjectID()
	app := newApp(&mockUseCase{
		viewCustomerDetailFn: func(_ context.Context, _ string) mo.Result[mo.Option[domain.Customer]] {
			return mo.Ok(mo.None[domain.Customer]())
		},
	})

	req := httptest.NewRequest("GET", "/customers/"+id.Hex(), nil)
	resp, _ := app.Test(req)
	if resp.StatusCode != fiber.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

func TestAPI_ViewCustomerDetail_invalidID(t *testing.T) {
	app := newApp(&mockUseCase{
		viewCustomerDetailFn: func(_ context.Context, _ string) mo.Result[mo.Option[domain.Customer]] {
			return mo.Err[mo.Option[domain.Customer]](application.ErrInvalidObjectID)
		},
	})

	req := httptest.NewRequest("GET", "/customers/not-an-id", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

// --- POST /customers ---

func TestAPI_Create_success(t *testing.T) {
	customer := &domain.Customer{ID: primitive.NewObjectID(), Name: "Alice", Email: "alice@example.com", Status: domain.StatusActive}
	app := newApp(&mockUseCase{
		createFn: func(_ context.Context, _ domain.CreateInput) mo.Result[*domain.Customer] {
			return mo.Ok(customer)
		},
	})

	body := mustJSON(t, domain.CreateInput{Name: "Alice", Email: "alice@example.com"})
	req := httptest.NewRequest("POST", "/customers", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	if resp.StatusCode != fiber.StatusCreated {
		t.Errorf("expected 201, got %d", resp.StatusCode)
	}
}

func TestAPI_Create_invalidName(t *testing.T) {
	app := newApp(&mockUseCase{
		createFn: func(_ context.Context, _ domain.CreateInput) mo.Result[*domain.Customer] {
			return mo.Err[*domain.Customer](application.ErrInvalidName)
		},
	})

	body := mustJSON(t, domain.CreateInput{Name: "", Email: "alice@example.com"})
	req := httptest.NewRequest("POST", "/customers", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func TestAPI_Create_invalidEmail(t *testing.T) {
	app := newApp(&mockUseCase{
		createFn: func(_ context.Context, _ domain.CreateInput) mo.Result[*domain.Customer] {
			return mo.Err[*domain.Customer](application.ErrInvalidEmail)
		},
	})

	body := mustJSON(t, domain.CreateInput{Name: "Alice", Email: ""})
	req := httptest.NewRequest("POST", "/customers", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func TestAPI_Create_invalidJSON(t *testing.T) {
	app := newApp(&mockUseCase{})

	req := httptest.NewRequest("POST", "/customers", bytes.NewReader([]byte("not-json")))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

// --- POST /customers/:id/deactivate ---

func TestAPI_Deactivate_success(t *testing.T) {
	id := primitive.NewObjectID()
	customer := &domain.Customer{ID: primitive.NewObjectID(), Status: domain.StatusDeactivated}
	app := newApp(&mockUseCase{
		deactivateFn: func(_ context.Context, _ string) mo.Result[*domain.Customer] {
			return mo.Ok(customer)
		},
	})

	req := httptest.NewRequest("POST", "/customers/"+id.Hex()+"/deactivate", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestAPI_Deactivate_notFound(t *testing.T) {
	id := primitive.NewObjectID()
	app := newApp(&mockUseCase{
		deactivateFn: func(_ context.Context, _ string) mo.Result[*domain.Customer] {
			return mo.Err[*domain.Customer](port.ErrNotFound)
		},
	})

	req := httptest.NewRequest("POST", "/customers/"+id.Hex()+"/deactivate", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode != fiber.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}
