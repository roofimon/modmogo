package catalog

import (
	"context"

	"github.com/samber/mo"

	customerapplication "modmono/internal/customer/application"
	"modmono/internal/order/port"
)

// CustomerCatalogAdapter adapts customer Service to the order port.CustomerCatalog.
type CustomerCatalogAdapter struct {
	svc *customerapplication.Service
}

func NewCustomerCatalogAdapter(svc *customerapplication.Service) *CustomerCatalogAdapter {
	return &CustomerCatalogAdapter{svc: svc}
}

func (a *CustomerCatalogAdapter) ListActiveCustomers(ctx context.Context, limit int64) mo.Result[[]port.CatalogCustomer] {
	res := a.svc.List(ctx, limit)
	if res.IsError() {
		return mo.Err[[]port.CatalogCustomer](res.Error())
	}
	customers := res.MustGet()
	out := make([]port.CatalogCustomer, len(customers))
	for i, c := range customers {
		out[i] = port.CatalogCustomer{ID: c.ID.Hex(), Name: c.Name, Phone: c.Phone}
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
