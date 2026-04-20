package customer

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

// ErrNotFound is returned when no customer exists for the given id.
var ErrNotFound = errors.New("customer: not found")

// Repository persists customers.
type Repository interface {
	Create(ctx context.Context, c *Customer) mo.Result[*Customer]
	GetByID(ctx context.Context, id primitive.ObjectID) (mo.Option[Customer], error)
	List(ctx context.Context, limit int64) mo.Result[[]Customer]
	ListInactive(ctx context.Context, limit int64) mo.Result[[]Customer]
	Deactivate(ctx context.Context, id primitive.ObjectID, at time.Time) mo.Result[*Customer]
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
	return client.Database(r.dbName).Collection("customers"), nil
}

// Create inserts a customer and returns it with server-assigned fields.
func (r *MongoRepository) Create(ctx context.Context, c *Customer) mo.Result[*Customer] {
	coll, err := r.collection(ctx)
	if err != nil {
		return mo.Err[*Customer](err)
	}
	if c.ID.IsZero() {
		c.ID = primitive.NewObjectID()
	}
	if c.CreatedAt.IsZero() {
		c.CreatedAt = c.ID.Timestamp()
	}
	_, err = coll.InsertOne(ctx, c)
	if err != nil {
		return mo.Err[*Customer](err)
	}
	return mo.Ok(c)
}

// GetByID returns Some(customer) if found, None if missing, or an error for infrastructure failures.
func (r *MongoRepository) GetByID(ctx context.Context, id primitive.ObjectID) (mo.Option[Customer], error) {
	coll, err := r.collection(ctx)
	if err != nil {
		return mo.None[Customer](), err
	}
	var out Customer
	err = coll.FindOne(ctx, bson.M{"_id": id}).Decode(&out)
	if err != nil {
		if err == mongodriver.ErrNoDocuments {
			return mo.None[Customer](), nil
		}
		return mo.None[Customer](), err
	}
	return mo.Some(out), nil
}

// List returns up to limit active customers, newest first.
func (r *MongoRepository) List(ctx context.Context, limit int64) mo.Result[[]Customer] {
	coll, err := r.collection(ctx)
	if err != nil {
		return mo.Err[[]Customer](err)
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
		return mo.Err[[]Customer](err)
	}
	defer cur.Close(ctx)

	var items []Customer
	for cur.Next(ctx) {
		var c Customer
		if err := cur.Decode(&c); err != nil {
			return mo.Err[[]Customer](err)
		}
		items = append(items, c)
	}
	if err := cur.Err(); err != nil {
		return mo.Err[[]Customer](err)
	}
	return mo.Ok(items)
}

// ListInactive returns up to limit deactivated customers, newest deactivation first.
func (r *MongoRepository) ListInactive(ctx context.Context, limit int64) mo.Result[[]Customer] {
	coll, err := r.collection(ctx)
	if err != nil {
		return mo.Err[[]Customer](err)
	}
	if limit <= 0 {
		limit = 50
	}
	inactive := bson.M{"deactivated_at": bson.M{"$exists": true, "$ne": nil}}
	opts := options.Find().SetSort(bson.D{{Key: "deactivated_at", Value: -1}}).SetLimit(limit)
	cur, err := coll.Find(ctx, inactive, opts)
	if err != nil {
		return mo.Err[[]Customer](err)
	}
	defer cur.Close(ctx)

	var items []Customer
	for cur.Next(ctx) {
		var c Customer
		if err := cur.Decode(&c); err != nil {
			return mo.Err[[]Customer](err)
		}
		items = append(items, c)
	}
	if err := cur.Err(); err != nil {
		return mo.Err[[]Customer](err)
	}
	return mo.Ok(items)
}

// Deactivate sets deactivated_at and returns the updated document. Idempotent: repeated calls refresh the timestamp.
func (r *MongoRepository) Deactivate(ctx context.Context, id primitive.ObjectID, at time.Time) mo.Result[*Customer] {
	coll, err := r.collection(ctx)
	if err != nil {
		return mo.Err[*Customer](err)
	}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var out Customer
	err = coll.FindOneAndUpdate(ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"deactivated_at": at}},
		opts,
	).Decode(&out)
	if err != nil {
		if err == mongodriver.ErrNoDocuments {
			return mo.Err[*Customer](ErrNotFound)
		}
		return mo.Err[*Customer](err)
	}
	return mo.Ok(&out)
}
