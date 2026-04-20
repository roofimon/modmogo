package product

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

const defaultListLimit int64 = 50
const maxListLimit int64 = 100

// RegisterRoutes mounts product HTTP routes on app.
func RegisterRoutes(app *fiber.App, svc *Service) {
	g := app.Group("/products")
	g.Post("/", handleCreate(svc))
	g.Get("/", handleList(svc))
	g.Get("/inactive", handleListInactive(svc))
	g.Post("/:id/deactivate", handleDeactivate(svc))
	g.Get("/:id", handleGetByID(svc))
}

func handleCreate(svc *Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var body CreateInput
		if err := c.BodyParser(&body); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid JSON body")
		}
		res := svc.Create(c.UserContext(), body)
		if res.IsError() {
			if errors.Is(res.Error(), ErrInvalidName) || errors.Is(res.Error(), ErrInvalidPrice) {
				return fiber.NewError(fiber.StatusBadRequest, res.Error().Error())
			}
			return fiber.NewError(fiber.StatusInternalServerError, res.Error().Error())
		}
		return c.Status(fiber.StatusCreated).JSON(res.MustGet())
	}
}

func handleGetByID(svc *Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		opt, err := svc.GetByID(c.UserContext(), id)
		if err != nil {
			if errors.Is(err, ErrInvalidObjectID) {
				return fiber.NewError(fiber.StatusBadRequest, "invalid product id")
			}
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		if opt.IsAbsent() {
			return fiber.NewError(fiber.StatusNotFound, "product not found")
		}
		p, _ := opt.Get()
		return c.JSON(p)
	}
}

func handleDeactivate(svc *Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		res := svc.Deactivate(c.UserContext(), id)
		if res.IsError() {
			if errors.Is(res.Error(), ErrInvalidObjectID) {
				return fiber.NewError(fiber.StatusBadRequest, "invalid product id")
			}
			if errors.Is(res.Error(), ErrNotFound) {
				return fiber.NewError(fiber.StatusNotFound, "product not found")
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
			items = []Product{}
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
			items = []Product{}
		}
		return c.JSON(items)
	}
}
