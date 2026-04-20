package product

import (
	"context"
	"errors"
	"time"

	"github.com/samber/mo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	mongodriver "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	platformmongo "modmono/internal/platform/mongo"
)

// ErrNotFound is returned when no product exists for the given id.
var ErrNotFound = errors.New("product: not found")

// Repository persists products.
type Repository interface {
	Create(ctx context.Context, p *Product) mo.Result[*Product]
	GetByID(ctx context.Context, id primitive.ObjectID) (mo.Option[Product], error)
	List(ctx context.Context, limit int64) mo.Result[[]Product]
	ListInactive(ctx context.Context, limit int64) mo.Result[[]Product]
	Deactivate(ctx context.Context, id primitive.ObjectID, at time.Time) mo.Result[*Product]
}

// MongoRepository implements Repository using a MongoDB collection.
type MongoRepository struct {
	lazy   *platformmongo.LazyClient
	dbName string
}

// NewMongoRepository builds a Mongo-backed repository. The lazy client connects on first use.
func NewMongoRepository(lazy *platformmongo.LazyClient, dbName string) *MongoRepository {
	return &MongoRepository{lazy: lazy, dbName: dbName}
}

func (r *MongoRepository) collection(ctx context.Context) (*mongodriver.Collection, error) {
	client, err := r.lazy.Get(ctx)
	if err != nil {
		return nil, err
	}
	return client.Database(r.dbName).Collection("products"), nil
}

// Create inserts a product and returns it with server-assigned fields.
func (r *MongoRepository) Create(ctx context.Context, p *Product) mo.Result[*Product] {
	coll, err := r.collection(ctx)
	if err != nil {
		return mo.Err[*Product](err)
	}
	if p.ID.IsZero() {
		p.ID = primitive.NewObjectID()
	}
	if p.CreatedAt.IsZero() {
		p.CreatedAt = p.ID.Timestamp()
	}
	_, err = coll.InsertOne(ctx, p)
	if err != nil {
		return mo.Err[*Product](err)
	}
	return mo.Ok(p)
}

// GetByID returns Some(product) if found, None if missing, or an error for infrastructure failures.
func (r *MongoRepository) GetByID(ctx context.Context, id primitive.ObjectID) (mo.Option[Product], error) {
	coll, err := r.collection(ctx)
	if err != nil {
		return mo.None[Product](), err
	}
	var out Product
	err = coll.FindOne(ctx, bson.M{"_id": id}).Decode(&out)
	if err != nil {
		if err == mongodriver.ErrNoDocuments {
			return mo.None[Product](), nil
		}
		return mo.None[Product](), err
	}
	return mo.Some(out), nil
}

// List returns up to limit products, newest first.
func (r *MongoRepository) List(ctx context.Context, limit int64) mo.Result[[]Product] {
	coll, err := r.collection(ctx)
	if err != nil {
		return mo.Err[[]Product](err)
	}
	if limit <= 0 {
		limit = 50
	}
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).SetLimit(limit)
	activeOnly := bson.M{"$or": []bson.M{
		{"deactivated_at": bson.M{"$exists": false}},
		{"deactivated_at": nil},
	}}
	cur, err := coll.Find(ctx, activeOnly, opts)
	if err != nil {
		return mo.Err[[]Product](err)
	}
	defer cur.Close(ctx)

	var items []Product
	for cur.Next(ctx) {
		var p Product
		if err := cur.Decode(&p); err != nil {
			return mo.Err[[]Product](err)
		}
		items = append(items, p)
	}
	if err := cur.Err(); err != nil {
		return mo.Err[[]Product](err)
	}
	return mo.Ok(items)
}

// ListInactive returns up to limit deactivated products, newest deactivation first (by deactivated_at desc, then created_at).
func (r *MongoRepository) ListInactive(ctx context.Context, limit int64) mo.Result[[]Product] {
	coll, err := r.collection(ctx)
	if err != nil {
		return mo.Err[[]Product](err)
	}
	if limit <= 0 {
		limit = 50
	}
	inactive := bson.M{"deactivated_at": bson.M{"$exists": true, "$ne": nil}}
	opts := options.Find().SetSort(bson.D{{Key: "deactivated_at", Value: -1}}).SetLimit(limit)
	cur, err := coll.Find(ctx, inactive, opts)
	if err != nil {
		return mo.Err[[]Product](err)
	}
	defer cur.Close(ctx)

	var items []Product
	for cur.Next(ctx) {
		var p Product
		if err := cur.Decode(&p); err != nil {
			return mo.Err[[]Product](err)
		}
		items = append(items, p)
	}
	if err := cur.Err(); err != nil {
		return mo.Err[[]Product](err)
	}
	return mo.Ok(items)
}

// Deactivate sets deactivated_at and returns the updated document. Idempotent: repeated calls refresh the timestamp.
func (r *MongoRepository) Deactivate(ctx context.Context, id primitive.ObjectID, at time.Time) mo.Result[*Product] {
	coll, err := r.collection(ctx)
	if err != nil {
		return mo.Err[*Product](err)
	}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var out Product
	err = coll.FindOneAndUpdate(ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"deactivated_at": at}},
		opts,
	).Decode(&out)
	if err != nil {
		if err == mongodriver.ErrNoDocuments {
			return mo.Err[*Product](ErrNotFound)
		}
		return mo.Err[*Product](err)
	}
	return mo.Ok(&out)
}
