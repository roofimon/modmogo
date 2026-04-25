package main

import (
	"context"
	"log"
	"os"

	platformmongo "modmono/domain/platform/mongo"
	productadapter "modmono/domain/product/adapter"
	productdomain "modmono/domain/product/domain"
)

// seedRows are demo catalog items with vendor-style SKUs.
var seedRows = []struct {
	sku   string
	name  string
	price float64
}{
	{"HYDRO-SS-750-SLV", "Stainless Steel Insulated Water Bottle 750ml", 28.99},
	{"APP-TEE-CRW-ORG-NVY-M", "Organic Cotton Crew Neck T-Shirt, Navy, M", 32.50},
	{"CBL-USBC-USBC-2M-BLK", "USB-C to USB-C Charging Cable 2m", 14.95},
	{"BRW-PO-DIP-CER-001", "Ceramic Pour-Over Coffee Dripper", 22.00},
	{"AUD-TWS-NC-BT52-BLK", "Bluetooth Noise-Canceling Earbuds", 129.99},
	{"STN-NB-A5-DOT-HC-IV", "Hardcover Notebook A5, Dotted", 18.75},
	{"LGT-LED-DESK-DIM-WHT", "LED Desk Lamp with Touch Dimming", 45.00},
	{"CKW-SKL-CI10-PS", "Cast Iron Skillet 10 inch, Pre-Seasoned", 39.95},
	{"HOM-THR-WOOL-RWG-130", "Recycled Wool Throw Blanket", 79.00},
	{"INP-KB-MCH-TAC-RGB", "Mechanical Keyboard, Tactile Switches", 119.00},
}

type config struct {
	MongoURI string
	MongoDB  string
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
	}
}

func main() {
	cfg := loadConfig()
	ctx := context.Background()

	lazy := platformmongo.NewLazyClient(cfg.MongoURI)
	defer lazy.Disconnect()

	repo := productadapter.NewMongoRepository(lazy, cfg.MongoDB)

	for i, row := range seedRows {
		p := &productdomain.Product{
			SKU:   row.sku,
			Name:  row.name,
			Price: row.price,
		}
		res := repo.Create(ctx, p)
		if res.IsError() {
			log.Fatalf("seed product %s (%d): %v", row.sku, i+1, res.Error())
		}
	}

	log.Printf("inserted 10 products into db %q collection products", cfg.MongoDB)
}
