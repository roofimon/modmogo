package adapter

import (
	"context"
	"time"

	"github.com/samber/mo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	mongodriver "go.mongodb.org/mongo-driver/mongo"

	"modmono/internal/product/domain"
	"modmono/internal/product/port"
	platformmongo "modmono/internal/platform/mongo"
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
	return client.Database(r.dbName).Collection("products"), nil
}

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

func (r *MongoRepository) Create(ctx context.Context, p *domain.Product) mo.Result[*domain.Product] {
	coll, err := r.collection(ctx)
	if err != nil {
		return mo.Err[*domain.Product](err)
	}
	if p.ID.IsZero() {
		p.ID = primitive.NewObjectID()
	}
	if p.CreatedAt.IsZero() {
		p.CreatedAt = p.ID.Timestamp()
	}
	p.Status = domain.StatusActive
	_, err = coll.InsertOne(ctx, p)
	if err != nil {
		return mo.Err[*domain.Product](err)
	}
	return mo.Ok(p)
}

func (r *MongoRepository) GetByID(ctx context.Context, id primitive.ObjectID) mo.Result[mo.Option[domain.Product]] {
	coll, err := r.collection(ctx)
	if err != nil {
		return mo.Err[mo.Option[domain.Product]](err)
	}
	var out domain.Product
	err = coll.FindOne(ctx, bson.M{"_id": id}).Decode(&out)
	if err != nil {
		if err == mongodriver.ErrNoDocuments {
			return mo.Ok(mo.None[domain.Product]())
		}
		return mo.Err[mo.Option[domain.Product]](err)
	}
	return mo.Ok(mo.Some(out))
}

func (r *MongoRepository) GetBySKU(ctx context.Context, sku string) mo.Result[mo.Option[domain.Product]] {
	coll, err := r.collection(ctx)
	if err != nil {
		return mo.Err[mo.Option[domain.Product]](err)
	}
	var out domain.Product
	err = coll.FindOne(ctx, bson.M{"sku": sku, "status": domain.StatusActive}).Decode(&out)
	if err != nil {
		if err == mongodriver.ErrNoDocuments {
			return mo.Ok(mo.None[domain.Product]())
		}
		return mo.Err[mo.Option[domain.Product]](err)
	}
	return mo.Ok(mo.Some(out))
}

func (r *MongoRepository) List(ctx context.Context, limit int64) mo.Result[[]domain.Product] {
	coll, err := r.collection(ctx)
	if err != nil {
		return mo.Err[[]domain.Product](err)
	}
	if limit <= 0 {
		limit = 50
	}
	cur, err := coll.Aggregate(ctx, currentStatePipeline(domain.StatusActive, limit))
	if err != nil {
		return mo.Err[[]domain.Product](err)
	}
	return decodeProducts(ctx, cur)
}

func (r *MongoRepository) ListInactive(ctx context.Context, limit int64) mo.Result[[]domain.Product] {
	coll, err := r.collection(ctx)
	if err != nil {
		return mo.Err[[]domain.Product](err)
	}
	if limit <= 0 {
		limit = 50
	}
	cur, err := coll.Aggregate(ctx, currentStatePipeline(domain.StatusDeactivated, limit))
	if err != nil {
		return mo.Err[[]domain.Product](err)
	}
	return decodeProducts(ctx, cur)
}

func (r *MongoRepository) Deactivate(ctx context.Context, id primitive.ObjectID, at time.Time) mo.Result[*domain.Product] {
	coll, err := r.collection(ctx)
	if err != nil {
		return mo.Err[*domain.Product](err)
	}
	var orig domain.Product
	if err := coll.FindOne(ctx, bson.M{"_id": id}).Decode(&orig); err != nil {
		if err == mongodriver.ErrNoDocuments {
			return mo.Err[*domain.Product](port.ErrNotFound)
		}
		return mo.Err[*domain.Product](err)
	}
	next := &domain.Product{
		ID:            primitive.NewObjectID(),
		SKU:           orig.SKU,
		Name:          orig.Name,
		Price:         orig.Price,
		Status:        domain.StatusDeactivated,
		OriginalID:    &orig.ID,
		CreatedAt:     orig.CreatedAt,
		DeactivatedAt: &at,
	}
	if _, err := coll.InsertOne(ctx, next); err != nil {
		return mo.Err[*domain.Product](err)
	}
	return mo.Ok(next)
}

func (r *MongoRepository) Activate(ctx context.Context, id primitive.ObjectID) mo.Result[*domain.Product] {
	coll, err := r.collection(ctx)
	if err != nil {
		return mo.Err[*domain.Product](err)
	}
	var orig domain.Product
	if err := coll.FindOne(ctx, bson.M{"_id": id}).Decode(&orig); err != nil {
		if err == mongodriver.ErrNoDocuments {
			return mo.Err[*domain.Product](port.ErrNotFound)
		}
		return mo.Err[*domain.Product](err)
	}
	next := &domain.Product{
		ID:         primitive.NewObjectID(),
		SKU:        orig.SKU,
		Name:       orig.Name,
		Price:      orig.Price,
		Status:     domain.StatusActive,
		OriginalID: &orig.ID,
		CreatedAt:  orig.CreatedAt,
	}
	if _, err := coll.InsertOne(ctx, next); err != nil {
		return mo.Err[*domain.Product](err)
	}
	return mo.Ok(next)
}

func decodeProducts(ctx context.Context, cur *mongodriver.Cursor) mo.Result[[]domain.Product] {
	defer cur.Close(ctx)
	var items []domain.Product
	for cur.Next(ctx) {
		var p domain.Product
		if err := cur.Decode(&p); err != nil {
			return mo.Err[[]domain.Product](err)
		}
		items = append(items, p)
	}
	if err := cur.Err(); err != nil {
		return mo.Err[[]domain.Product](err)
	}
	return mo.Ok(items)
}
