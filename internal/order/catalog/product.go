package catalog

import (
	"context"

	"github.com/samber/mo"

	"modmono/internal/order"
	"modmono/internal/product"
)

// ProductCatalogAdapter adapts product.Service to the order.ProductCatalog port.
type ProductCatalogAdapter struct {
	svc *product.Service
}

func NewProductCatalogAdapter(svc *product.Service) *ProductCatalogAdapter {
	return &ProductCatalogAdapter{svc: svc}
}

func (a *ProductCatalogAdapter) ListActiveProducts(ctx context.Context, limit int64) mo.Result[[]order.CatalogProduct] {
	res := a.svc.List(ctx, limit)
	if res.IsError() {
		return mo.Err[[]order.CatalogProduct](res.Error())
	}
	products := res.MustGet()
	out := make([]order.CatalogProduct, len(products))
	for i, p := range products {
		out[i] = order.CatalogProduct{SKU: p.SKU, Name: p.Name, Price: p.Price}
	}
	return mo.Ok(out)
}

func (a *ProductCatalogAdapter) ResolveProductName(ctx context.Context, sku string) string {
	res := a.svc.GetBySKU(ctx, sku)
	if res.IsError() || res.MustGet().IsAbsent() {
		return ""
	}
	p, _ := res.MustGet().Get()
	return p.Name
}
