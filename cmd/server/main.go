package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"

	"modmono/internal/health"
	platformmongo "modmono/internal/platform/mongo"
	"modmono/internal/product"
)

type config struct {
	MongoURI           string
	MongoDB            string
	HTTPAddr           string
	CORSAllowedOrigins string // comma-separated; empty disables CORS middleware
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func loadConfig() config {
	return config{
		MongoURI:           getenv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDB:            getenv("MONGO_DB", "modmono"),
		HTTPAddr:           getenv("HTTP_ADDR", ":8080"),
		CORSAllowedOrigins: getenv("CORS_ALLOWED_ORIGINS", ""),
	}
}

// corsAllowOrigins returns a comma-separated list suitable for fiber cors, or "" if none configured.
func corsAllowOrigins(csv string) string {
	parts := strings.Split(csv, ",")
	var out []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return strings.Join(out, ",")
}

func newProductService(lazy *platformmongo.LazyClient, dbName string) *product.Service {
	repo := product.NewMongoRepository(lazy, dbName)
	return product.NewService(repo)
}

func newFiberApp(svc *product.Service, lazy *platformmongo.LazyClient, cfg config) *fiber.App {
	app := fiber.New(fiber.Config{AppName: "modmono"})
	app.Use(recover.New(), requestid.New(), logger.New())
	if allow := corsAllowOrigins(cfg.CORSAllowedOrigins); allow != "" {
		app.Use(cors.New(cors.Config{
			AllowOrigins: allow,
			AllowMethods: "GET,POST,OPTIONS,HEAD",
			AllowHeaders: "Origin,Content-Type,Accept,Authorization,X-Requested-With",
		}))
	}
	product.RegisterRoutes(app, svc)
	health.RegisterRoutes(app, lazy)
	return app
}

func runHTTPServer(ctx context.Context, app *fiber.App, addr string) error {
	go func() {
		if err := app.Listen(addr); err != nil {
			log.Printf("fiber listen: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		return fmt.Errorf("fiber shutdown: %w", err)
	}
	return nil
}

func run(ctx context.Context) error {
	cfg := loadConfig()

	lazy := platformmongo.NewLazyClient(cfg.MongoURI)
	defer lazy.Disconnect()

	svc := newProductService(lazy, cfg.MongoDB)
	app := newFiberApp(svc, lazy, cfg)

	return runHTTPServer(ctx, app, cfg.HTTPAddr)
}

func main() {
	if err := run(context.Background()); err != nil {
		log.Fatal(err)
	}
}
