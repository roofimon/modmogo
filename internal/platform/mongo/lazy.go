package mongo

import (
	"context"
	"log"
	"sync"
	"time"

	mongodriver "go.mongodb.org/mongo-driver/mongo"
)

// LazyClient defers Connect until the first Get. Safe for concurrent use.
type LazyClient struct {
	uri    string
	mu     sync.Mutex
	client *mongodriver.Client
}

// NewLazyClient returns a client that does not dial Mongo until Get is called.
func NewLazyClient(uri string) *LazyClient {
	return &LazyClient{uri: uri}
}

// Get returns the shared client, connecting on first successful call.
func (l *LazyClient) Get(ctx context.Context) (*mongodriver.Client, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.client != nil {
		return l.client, nil
	}
	// First connection uses ConnectIO so the dial is modeled as mo.IOEither.
	either := ConnectIO(ctx, l.uri).Run()
	if either.IsLeft() {
		return nil, either.MustLeft()
	}
	l.client = either.MustRight()
	return l.client, nil
}

// Disconnect closes the client if it was connected.
func (l *LazyClient) Disconnect() {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.client == nil {
		return
	}
	dctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := l.client.Disconnect(dctx); err != nil {
		log.Printf("mongo disconnect: %v", err)
	}
	l.client = nil
}
