package catalog

import (
	"context"

	"github.com/samber/mo"

	"modmono/domain/order/port"
	productapplication "modmono/domain/product/application"
)

// ProductCatalogAdapter adapts product Service to the order port.ProductCatalog.
type ProductCatalogAdapter struct {
	svc *productapplication.Service
}

func NewProductCatalogAdapter(svc *productapplication.Service) *ProductCatalogAdapter {
	return &ProductCatalogAdapter{svc: svc}
}

func (a *ProductCatalogAdapter) ListActiveProducts(ctx context.Context, limit int64) mo.Result[[]port.CatalogProduct] {
	res := a.svc.List(ctx, limit)
	if res.IsError() {
		return mo.Err[[]port.CatalogProduct](res.Error())
	}
	products := res.MustGet()
	out := make([]port.CatalogProduct, len(products))
	for i, p := range products {
		out[i] = port.CatalogProduct{SKU: p.SKU, Name: p.Name, Price: p.Price}
	}
	return mo.Ok(out)
}

func (a *ProductCatalogAdapter) ResolveProductName(ctx context.Context, sku string) string {
	res := a.svc.FindProductBySKU(ctx, sku)
	if res.IsError() || res.MustGet().IsAbsent() {
		return ""
	}
	p, _ := res.MustGet().Get()
	return p.Name
}
