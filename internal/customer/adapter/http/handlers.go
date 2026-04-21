package httpadapter

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"

	"modmono/internal/customer/application"
	"modmono/internal/customer/domain"
	"modmono/internal/customer/port"
)

const defaultListLimit int64 = 50
const maxListLimit int64 = 100

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

// createErrorToHTTP maps domain validation errors to fiber HTTP errors.
func createErrorToHTTP(err error) error {
	if errors.Is(err, application.ErrInvalidName) || errors.Is(err, application.ErrInvalidEmail) {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return fiber.NewError(fiber.StatusInternalServerError, err.Error())
}

// idErrorToHTTP maps domain ID/lookup errors to fiber HTTP errors.
func idErrorToHTTP(err error) error {
	if errors.Is(err, application.ErrInvalidObjectID) {
		return fiber.NewError(fiber.StatusBadRequest, "invalid customer id")
	}
	if errors.Is(err, port.ErrNotFound) {
		return fiber.NewError(fiber.StatusNotFound, "customer not found")
	}
	return fiber.NewError(fiber.StatusInternalServerError, err.Error())
}

// RegisterRoutes mounts customer HTTP routes on app.
func RegisterRoutes(app *fiber.App, svc port.UseCase) {
	g := app.Group("/customers")
	g.Post("/", handleCreate(svc))
	g.Get("/", handleList(svc))
	g.Get("/inactive", handleListInactive(svc))
	g.Post("/:id/deactivate", handleDeactivate(svc))
	g.Get("/:id", handleViewCustomerDetail(svc))
}

func handleCreate(svc port.UseCase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var body domain.CreateInput
		if err := c.BodyParser(&body); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid JSON body")
		}
		res := svc.Create(c.UserContext(), body)
		if res.IsError() {
			return createErrorToHTTP(res.Error())
		}
		return c.Status(fiber.StatusCreated).JSON(res.MustGet())
	}
}

func handleViewCustomerDetail(svc port.UseCase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		res := svc.ViewCustomerDetail(c.UserContext(), c.Params("id"))
		if res.IsError() {
			return idErrorToHTTP(res.Error())
		}
		if res.MustGet().IsAbsent() {
			return fiber.NewError(fiber.StatusNotFound, "customer not found")
		}
		cust, _ := res.MustGet().Get()
		return c.JSON(cust)
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
		return c.JSON(res.OrElse([]domain.Customer{}))
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
		return c.JSON(res.OrElse([]domain.Customer{}))
	}
}
