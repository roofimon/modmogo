package health

import (
	"context"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"

	platformmongo "modmono/internal/platform/mongo"
)

const readyTimeout = 2 * time.Second

// RegisterRoutes mounts liveness and readiness probes on app.
// Readiness verifies MongoDB with a short Ping after resolving the lazy client.
func RegisterRoutes(app *fiber.App, lazy *platformmongo.LazyClient) {
	app.Get("/health/live", handleLive)
	app.Get("/health/ready", handleReady(lazy))
}

func handleLive(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusOK)
}

func handleReady(lazy *platformmongo.LazyClient) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), readyTimeout)
		defer cancel()
		client, err := lazy.Get(ctx)
		if err != nil {
			log.Printf("health ready: mongo: %v", err)
			return c.Status(fiber.StatusServiceUnavailable).SendString("not ready")
		}
		if err := client.Ping(ctx, nil); err != nil {
			log.Printf("health ready: mongo ping: %v", err)
			return c.Status(fiber.StatusServiceUnavailable).SendString("not ready")
		}
		return c.SendStatus(fiber.StatusOK)
	}
}
