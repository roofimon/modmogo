package adapter

import (
	"context"
	"time"

	"github.com/samber/mo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	mongodriver "go.mongodb.org/mongo-driver/mongo"

	"modmono/internal/order/domain"
	"modmono/internal/order/port"
	platformmongo "modmono/internal/platform/mongo"
)

// MongoRepository is the Adapter — implements port.Repository against MongoDB.
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

func (r *MongoRepository) Create(ctx context.Context, o *domain.Order) mo.Result[*domain.Order] {
	coll, err := r.collection(ctx)
	if err != nil {
		return mo.Err[*domain.Order](err)
	}
	if o.ID.IsZero() {
		o.ID = primitive.NewObjectID()
	}
	if o.CreatedAt.IsZero() {
		o.CreatedAt = o.ID.Timestamp()
	}
	_, err = coll.InsertOne(ctx, o)
	if err != nil {
		return mo.Err[*domain.Order](err)
	}
	o.ComputeTotal()
	return mo.Ok(o)
}

func (r *MongoRepository) GetByID(ctx context.Context, id primitive.ObjectID) mo.Result[mo.Option[domain.Order]] {
	coll, err := r.collection(ctx)
	if err != nil {
		return mo.Err[mo.Option[domain.Order]](err)
	}
	var out domain.Order
	err = coll.FindOne(ctx, bson.M{"_id": id}).Decode(&out)
	if err != nil {
		if err == mongodriver.ErrNoDocuments {
			return mo.Ok(mo.None[domain.Order]())
		}
		return mo.Err[mo.Option[domain.Order]](err)
	}
	out.ComputeTotal()
	return mo.Ok(mo.Some(out))
}

func (r *MongoRepository) List(ctx context.Context, limit int64) mo.Result[[]domain.Order] {
	coll, err := r.collection(ctx)
	if err != nil {
		return mo.Err[[]domain.Order](err)
	}
	if limit <= 0 {
		limit = 50
	}
	activeFilter := bson.M{
		"original_order_id": bson.M{"$exists": false},
		"deactivated_at":    bson.M{"$exists": false},
	}
	cur, err := coll.Aggregate(ctx, currentStatePipeline(activeFilter, limit))
	if err != nil {
		return mo.Err[[]domain.Order](err)
	}
	return decodeOrders(ctx, cur)
}

func (r *MongoRepository) ListInactive(ctx context.Context, limit int64) mo.Result[[]domain.Order] {
	coll, err := r.collection(ctx)
	if err != nil {
		return mo.Err[[]domain.Order](err)
	}
	if limit <= 0 {
		limit = 50
	}
	cur, err := coll.Aggregate(ctx, currentStatePipeline(bson.M{"deactivated_at": bson.M{"$exists": true}}, limit))
	if err != nil {
		return mo.Err[[]domain.Order](err)
	}
	return decodeOrders(ctx, cur)
}

func (r *MongoRepository) ListPaymentCompleted(ctx context.Context, limit int64) mo.Result[[]domain.Order] {
	coll, err := r.collection(ctx)
	if err != nil {
		return mo.Err[[]domain.Order](err)
	}
	if limit <= 0 {
		limit = 50
	}
	cur, err := coll.Aggregate(ctx, currentStatePipeline(bson.M{"status": domain.StatusPaymentCompleted}, limit))
	if err != nil {
		return mo.Err[[]domain.Order](err)
	}
	return decodeOrders(ctx, cur)
}

func (r *MongoRepository) Deactivate(ctx context.Context, id primitive.ObjectID, at time.Time) mo.Result[*domain.Order] {
	coll, err := r.collection(ctx)
	if err != nil {
		return mo.Err[*domain.Order](err)
	}
	var orig domain.Order
	if err := coll.FindOne(ctx, bson.M{"_id": id}).Decode(&orig); err != nil {
		if err == mongodriver.ErrNoDocuments {
			return mo.Err[*domain.Order](port.ErrNotFound)
		}
		return mo.Err[*domain.Order](err)
	}
	next := &domain.Order{
		ID:              primitive.NewObjectID(),
		CustomerID:      orig.CustomerID,
		Items:           orig.Items,
		Status:          domain.StatusDeactivated,
		OriginalOrderID: &orig.ID,
		CreatedAt:       orig.CreatedAt,
		DeactivatedAt:   &at,
	}
	if _, err := coll.InsertOne(ctx, next); err != nil {
		return mo.Err[*domain.Order](err)
	}
	next.ComputeTotal()
	return mo.Ok(next)
}

func decodeOrders(ctx context.Context, cur *mongodriver.Cursor) mo.Result[[]domain.Order] {
	defer cur.Close(ctx)
	var items []domain.Order
	for cur.Next(ctx) {
		var o domain.Order
		if err := cur.Decode(&o); err != nil {
			return mo.Err[[]domain.Order](err)
		}
		o.ComputeTotal()
		items = append(items, o)
	}
	if err := cur.Err(); err != nil {
		return mo.Err[[]domain.Order](err)
	}
	return mo.Ok(items)
}
