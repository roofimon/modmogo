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
	g.Get("/:id", handleGetByID(svc, catalog, customers))
}

func handleListProducts(catalog ProductCatalog) fiber.Handler {
	return func(c *fiber.Ctx) error {
		limit, err := parseListLimit(c)
		if err != nil {
			return err
		}
		products, err := catalog.ListActiveProducts(c.UserContext(), limit)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		if products == nil {
			products = []CatalogProduct{}
		}
		return c.JSON(products)
	}
}

func handleListCustomers(customers CustomerCatalog) fiber.Handler {
	return func(c *fiber.Ctx) error {
		limit, err := parseListLimit(c)
		if err != nil {
			return err
		}
		result, err := customers.ListActiveCustomers(c.UserContext(), limit)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		if result == nil {
			result = []CatalogCustomer{}
		}
		return c.JSON(result)
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

func handleGetByID(svc *Service, products ProductCatalog, customers CustomerCatalog) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		opt, err := svc.GetByID(c.UserContext(), id)
		if err != nil {
			if errors.Is(err, ErrInvalidObjectID) {
				return fiber.NewError(fiber.StatusBadRequest, "invalid order id")
			}
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		if opt.IsAbsent() {
			return fiber.NewError(fiber.StatusNotFound, "order not found")
		}
		o, _ := opt.Get()

		customerName := ""
		if o.CustomerID != nil {
			customerName = customers.ResolveCustomerName(c.UserContext(), o.CustomerID.Hex())
		}

		productNames := make(map[string]string)
		for _, item := range o.Items {
			if _, seen := productNames[item.SKU]; !seen {
				productNames[item.SKU] = products.ResolveProductName(c.UserContext(), item.SKU)
			}
		}

		return c.JSON(toOrderView(o, customerName, productNames))
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
