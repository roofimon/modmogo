package product

import (
	"context"

	"github.com/samber/mo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	mongodriver "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Repository persists products.
type Repository interface {
	Create(ctx context.Context, p *Product) mo.Result[*Product]
	GetByID(ctx context.Context, id primitive.ObjectID) (mo.Option[Product], error)
	List(ctx context.Context, limit int64) mo.Result[[]Product]
}

// MongoRepository implements Repository using a MongoDB collection.
type MongoRepository struct {
	coll *mongodriver.Collection
}

// NewMongoRepository builds a Mongo-backed repository.
func NewMongoRepository(db *mongodriver.Database) *MongoRepository {
	return &MongoRepository{coll: db.Collection("products")}
}

// Create inserts a product and returns it with server-assigned fields.
func (r *MongoRepository) Create(ctx context.Context, p *Product) mo.Result[*Product] {
	if p.ID.IsZero() {
		p.ID = primitive.NewObjectID()
	}
	if p.CreatedAt.IsZero() {
		p.CreatedAt = p.ID.Timestamp()
	}
	_, err := r.coll.InsertOne(ctx, p)
	if err != nil {
		return mo.Err[*Product](err)
	}
	return mo.Ok(p)
}

// GetByID returns Some(product) if found, None if missing, or an error for infrastructure failures.
func (r *MongoRepository) GetByID(ctx context.Context, id primitive.ObjectID) (mo.Option[Product], error) {
	var out Product
	err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&out)
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
	if limit <= 0 {
		limit = 50
	}
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).SetLimit(limit)
	cur, err := r.coll.Find(ctx, bson.M{}, opts)
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
