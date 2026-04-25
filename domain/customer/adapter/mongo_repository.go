package adapter

import (
	"context"
	"time"

	"github.com/samber/mo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	mongodriver "go.mongodb.org/mongo-driver/mongo"

	"modmono/domain/customer/domain"
	"modmono/domain/customer/port"
	platformmongo "modmono/domain/platform/mongo"
)

// MongoRepository is the Adapter — implements port.Repository against MongoDB.
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

func currentStatePipeline(status string, limit int64) mongodriver.Pipeline {
	return mongodriver.Pipeline{
		{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "customers"},
			{Key: "localField", Value: "_id"},
			{Key: "foreignField", Value: "original_id"},
			{Key: "as", Value: "_successors"},
		}}},
		{{Key: "$match", Value: bson.M{
			"_successors": bson.M{"$size": 0},
			"status":      status,
		}}},
		{{Key: "$sort", Value: bson.D{{Key: "created_at", Value: -1}}}},
		{{Key: "$limit", Value: limit}},
	}
}

func (r *MongoRepository) Create(ctx context.Context, c *domain.Customer) mo.Result[*domain.Customer] {
	coll, err := r.collection(ctx)
	if err != nil {
		return mo.Err[*domain.Customer](err)
	}
	if c.ID.IsZero() {
		c.ID = primitive.NewObjectID()
	}
	if c.CreatedAt.IsZero() {
		c.CreatedAt = c.ID.Timestamp()
	}
	c.Status = domain.StatusActive
	_, err = coll.InsertOne(ctx, c)
	if err != nil {
		return mo.Err[*domain.Customer](err)
	}
	return mo.Ok(c)
}

func (r *MongoRepository) GetByID(ctx context.Context, id primitive.ObjectID) mo.Result[mo.Option[domain.Customer]] {
	coll, err := r.collection(ctx)
	if err != nil {
		return mo.Err[mo.Option[domain.Customer]](err)
	}
	var out domain.Customer
	err = coll.FindOne(ctx, bson.M{"_id": id}).Decode(&out)
	if err != nil {
		if err == mongodriver.ErrNoDocuments {
			return mo.Ok(mo.None[domain.Customer]())
		}
		return mo.Err[mo.Option[domain.Customer]](err)
	}
	return mo.Ok(mo.Some(out))
}

func (r *MongoRepository) List(ctx context.Context, limit int64) mo.Result[[]domain.Customer] {
	coll, err := r.collection(ctx)
	if err != nil {
		return mo.Err[[]domain.Customer](err)
	}
	if limit <= 0 {
		limit = 50
	}
	cur, err := coll.Aggregate(ctx, currentStatePipeline(domain.StatusActive, limit))
	if err != nil {
		return mo.Err[[]domain.Customer](err)
	}
	return decodeCustomers(ctx, cur)
}

func (r *MongoRepository) ListInactive(ctx context.Context, limit int64) mo.Result[[]domain.Customer] {
	coll, err := r.collection(ctx)
	if err != nil {
		return mo.Err[[]domain.Customer](err)
	}
	if limit <= 0 {
		limit = 50
	}
	cur, err := coll.Aggregate(ctx, currentStatePipeline(domain.StatusDeactivated, limit))
	if err != nil {
		return mo.Err[[]domain.Customer](err)
	}
	return decodeCustomers(ctx, cur)
}

func (r *MongoRepository) Deactivate(ctx context.Context, id primitive.ObjectID, at time.Time) mo.Result[*domain.Customer] {
	coll, err := r.collection(ctx)
	if err != nil {
		return mo.Err[*domain.Customer](err)
	}
	var orig domain.Customer
	if err := coll.FindOne(ctx, bson.M{"_id": id}).Decode(&orig); err != nil {
		if err == mongodriver.ErrNoDocuments {
			return mo.Err[*domain.Customer](port.ErrNotFound)
		}
		return mo.Err[*domain.Customer](err)
	}
	next := &domain.Customer{
		ID:            primitive.NewObjectID(),
		Name:          orig.Name,
		Email:         orig.Email,
		Phone:         orig.Phone,
		Status:        domain.StatusDeactivated,
		OriginalID:    &orig.ID,
		CreatedAt:     orig.CreatedAt,
		DeactivatedAt: &at,
	}
	if _, err := coll.InsertOne(ctx, next); err != nil {
		return mo.Err[*domain.Customer](err)
	}
	return mo.Ok(next)
}

func decodeCustomers(ctx context.Context, cur *mongodriver.Cursor) mo.Result[[]domain.Customer] {
	defer cur.Close(ctx)
	var items []domain.Customer
	for cur.Next(ctx) {
		var c domain.Customer
		if err := cur.Decode(&c); err != nil {
			return mo.Err[[]domain.Customer](err)
		}
		items = append(items, c)
	}
	if err := cur.Err(); err != nil {
		return mo.Err[[]domain.Customer](err)
	}
	return mo.Ok(items)
}
