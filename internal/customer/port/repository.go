package port

import (
	"context"
	"errors"
	"time"

	"github.com/samber/mo"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"modmono/internal/customer/domain"
)

// ErrNotFound is returned when no customer exists for the given id.
var ErrNotFound = errors.New("customer: not found")

// Repository is the Port — the contract the application depends on.
type Repository interface {
	Create(ctx context.Context, c *domain.Customer) mo.Result[*domain.Customer]
	GetByID(ctx context.Context, id primitive.ObjectID) mo.Result[mo.Option[domain.Customer]]
	List(ctx context.Context, limit int64) mo.Result[[]domain.Customer]
	ListInactive(ctx context.Context, limit int64) mo.Result[[]domain.Customer]
	Deactivate(ctx context.Context, id primitive.ObjectID, at time.Time) mo.Result[*domain.Customer]
}

// UseCase is the inbound port consumed by transport adapters.
type UseCase interface {
	Create(ctx context.Context, in domain.CreateInput) mo.Result[*domain.Customer]
	ViewCustomerDetail(ctx context.Context, id string) mo.Result[mo.Option[domain.Customer]]
	List(ctx context.Context, limit int64) mo.Result[[]domain.Customer]
	ListInactive(ctx context.Context, limit int64) mo.Result[[]domain.Customer]
	Deactivate(ctx context.Context, id string) mo.Result[*domain.Customer]
}
