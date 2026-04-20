import { Routes } from '@angular/router';

export const routes: Routes = [
  { path: '', pathMatch: 'full', redirectTo: 'products' },
  {
    path: 'products',
    loadComponent: () =>
      import('./pages/product-list/product-list.component').then((m) => m.ProductListComponent),
  },
  {
    path: 'products/inactive',
    loadComponent: () =>
      import('./pages/product-inactive-list/product-inactive-list.component').then(
        (m) => m.ProductInactiveListComponent,
      ),
  },
  { path: 'products/new', redirectTo: 'products' },
  {
    path: 'products/:id',
    loadComponent: () =>
      import('./pages/product-detail/product-detail.component').then((m) => m.ProductDetailComponent),
  },
  {
    path: 'customers',
    loadComponent: () =>
      import('./pages/customer-list/customer-list.component').then((m) => m.CustomerListComponent),
  },
  {
    path: 'customers/inactive',
    loadComponent: () =>
      import('./pages/customer-inactive-list/customer-inactive-list.component').then(
        (m) => m.CustomerInactiveListComponent,
      ),
  },
  { path: 'customers/new', redirectTo: 'customers' },
  {
    path: 'customers/:id',
    loadComponent: () =>
      import('./pages/customer-detail/customer-detail.component').then((m) => m.CustomerDetailComponent),
  },
  { path: '**', redirectTo: 'products' },
];
