package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"

	"modmono/internal/health"
	platformmongo "modmono/internal/platform/mongo"
	"modmono/internal/product"
)

type config struct {
	MongoURI string
	MongoDB  string
	HTTPAddr string
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func loadConfig() config {
	return config{
		MongoURI: getenv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDB:  getenv("MONGO_DB", "modmono"),
		HTTPAddr: getenv("HTTP_ADDR", ":8080"),
	}
}

func newProductService(lazy *platformmongo.LazyClient, dbName string) *product.Service {
	repo := product.NewMongoRepository(lazy, dbName)
	return product.NewService(repo)
}

func newFiberApp(svc *product.Service, lazy *platformmongo.LazyClient) *fiber.App {
	app := fiber.New(fiber.Config{AppName: "modmono"})
	app.Use(recover.New(), requestid.New(), logger.New())
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
	app := newFiberApp(svc, lazy)

	return runHTTPServer(ctx, app, cfg.HTTPAddr)
}

func main() {
	if err := run(context.Background()); err != nil {
		log.Fatal(err)
	}
}
