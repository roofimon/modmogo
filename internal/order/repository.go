package order

import (
	"context"
	"errors"
	"time"

	"github.com/samber/mo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	mongodriver "go.mongodb.org/mongo-driver/mongo"

	platformmongo "modmono/internal/platform/mongo"
)

// ErrNotFound is returned when no order exists for the given id.
var ErrNotFound = errors.New("order: not found")

// Repository persists orders.
type Repository interface {
	Create(ctx context.Context, o *Order) mo.Result[*Order]
	GetByID(ctx context.Context, id primitive.ObjectID) mo.Result[mo.Option[Order]]
	List(ctx context.Context, limit int64) mo.Result[[]Order]
	ListInactive(ctx context.Context, limit int64) mo.Result[[]Order]
	ListPaymentCompleted(ctx context.Context, limit int64) mo.Result[[]Order]
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

// currentStatePipeline returns an aggregation pipeline that filters to non-superseded records
// matching statusMatch, sorted newest first and capped at limit.
func currentStatePipeline(statusMatch bson.M, limit int64) mongodriver.Pipeline {
	match := bson.M{"_successors": bson.M{"$size": 0}}
	for k, v := range statusMatch {
		match[k] = v
	}
	return mongodriver.Pipeline{
		{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "orders"},
			{Key: "localField", Value: "_id"},
			{Key: "foreignField", Value: "original_order_id"},
			{Key: "as", Value: "_successors"},
		}}},
		{{Key: "$match", Value: match}},
		{{Key: "$sort", Value: bson.D{{Key: "created_at", Value: -1}}}},
		{{Key: "$limit", Value: limit}},
	}
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

// GetByID returns Ok(Some(order)) if found, Ok(None) if missing, or Err for infrastructure failures.
func (r *MongoRepository) GetByID(ctx context.Context, id primitive.ObjectID) mo.Result[mo.Option[Order]] {
	coll, err := r.collection(ctx)
	if err != nil {
		return mo.Err[mo.Option[Order]](err)
	}
	var out Order
	err = coll.FindOne(ctx, bson.M{"_id": id}).Decode(&out)
	if err != nil {
		if err == mongodriver.ErrNoDocuments {
			return mo.Ok(mo.None[Order]())
		}
		return mo.Err[mo.Option[Order]](err)
	}
	out.ComputeTotal()
	return mo.Ok(mo.Some(out))
}

// List returns up to limit pending (non-superseded) orders, newest first.
func (r *MongoRepository) List(ctx context.Context, limit int64) mo.Result[[]Order] {
	coll, err := r.collection(ctx)
	if err != nil {
		return mo.Err[[]Order](err)
	}
	if limit <= 0 {
		limit = 50
	}
	statusFilter := bson.M{"$or": []bson.M{{"status": bson.M{"$exists": false}}, {"status": ""}}}
	cur, err := coll.Aggregate(ctx, currentStatePipeline(statusFilter, limit))
	if err != nil {
		return mo.Err[[]Order](err)
	}
	return decodeOrders(ctx, cur)
}

// ListInactive returns up to limit deactivated (non-superseded) orders, newest first.
func (r *MongoRepository) ListInactive(ctx context.Context, limit int64) mo.Result[[]Order] {
	coll, err := r.collection(ctx)
	if err != nil {
		return mo.Err[[]Order](err)
	}
	if limit <= 0 {
		limit = 50
	}
	cur, err := coll.Aggregate(ctx, currentStatePipeline(bson.M{"status": StatusDeactivated}, limit))
	if err != nil {
		return mo.Err[[]Order](err)
	}
	return decodeOrders(ctx, cur)
}

// ListPaymentCompleted returns up to limit payment-completed (non-superseded) orders, newest first.
func (r *MongoRepository) ListPaymentCompleted(ctx context.Context, limit int64) mo.Result[[]Order] {
	coll, err := r.collection(ctx)
	if err != nil {
		return mo.Err[[]Order](err)
	}
	if limit <= 0 {
		limit = 50
	}
	cur, err := coll.Aggregate(ctx, currentStatePipeline(bson.M{"status": StatusPaymentCompleted}, limit))
	if err != nil {
		return mo.Err[[]Order](err)
	}
	return decodeOrders(ctx, cur)
}

// Deactivate creates a new deactivated record linked to the original. No records are mutated.
func (r *MongoRepository) Deactivate(ctx context.Context, id primitive.ObjectID, at time.Time) mo.Result[*Order] {
	coll, err := r.collection(ctx)
	if err != nil {
		return mo.Err[*Order](err)
	}
	var orig Order
	if err := coll.FindOne(ctx, bson.M{"_id": id}).Decode(&orig); err != nil {
		if err == mongodriver.ErrNoDocuments {
			return mo.Err[*Order](ErrNotFound)
		}
		return mo.Err[*Order](err)
	}
	next := &Order{
		ID:              primitive.NewObjectID(),
		CustomerID:      orig.CustomerID,
		Items:           orig.Items,
		Status:          StatusDeactivated,
		OriginalOrderID: &orig.ID,
		CreatedAt:       orig.CreatedAt,
		DeactivatedAt:   &at,
	}
	if _, err := coll.InsertOne(ctx, next); err != nil {
		return mo.Err[*Order](err)
	}
	next.ComputeTotal()
	return mo.Ok(next)
}

// decodeOrders drains a cursor into a slice of Orders, computing totals on each.
func decodeOrders(ctx context.Context, cur *mongodriver.Cursor) mo.Result[[]Order] {
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
