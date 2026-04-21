package catalog

import (
	"context"

	"github.com/samber/mo"

	customerapplication "modmono/internal/customer/application"
	"modmono/internal/order"
)

// CustomerCatalogAdapter adapts customer Service to the order.CustomerCatalog port.
type CustomerCatalogAdapter struct {
	svc *customerapplication.Service
}

func NewCustomerCatalogAdapter(svc *customerapplication.Service) *CustomerCatalogAdapter {
	return &CustomerCatalogAdapter{svc: svc}
}

func (a *CustomerCatalogAdapter) ListActiveCustomers(ctx context.Context, limit int64) mo.Result[[]order.CatalogCustomer] {
	res := a.svc.List(ctx, limit)
	if res.IsError() {
		return mo.Err[[]order.CatalogCustomer](res.Error())
	}
	customers := res.MustGet()
	out := make([]order.CatalogCustomer, len(customers))
	for i, c := range customers {
		out[i] = order.CatalogCustomer{ID: c.ID.Hex(), Name: c.Name, Phone: c.Phone}
	}
	return mo.Ok(out)
}

func (a *CustomerCatalogAdapter) ResolveCustomerName(ctx context.Context, hexID string) string {
	res := a.svc.ViewCustomerDetail(ctx, hexID)
	if res.IsError() || res.MustGet().IsAbsent() {
		return ""
	}
	c, _ := res.MustGet().Get()
	return c.Name
}
