package order

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

// ErrNotFound is returned when no order exists for the given id.
var ErrNotFound = errors.New("order: not found")

// Repository persists orders.
type Repository interface {
	Create(ctx context.Context, o *Order) mo.Result[*Order]
	GetByID(ctx context.Context, id primitive.ObjectID) (mo.Option[Order], error)
	List(ctx context.Context, limit int64) mo.Result[[]Order]
	ListInactive(ctx context.Context, limit int64) mo.Result[[]Order]
	Deactivate(ctx context.Context, id primitive.ObjectID, at time.Time) mo.Result[*Order]
}

// MongoRepository implements Repository using a MongoDB collection.
type MongoRepository struct {
	lazy   *platformmongo.LazyClient
	dbName string
}

// NewMongoRepository builds a Mongo-backed repository.
func NewMongoRepository(lazy *platformmongo.LazyClient, dbName string) *MongoRepository {
	return &MongoRepository{lazy: lazy, dbName: dbName}
}

func (r *MongoRepository) collection(ctx context.Context) (*mongodriver.Collection, error) {
	client, err := r.lazy.Get(ctx)
	if err != nil {
		return nil, err
	}
	return client.Database(r.dbName).Collection("orders"), nil
}

// Create inserts an order and returns it with server-assigned fields.
func (r *MongoRepository) Create(ctx context.Context, o *Order) mo.Result[*Order] {
	coll, err := r.collection(ctx)
	if err != nil {
		return mo.Err[*Order](err)
	}
	if o.ID.IsZero() {
		o.ID = primitive.NewObjectID()
	}
	if o.CreatedAt.IsZero() {
		o.CreatedAt = o.ID.Timestamp()
	}
	_, err = coll.InsertOne(ctx, o)
	if err != nil {
		return mo.Err[*Order](err)
	}
	o.ComputeTotal()
	return mo.Ok(o)
}

// GetByID returns Some(order) if found, None if missing, or an error for infrastructure failures.
func (r *MongoRepository) GetByID(ctx context.Context, id primitive.ObjectID) (mo.Option[Order], error) {
	coll, err := r.collection(ctx)
	if err != nil {
		return mo.None[Order](), err
	}
	var out Order
	err = coll.FindOne(ctx, bson.M{"_id": id}).Decode(&out)
	if err != nil {
		if err == mongodriver.ErrNoDocuments {
			return mo.None[Order](), nil
		}
		return mo.None[Order](), err
	}
	out.ComputeTotal()
	return mo.Some(out), nil
}

// List returns up to limit active orders, newest first.
func (r *MongoRepository) List(ctx context.Context, limit int64) mo.Result[[]Order] {
	coll, err := r.collection(ctx)
	if err != nil {
		return mo.Err[[]Order](err)
	}
	if limit <= 0 {
		limit = 50
	}
	activeOnly := bson.M{"$or": []bson.M{
		{"deactivated_at": bson.M{"$exists": false}},
		{"deactivated_at": nil},
	}}
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).SetLimit(limit)
	cur, err := coll.Find(ctx, activeOnly, opts)
	if err != nil {
		return mo.Err[[]Order](err)
	}
	defer cur.Close(ctx)

	var items []Order
	for cur.Next(ctx) {
		var o Order
		if err := cur.Decode(&o); err != nil {
			return mo.Err[[]Order](err)
		}
		o.ComputeTotal()
		items = append(items, o)
	}
	if err := cur.Err(); err != nil {
		return mo.Err[[]Order](err)
	}
	return mo.Ok(items)
}

// ListInactive returns up to limit deactivated orders, newest deactivation first.
func (r *MongoRepository) ListInactive(ctx context.Context, limit int64) mo.Result[[]Order] {
	coll, err := r.collection(ctx)
	if err != nil {
		return mo.Err[[]Order](err)
	}
	if limit <= 0 {
		limit = 50
	}
	inactive := bson.M{"deactivated_at": bson.M{"$exists": true, "$ne": nil}}
	opts := options.Find().SetSort(bson.D{{Key: "deactivated_at", Value: -1}}).SetLimit(limit)
	cur, err := coll.Find(ctx, inactive, opts)
	if err != nil {
		return mo.Err[[]Order](err)
	}
	defer cur.Close(ctx)

	var items []Order
	for cur.Next(ctx) {
		var o Order
		if err := cur.Decode(&o); err != nil {
			return mo.Err[[]Order](err)
		}
		o.ComputeTotal()
		items = append(items, o)
	}
	if err := cur.Err(); err != nil {
		return mo.Err[[]Order](err)
	}
	return mo.Ok(items)
}

// Deactivate sets deactivated_at and returns the updated document.
func (r *MongoRepository) Deactivate(ctx context.Context, id primitive.ObjectID, at time.Time) mo.Result[*Order] {
	coll, err := r.collection(ctx)
	if err != nil {
		return mo.Err[*Order](err)
	}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var out Order
	err = coll.FindOneAndUpdate(ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"deactivated_at": at}},
		opts,
	).Decode(&out)
	if err != nil {
		if err == mongodriver.ErrNoDocuments {
			return mo.Err[*Order](ErrNotFound)
		}
		return mo.Err[*Order](err)
	}
	out.ComputeTotal()
	return mo.Ok(&out)
}
