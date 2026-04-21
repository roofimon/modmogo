package product

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

// ErrNotFound is returned when no product exists for the given id.
var ErrNotFound = errors.New("product: not found")

// Repository persists products.
type Repository interface {
	Create(ctx context.Context, p *Product) mo.Result[*Product]
	GetByID(ctx context.Context, id primitive.ObjectID) mo.Result[mo.Option[Product]]
	GetBySKU(ctx context.Context, sku string) mo.Result[mo.Option[Product]]
	List(ctx context.Context, limit int64) mo.Result[[]Product]
	ListInactive(ctx context.Context, limit int64) mo.Result[[]Product]
	Deactivate(ctx context.Context, id primitive.ObjectID, at time.Time) mo.Result[*Product]
	Activate(ctx context.Context, id primitive.ObjectID) mo.Result[*Product]
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

// currentStatePipeline returns an aggregation pipeline that filters to non-superseded records
// with the given status, sorted newest first and capped at limit.
func currentStatePipeline(status string, limit int64) mongodriver.Pipeline {
	return mongodriver.Pipeline{
		{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "products"},
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
	p.Status = StatusActive
	_, err = coll.InsertOne(ctx, p)
	if err != nil {
		return mo.Err[*Product](err)
	}
	return mo.Ok(p)
}

// GetByID returns Ok(Some(product)) if found, Ok(None) if missing, or Err for infrastructure failures.
func (r *MongoRepository) GetByID(ctx context.Context, id primitive.ObjectID) mo.Result[mo.Option[Product]] {
	coll, err := r.collection(ctx)
	if err != nil {
		return mo.Err[mo.Option[Product]](err)
	}
	var out Product
	err = coll.FindOne(ctx, bson.M{"_id": id}).Decode(&out)
	if err != nil {
		if err == mongodriver.ErrNoDocuments {
			return mo.Ok(mo.None[Product]())
		}
		return mo.Err[mo.Option[Product]](err)
	}
	return mo.Ok(mo.Some(out))
}

// GetBySKU returns Ok(Some(product)) if found by SKU, Ok(None) if missing, or Err for infrastructure failures.
func (r *MongoRepository) GetBySKU(ctx context.Context, sku string) mo.Result[mo.Option[Product]] {
	coll, err := r.collection(ctx)
	if err != nil {
		return mo.Err[mo.Option[Product]](err)
	}
	var out Product
	err = coll.FindOne(ctx, bson.M{"sku": sku, "status": StatusActive}).Decode(&out)
	if err != nil {
		if err == mongodriver.ErrNoDocuments {
			return mo.Ok(mo.None[Product]())
		}
		return mo.Err[mo.Option[Product]](err)
	}
	return mo.Ok(mo.Some(out))
}

// List returns up to limit active (non-superseded) products, newest first.
func (r *MongoRepository) List(ctx context.Context, limit int64) mo.Result[[]Product] {
	coll, err := r.collection(ctx)
	if err != nil {
		return mo.Err[[]Product](err)
	}
	if limit <= 0 {
		limit = 50
	}
	cur, err := coll.Aggregate(ctx, currentStatePipeline(StatusActive, limit))
	if err != nil {
		return mo.Err[[]Product](err)
	}
	return decodeProducts(ctx, cur)
}

// ListInactive returns up to limit deactivated (non-superseded) products, newest first.
func (r *MongoRepository) ListInactive(ctx context.Context, limit int64) mo.Result[[]Product] {
	coll, err := r.collection(ctx)
	if err != nil {
		return mo.Err[[]Product](err)
	}
	if limit <= 0 {
		limit = 50
	}
	cur, err := coll.Aggregate(ctx, currentStatePipeline(StatusDeactivated, limit))
	if err != nil {
		return mo.Err[[]Product](err)
	}
	return decodeProducts(ctx, cur)
}

// Deactivate creates a new deactivated record linked to the original. No records are mutated.
func (r *MongoRepository) Deactivate(ctx context.Context, id primitive.ObjectID, at time.Time) mo.Result[*Product] {
	coll, err := r.collection(ctx)
	if err != nil {
		return mo.Err[*Product](err)
	}
	var orig Product
	if err := coll.FindOne(ctx, bson.M{"_id": id}).Decode(&orig); err != nil {
		if err == mongodriver.ErrNoDocuments {
			return mo.Err[*Product](ErrNotFound)
		}
		return mo.Err[*Product](err)
	}
	next := &Product{
		ID:            primitive.NewObjectID(),
		SKU:           orig.SKU,
		Name:          orig.Name,
		Price:         orig.Price,
		Status:        StatusDeactivated,
		OriginalID:    &orig.ID,
		CreatedAt:     orig.CreatedAt,
		DeactivatedAt: &at,
	}
	if _, err := coll.InsertOne(ctx, next); err != nil {
		return mo.Err[*Product](err)
	}
	return mo.Ok(next)
}

// Activate creates a new active record linked to the original. No records are mutated.
func (r *MongoRepository) Activate(ctx context.Context, id primitive.ObjectID) mo.Result[*Product] {
	coll, err := r.collection(ctx)
	if err != nil {
		return mo.Err[*Product](err)
	}
	var orig Product
	if err := coll.FindOne(ctx, bson.M{"_id": id}).Decode(&orig); err != nil {
		if err == mongodriver.ErrNoDocuments {
			return mo.Err[*Product](ErrNotFound)
		}
		return mo.Err[*Product](err)
	}
	next := &Product{
		ID:         primitive.NewObjectID(),
		SKU:        orig.SKU,
		Name:       orig.Name,
		Price:      orig.Price,
		Status:     StatusActive,
		OriginalID: &orig.ID,
		CreatedAt:  orig.CreatedAt,
	}
	if _, err := coll.InsertOne(ctx, next); err != nil {
		return mo.Err[*Product](err)
	}
	return mo.Ok(next)
}

// decodeProducts drains a cursor into a slice of Products.
func decodeProducts(ctx context.Context, cur *mongodriver.Cursor) mo.Result[[]Product] {
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

