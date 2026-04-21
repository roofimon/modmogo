package customer

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

// createErrorToHTTP maps domain validation errors to fiber HTTP errors.
func createErrorToHTTP(err error) error {
	if errors.Is(err, ErrInvalidName) || errors.Is(err, ErrInvalidEmail) {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return fiber.NewError(fiber.StatusInternalServerError, err.Error())
}

// idErrorToHTTP maps domain ID/lookup errors to fiber HTTP errors.
func idErrorToHTTP(err error) error {
	if errors.Is(err, ErrInvalidObjectID) {
		return fiber.NewError(fiber.StatusBadRequest, "invalid customer id")
	}
	if errors.Is(err, ErrNotFound) {
		return fiber.NewError(fiber.StatusNotFound, "customer not found")
	}
	return fiber.NewError(fiber.StatusInternalServerError, err.Error())
}

// --- Orchestration ---

// RegisterRoutes mounts customer HTTP routes on app.
func RegisterRoutes(app *fiber.App, svc *Service) {
	g := app.Group("/customers")
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
			return createErrorToHTTP(res.Error())
		}
		return c.Status(fiber.StatusCreated).JSON(res.MustGet())
	}
}

func handleGetByID(svc *Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		res := svc.GetByID(c.UserContext(), c.Params("id"))
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

func handleDeactivate(svc *Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		res := svc.Deactivate(c.UserContext(), c.Params("id"))
		if res.IsError() {
			return idErrorToHTTP(res.Error())
		}
		return c.JSON(res.MustGet())
	}
}

func handleList(svc *Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		limit, err := parseLimit(c.Query("limit"), defaultListLimit, maxListLimit)
		if err != nil {
			return err
		}
		res := svc.List(c.UserContext(), limit)
		if res.IsError() {
			return fiber.NewError(fiber.StatusInternalServerError, res.Error().Error())
		}
		items := res.OrElse([]Customer{})
		return c.JSON(items)
	}
}

func handleListInactive(svc *Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		limit, err := parseLimit(c.Query("limit"), defaultListLimit, maxListLimit)
		if err != nil {
			return err
		}
		res := svc.ListInactive(c.UserContext(), limit)
		if res.IsError() {
			return fiber.NewError(fiber.StatusInternalServerError, res.Error().Error())
		}
		items := res.OrElse([]Customer{})
		return c.JSON(items)
	}
}


// parseObjectID converts a 24-char hex string to a MongoDB ObjectID.
func parseObjectID(s string) (primitive.ObjectID, error) {
	if len(s) != 24 {
		return primitive.NilObjectID, ErrInvalidObjectID
	}
	id, err := primitive.ObjectIDFromHex(s)
	if err != nil {
		return primitive.NilObjectID, ErrInvalidObjectID
	}
	return id, nil
}