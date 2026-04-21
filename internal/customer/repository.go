package customer

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

// ErrNotFound is returned when no customer exists for the given id.
var ErrNotFound = errors.New("customer: not found")

// Repository persists customers.
type Repository interface {
	Create(ctx context.Context, c *Customer) mo.Result[*Customer]
	GetByID(ctx context.Context, id primitive.ObjectID) mo.Result[mo.Option[Customer]]
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

// currentStatePipeline returns an aggregation pipeline that filters to non-superseded records
// with the given status, sorted newest first and capped at limit.
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
	c.Status = StatusActive
	_, err = coll.InsertOne(ctx, c)
	if err != nil {
		return mo.Err[*Customer](err)
	}
	return mo.Ok(c)
}

// GetByID returns Ok(Some(customer)) if found, Ok(None) if missing, or Err for infrastructure failures.
func (r *MongoRepository) GetByID(ctx context.Context, id primitive.ObjectID) mo.Result[mo.Option[Customer]] {
	coll, err := r.collection(ctx)
	if err != nil {
		return mo.Err[mo.Option[Customer]](err)
	}
	var out Customer
	err = coll.FindOne(ctx, bson.M{"_id": id}).Decode(&out)
	if err != nil {
		if err == mongodriver.ErrNoDocuments {
			return mo.Ok(mo.None[Customer]())
		}
		return mo.Err[mo.Option[Customer]](err)
	}
	return mo.Ok(mo.Some(out))
}

// List returns up to limit active (non-superseded) customers, newest first.
func (r *MongoRepository) List(ctx context.Context, limit int64) mo.Result[[]Customer] {
	coll, err := r.collection(ctx)
	if err != nil {
		return mo.Err[[]Customer](err)
	}
	if limit <= 0 {
		limit = 50
	}
	cur, err := coll.Aggregate(ctx, currentStatePipeline(StatusActive, limit))
	if err != nil {
		return mo.Err[[]Customer](err)
	}
	return decodeCustomers(ctx, cur)
}

// ListInactive returns up to limit deactivated (non-superseded) customers, newest first.
func (r *MongoRepository) ListInactive(ctx context.Context, limit int64) mo.Result[[]Customer] {
	coll, err := r.collection(ctx)
	if err != nil {
		return mo.Err[[]Customer](err)
	}
	if limit <= 0 {
		limit = 50
	}
	cur, err := coll.Aggregate(ctx, currentStatePipeline(StatusDeactivated, limit))
	if err != nil {
		return mo.Err[[]Customer](err)
	}
	return decodeCustomers(ctx, cur)
}

// Deactivate creates a new deactivated record linked to the original. No records are mutated.
func (r *MongoRepository) Deactivate(ctx context.Context, id primitive.ObjectID, at time.Time) mo.Result[*Customer] {
	coll, err := r.collection(ctx)
	if err != nil {
		return mo.Err[*Customer](err)
	}
	var orig Customer
	if err := coll.FindOne(ctx, bson.M{"_id": id}).Decode(&orig); err != nil {
		if err == mongodriver.ErrNoDocuments {
			return mo.Err[*Customer](ErrNotFound)
		}
		return mo.Err[*Customer](err)
	}
	next := &Customer{
		ID:            primitive.NewObjectID(),
		Name:          orig.Name,
		Email:         orig.Email,
		Phone:         orig.Phone,
		Status:        StatusDeactivated,
		OriginalID:    &orig.ID,
		CreatedAt:     orig.CreatedAt,
		DeactivatedAt: &at,
	}
	if _, err := coll.InsertOne(ctx, next); err != nil {
		return mo.Err[*Customer](err)
	}
	return mo.Ok(next)
}

// decodeCustomers drains a cursor into a slice of Customers.
func decodeCustomers(ctx context.Context, cur *mongodriver.Cursor) mo.Result[[]Customer] {
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
