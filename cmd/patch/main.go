package main

import (
	"context"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/bson"
	platformmongo "modmono/domain/platform/mongo"
)

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

	client, err := lazy.Get(ctx)
	if err != nil {
		log.Fatalf("mongo connect: %v", err)
	}
	db := client.Database(cfg.MongoDB)

	// Backfill status="active" on legacy documents that predate the status field.
	legacyFilter := bson.M{"$or": []bson.M{
		{"status": bson.M{"$exists": false}},
		{"status": ""},
	}}
	update := bson.M{"$set": bson.M{"status": "active"}}

	for _, collName := range []string{"products", "customers"} {
		res, err := db.Collection(collName).UpdateMany(ctx, legacyFilter, update)
		if err != nil {
			log.Fatalf("patch %s: %v", collName, err)
		}
		log.Printf("patched %d %s document(s): set status=active", res.ModifiedCount, collName)
	}
}
