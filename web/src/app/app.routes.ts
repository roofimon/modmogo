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
  {
    path: 'orders',
    loadComponent: () =>
      import('./pages/order-list/order-list.component').then((m) => m.OrderListComponent),
  },
  {
    path: 'orders/inactive',
    loadComponent: () =>
      import('./pages/order-inactive-list/order-inactive-list.component').then(
        (m) => m.OrderInactiveListComponent,
      ),
  },
  {
    path: 'orders/payment-completed',
    loadComponent: () =>
      import('./pages/order-payment-completed-list/order-payment-completed-list.component').then(
        (m) => m.OrderPaymentCompletedListComponent,
      ),
  },
  { path: 'orders/new', redirectTo: 'orders' },
  {
    path: 'orders/:id',
    loadComponent: () =>
      import('./pages/order-detail/order-detail.component').then((m) => m.OrderDetailComponent),
  },
  { path: '**', redirectTo: 'products' },
];
