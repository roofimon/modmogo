package httpadapter

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"

	"modmono/domain/order/application"
	"modmono/domain/order/domain"
	"modmono/domain/order/port"
)

const defaultListLimit int64 = 50
const maxListLimit int64 = 100

// --- Pure Logic ---

// parseLimit parses a raw query string into a clamped list limit.
func parseLimit(raw string, defaultVal, maxVal int64) (int64, error) {
	if raw == "" {
		return defaultVal, nil
	}
	n, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || n < 1 {
		return 0, fiber.NewError(fiber.StatusBadRequest, "invalid limit")
	}
	if n > maxVal {
		n = maxVal
	}
	return n, nil
}

// createErrorToHTTP maps order validation errors to fiber HTTP errors.
func createErrorToHTTP(err error) error {
	switch {
	case errors.Is(err, application.ErrNoItems),
		errors.Is(err, application.ErrInvalidSKU),
		errors.Is(err, application.ErrInvalidQuantity),
		errors.Is(err, application.ErrInvalidUnitPrice),
		errors.Is(err, application.ErrInvalidCustomerID):
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return fiber.NewError(fiber.StatusInternalServerError, err.Error())
}

// idErrorToHTTP maps ID/lookup errors to fiber HTTP errors.
func idErrorToHTTP(err error) error {
	switch {
	case errors.Is(err, application.ErrInvalidObjectID):
		return fiber.NewError(fiber.StatusBadRequest, "invalid order id")
	case errors.Is(err, port.ErrNotFound):
		return fiber.NewError(fiber.StatusNotFound, "order not found")
	case errors.Is(err, application.ErrAlreadyCompleted):
		return fiber.NewError(fiber.StatusConflict, err.Error())
	}
	return fiber.NewError(fiber.StatusInternalServerError, err.Error())
}

// --- Orchestration ---

// RegisterRoutes mounts order HTTP routes on app.
func RegisterRoutes(app *fiber.App, svc port.UseCase, products port.ProductCatalog, customers port.CustomerCatalog) {
	g := app.Group("/orders")
	g.Post("/", handlePlaceOrder(svc))
	g.Get("/products", handleListProducts(products))
	g.Get("/customers", handleListCustomers(customers))
	g.Get("/", handleList(svc))
	g.Get("/inactive", handleListInactive(svc))
	g.Get("/payment-completed", handleListPaymentCompleted(svc))
	g.Post("/:id/deactivate", handleDeactivate(svc))
	g.Post("/:id/complete-payment", handleCompletePayment(svc))
	g.Get("/:id", handleViewOrderDetail(svc))
}

func handleListProducts(catalog port.ProductCatalog) fiber.Handler {
	return func(c *fiber.Ctx) error {
		limit, err := parseLimit(c.Query("limit"), defaultListLimit, maxListLimit)
		if err != nil {
			return err
		}
		res := catalog.ListActiveProducts(c.UserContext(), limit)
		if res.IsError() {
			return fiber.NewError(fiber.StatusInternalServerError, res.Error().Error())
		}
		return c.JSON(res.OrElse([]port.CatalogProduct{}))
	}
}

func handleListCustomers(customers port.CustomerCatalog) fiber.Handler {
	return func(c *fiber.Ctx) error {
		limit, err := parseLimit(c.Query("limit"), defaultListLimit, maxListLimit)
		if err != nil {
			return err
		}
		res := customers.ListActiveCustomers(c.UserContext(), limit)
		if res.IsError() {
			return fiber.NewError(fiber.StatusInternalServerError, res.Error().Error())
		}
		return c.JSON(res.OrElse([]port.CatalogCustomer{}))
	}
}

func handlePlaceOrder(svc port.UseCase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var body domain.CreateInput
		if err := c.BodyParser(&body); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid JSON body")
		}
		res := svc.PlaceOrder(c.UserContext(), body)
		if res.IsError() {
			return createErrorToHTTP(res.Error())
		}
		return c.Status(fiber.StatusCreated).JSON(res.MustGet())
	}
}

func handleViewOrderDetail(svc port.UseCase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		res := svc.ViewOrderDetail(c.UserContext(), c.Params("id"))
		if res.IsError() {
			return idErrorToHTTP(res.Error())
		}
		if res.MustGet().IsAbsent() {
			return fiber.NewError(fiber.StatusNotFound, "order not found")
		}
		view, _ := res.MustGet().Get()
		return c.JSON(view)
	}
}

func handleCompletePayment(svc port.UseCase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		res := svc.CompletePayment(c.UserContext(), c.Params("id"))
		if res.IsError() {
			return idErrorToHTTP(res.Error())
		}
		return c.Status(fiber.StatusCreated).JSON(res.MustGet())
	}
}

func handleDeactivate(svc port.UseCase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		res := svc.Deactivate(c.UserContext(), c.Params("id"))
		if res.IsError() {
			return idErrorToHTTP(res.Error())
		}
		return c.JSON(res.MustGet())
	}
}

func handleList(svc port.UseCase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		limit, err := parseLimit(c.Query("limit"), defaultListLimit, maxListLimit)
		if err != nil {
			return err
		}
		res := svc.List(c.UserContext(), limit)
		if res.IsError() {
			return fiber.NewError(fiber.StatusInternalServerError, res.Error().Error())
		}
		return c.JSON(res.OrElse([]domain.Order{}))
	}
}

func handleListInactive(svc port.UseCase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		limit, err := parseLimit(c.Query("limit"), defaultListLimit, maxListLimit)
		if err != nil {
			return err
		}
		res := svc.ListInactive(c.UserContext(), limit)
		if res.IsError() {
			return fiber.NewError(fiber.StatusInternalServerError, res.Error().Error())
		}
		return c.JSON(res.OrElse([]domain.Order{}))
	}
}

func handleListPaymentCompleted(svc port.UseCase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		limit, err := parseLimit(c.Query("limit"), defaultListLimit, maxListLimit)
		if err != nil {
			return err
		}
		res := svc.ListPaymentCompleted(c.UserContext(), limit)
		if res.IsError() {
			return fiber.NewError(fiber.StatusInternalServerError, res.Error().Error())
		}
		return c.JSON(res.OrElse([]domain.Order{}))
	}
}
