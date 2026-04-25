package port

import (
	"context"
	"errors"
	"time"

	"github.com/samber/mo"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"modmono/domain/product/domain"
)

// ErrNotFound is returned when no product exists for the given id.
var ErrNotFound = errors.New("product: not found")

// Repository is the PortOut — the contract the application depends on for persistence.
type Repository interface {
	Create(ctx context.Context, p *domain.Product) mo.Result[*domain.Product]
	GetByID(ctx context.Context, id primitive.ObjectID) mo.Result[mo.Option[domain.Product]]
	GetBySKU(ctx context.Context, sku string) mo.Result[mo.Option[domain.Product]]
	List(ctx context.Context, limit int64) mo.Result[[]domain.Product]
	ListInactive(ctx context.Context, limit int64) mo.Result[[]domain.Product]
	Deactivate(ctx context.Context, id primitive.ObjectID, at time.Time) mo.Result[*domain.Product]
	Activate(ctx context.Context, id primitive.ObjectID) mo.Result[*domain.Product]
}

// UseCase is the PortIn — the contract consumed by transport adapters.
type UseCase interface {
	Create(ctx context.Context, in domain.CreateInput) mo.Result[*domain.Product]
	ViewProductDetail(ctx context.Context, id string) mo.Result[mo.Option[domain.Product]]
	FindProductBySKU(ctx context.Context, sku string) mo.Result[mo.Option[domain.Product]]
	List(ctx context.Context, limit int64) mo.Result[[]domain.Product]
	ListInactive(ctx context.Context, limit int64) mo.Result[[]domain.Product]
	Activate(ctx context.Context, id string) mo.Result[*domain.Product]
	Deactivate(ctx context.Context, id string) mo.Result[*domain.Product]
}
