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

	customeradapter "modmono/domain/customer/adapter"
	customerhttp "modmono/domain/customer/adapter/http"
	customerapplication "modmono/domain/customer/application"
	"modmono/domain/health"
	orderadapter "modmono/domain/order/adapter"
	ordercatalog "modmono/domain/order/adapter/catalog"
	orderhttp "modmono/domain/order/adapter/http"
	orderapplication "modmono/domain/order/application"
	orderport "modmono/domain/order/port"
	platformevent "modmono/domain/platform/event"
	platformmongo "modmono/domain/platform/mongo"
	productadapter "modmono/domain/product/adapter"
	producthttp "modmono/domain/product/adapter/http"
	productapplication "modmono/domain/product/application"
)

type config struct {
	MongoURI           string
	MongoDB            string
	HTTPAddr           string
	CORSAllowedOrigins string
}

func run(ctx context.Context) error {
	cfg := loadConfig()
	lazy := platformmongo.NewLazyClient(cfg.MongoURI)
	defer lazy.Disconnect()

	pub             := platformevent.LogPublisher{}
	productSvc      := newProductService(lazy, cfg.MongoDB, pub)
	customerSvc     := newCustomerService(lazy, cfg.MongoDB, pub)
	productCatalog  := ordercatalog.NewProductCatalogAdapter(productSvc)
	customerCatalog := ordercatalog.NewCustomerCatalogAdapter(customerSvc)
	orderSvc        := newOrderService(lazy, cfg.MongoDB, productCatalog, customerCatalog, pub)

	app := newFiberApp(productSvc, customerSvc, orderSvc, productCatalog, customerCatalog, lazy, cfg)
	return runHTTPServer(ctx, app, cfg.HTTPAddr)
}

func main() {
	if err := run(context.Background()); err != nil {
		log.Fatal(err)
	}
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

func newProductService(lazy *platformmongo.LazyClient, dbName string, pub platformevent.Publisher) *productapplication.Service {
	repo := productadapter.NewMongoRepository(lazy, dbName)
	return productapplication.NewService(repo, pub)
}

func newCustomerService(lazy *platformmongo.LazyClient, dbName string, pub platformevent.Publisher) *customerapplication.Service {
	repo := customeradapter.NewMongoRepository(lazy, dbName)
	return customerapplication.NewService(repo, pub)
}

func newOrderService(lazy *platformmongo.LazyClient, dbName string, products orderport.ProductCatalog, customers orderport.CustomerCatalog, pub platformevent.Publisher) *orderapplication.Service {
	repo := orderadapter.NewMongoRepository(lazy, dbName)
	return orderapplication.NewService(repo, products, customers, pub)
}

func newFiberApp(
	productSvc      *productapplication.Service,
	customerSvc     *customerapplication.Service,
	orderSvc        *orderapplication.Service,
	productCatalog  orderport.ProductCatalog,
	customerCatalog orderport.CustomerCatalog,
	lazy            *platformmongo.LazyClient,
	cfg             config,
) *fiber.App {
	app := fiber.New(fiber.Config{AppName: "modmono"})
	app.Use(recover.New(), requestid.New(), logger.New())
	if allow := corsAllowOrigins(cfg.CORSAllowedOrigins); allow != "" {
		app.Use(cors.New(cors.Config{
			AllowOrigins: allow,
			AllowMethods: "GET,POST,OPTIONS,HEAD",
			AllowHeaders: "Origin,Content-Type,Accept,Authorization,X-Requested-With",
		}))
	}
	producthttp.RegisterRoutes(app, productSvc)
	customerhttp.RegisterRoutes(app, customerSvc)
	orderhttp.RegisterRoutes(app, orderSvc, productCatalog, customerCatalog)
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
