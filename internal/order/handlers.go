package order

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

const defaultListLimit int64 = 50
const maxListLimit int64 = 100

// RegisterRoutes mounts order HTTP routes on app.
func RegisterRoutes(app *fiber.App, svc *Service, catalog ProductCatalog, customers CustomerCatalog) {
	g := app.Group("/orders")
	g.Post("/", handleCreate(svc))
	g.Get("/products", handleListProducts(catalog))
	g.Get("/customers", handleListCustomers(customers))
	g.Get("/", handleList(svc))
	g.Get("/inactive", handleListInactive(svc))
	g.Get("/payment-completed", handleListPaymentCompleted(svc))
	g.Post("/:id/deactivate", handleDeactivate(svc))
	g.Post("/:id/complete-payment", handleCompletePayment(svc))
	g.Get("/:id", handleGetByID(svc))
}

func handleListProducts(catalog ProductCatalog) fiber.Handler {
	return func(c *fiber.Ctx) error {
		limit, err := parseListLimit(c)
		if err != nil {
			return err
		}
		res := catalog.ListActiveProducts(c.UserContext(), limit)
		if res.IsError() {
			return fiber.NewError(fiber.StatusInternalServerError, res.Error().Error())
		}
		return c.JSON(res.OrElse([]CatalogProduct{}))
	}
}

func handleListCustomers(customers CustomerCatalog) fiber.Handler {
	return func(c *fiber.Ctx) error {
		limit, err := parseListLimit(c)
		if err != nil {
			return err
		}
		res := customers.ListActiveCustomers(c.UserContext(), limit)
		if res.IsError() {
			return fiber.NewError(fiber.StatusInternalServerError, res.Error().Error())
		}
		return c.JSON(res.OrElse([]CatalogCustomer{}))
	}
}

func handleCreate(svc *Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var body CreateInput
		if err := c.BodyParser(&body); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid JSON body")
		}
		res := svc.Create(c.UserContext(), body)
		if res.IsError() {
			if errors.Is(res.Error(), ErrNoItems) ||
				errors.Is(res.Error(), ErrInvalidSKU) ||
				errors.Is(res.Error(), ErrInvalidQuantity) ||
				errors.Is(res.Error(), ErrInvalidUnitPrice) ||
				errors.Is(res.Error(), ErrInvalidCustomerID) {
				return fiber.NewError(fiber.StatusBadRequest, res.Error().Error())
			}
			return fiber.NewError(fiber.StatusInternalServerError, res.Error().Error())
		}
		return c.Status(fiber.StatusCreated).JSON(res.MustGet())
	}
}

func handleGetByID(svc *Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		res := svc.GetByID(c.UserContext(), c.Params("id"))
		if res.IsError() {
			if errors.Is(res.Error(), ErrInvalidObjectID) {
				return fiber.NewError(fiber.StatusBadRequest, "invalid order id")
			}
			return fiber.NewError(fiber.StatusInternalServerError, res.Error().Error())
		}
		if res.MustGet().IsAbsent() {
			return fiber.NewError(fiber.StatusNotFound, "order not found")
		}
		view, _ := res.MustGet().Get()
		return c.JSON(view)
	}
}

func handleCompletePayment(svc *Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		res := svc.CompletePayment(c.UserContext(), id)
		if res.IsError() {
			switch {
			case errors.Is(res.Error(), ErrInvalidObjectID):
				return fiber.NewError(fiber.StatusBadRequest, "invalid order id")
			case errors.Is(res.Error(), ErrNotFound):
				return fiber.NewError(fiber.StatusNotFound, "order not found")
			case errors.Is(res.Error(), ErrAlreadyCompleted):
				return fiber.NewError(fiber.StatusConflict, res.Error().Error())
			}
			return fiber.NewError(fiber.StatusInternalServerError, res.Error().Error())
		}
		return c.Status(fiber.StatusCreated).JSON(res.MustGet())
	}
}

func handleDeactivate(svc *Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		res := svc.Deactivate(c.UserContext(), id)
		if res.IsError() {
			if errors.Is(res.Error(), ErrInvalidObjectID) {
				return fiber.NewError(fiber.StatusBadRequest, "invalid order id")
			}
			if errors.Is(res.Error(), ErrNotFound) {
				return fiber.NewError(fiber.StatusNotFound, "order not found")
			}
			return fiber.NewError(fiber.StatusInternalServerError, res.Error().Error())
		}
		return c.JSON(res.MustGet())
	}
}

func parseListLimit(c *fiber.Ctx) (int64, error) {
	limit := defaultListLimit
	if raw := c.Query("limit"); raw != "" {
		n, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || n < 1 {
			return 0, fiber.NewError(fiber.StatusBadRequest, "invalid limit")
		}
		if n > maxListLimit {
			n = maxListLimit
		}
		limit = n
	}
	return limit, nil
}

func handleListPaymentCompleted(svc *Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		limit, err := parseListLimit(c)
		if err != nil {
			return err
		}
		res := svc.ListPaymentCompleted(c.UserContext(), limit)
		if res.IsError() {
			return fiber.NewError(fiber.StatusInternalServerError, res.Error().Error())
		}
		items := res.MustGet()
		if items == nil {
			items = []Order{}
		}
		return c.JSON(items)
	}
}

func handleList(svc *Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		limit, err := parseListLimit(c)
		if err != nil {
			return err
		}
		res := svc.List(c.UserContext(), limit)
		if res.IsError() {
			return fiber.NewError(fiber.StatusInternalServerError, res.Error().Error())
		}
		items := res.MustGet()
		if items == nil {
			items = []Order{}
		}
		return c.JSON(items)
	}
}

func handleListInactive(svc *Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		limit, err := parseListLimit(c)
		if err != nil {
			return err
		}
		res := svc.ListInactive(c.UserContext(), limit)
		if res.IsError() {
			return fiber.NewError(fiber.StatusInternalServerError, res.Error().Error())
		}
		items := res.MustGet()
		if items == nil {
			items = []Order{}
		}
		return c.JSON(items)
	}
}
