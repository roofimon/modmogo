package mongo

import (
	"context"
	"time"

	"github.com/samber/mo"
	mongodriver "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Connect dials MongoDB and pings before returning.
func Connect(ctx context.Context, uri string) (*mongodriver.Client, error) {
	opts := options.Client().ApplyURI(uri)
	client, err := mongodriver.Connect(ctx, opts)
	if err != nil {
		return nil, err
	}
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := client.Ping(pingCtx, nil); err != nil {
		_ = client.Disconnect(context.Background())
		return nil, err
	}
	return client, nil
}

// ConnectIO wraps Connect as a fallible IO action. Nothing runs until Run().
func ConnectIO(ctx context.Context, uri string) mo.IOEither[*mongodriver.Client] {
	return mo.NewIOEither(func() (*mongodriver.Client, error) {
		return Connect(ctx, uri)
	})
}
