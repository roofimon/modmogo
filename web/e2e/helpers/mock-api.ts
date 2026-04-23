import { Page, Route } from '@playwright/test';

// ── Fixtures ────────────────────────────────────────────────────────────────

export const PRODUCT = {
  id: '507f1f77bcf86cd799439011',
  sku: 'SKU-001',
  name: 'Widget Pro',
  price: 9.99,
  status: 'active',
  created_at: '2024-01-15T10:00:00Z',
  deactivated_at: null,
};

export const DEACTIVATED_PRODUCT = {
  ...PRODUCT,
  id: '507f1f77bcf86cd799439014',
  sku: 'SKU-002',
  name: 'Vintage Gadget',
  status: 'deactivated',
  deactivated_at: '2024-01-20T10:00:00Z',
};

export const CATALOG_PRODUCT = { sku: 'SKU-001', name: 'Widget Pro', price: 9.99 };

export const CUSTOMER = {
  id: '507f1f77bcf86cd799439012',
  name: 'Alice Smith',
  email: 'alice@example.com',
  phone: '+1 555 000 0001',
  status: 'active',
  created_at: '2024-01-15T10:00:00Z',
  deactivated_at: null,
};

export const CATALOG_CUSTOMER = { id: CUSTOMER.id, name: CUSTOMER.name, phone: CUSTOMER.phone };

export const ORDER = {
  id: '507f1f77bcf86cd799439013',
  customer_id: null,
  customer_name: null,
  items: [{ sku: 'SKU-001', product_name: 'Widget Pro', quantity: 2, unit_price: 9.99 }],
  total: 19.98,
  status: null,
  original_order_id: null,
  created_at: '2024-01-15T10:00:00Z',
  deactivated_at: null,
};

export const COMPLETED_ORDER = {
  ...ORDER,
  id: '507f1f77bcf86cd799439015',
  status: 'payment_completed',
  original_order_id: ORDER.id,
};

// ── Helpers ──────────────────────────────────────────────────────────────────

// Let browser navigation (document requests) pass through to the Angular dev
// server; only intercept fetch/XHR API calls.
function isApiCall(r: Route): boolean {
  return r.request().resourceType() !== 'document';
}

// ── Route mocks ──────────────────────────────────────────────────────────────
// Routes are registered most-specific first so the first match wins.
// Note: API requests go through the Angular dev server proxy to localhost:8080,
// but Playwright intercepts at the browser level, so we match the browser URL.

export async function mockProductApi(page: Page): Promise<void> {
  // Register most-general patterns first; Playwright uses LIFO so the last
  // registered handler wins — most-specific must be registered last.
  await page.route('**/products*', (r) => {
    if (r.request().resourceType() === 'document') return r.continue();
    if (r.request().method() === 'POST') return r.fulfill({ status: 201, json: PRODUCT });
    return r.fulfill({ json: [PRODUCT] });
  });
  await page.route('**/products/*', (r) => {
    if (r.request().resourceType() === 'document') return r.continue();
    return r.fulfill({ json: PRODUCT });
  });
  await page.route('**/products/*/activate', (r) => {
    if (r.request().resourceType() === 'document') return r.continue();
    return r.fulfill({ json: PRODUCT });
  });
  await page.route('**/products/*/deactivate', (r) => {
    if (r.request().resourceType() === 'document') return r.continue();
    return r.fulfill({ json: { ...PRODUCT, deactivated_at: '2024-01-20T10:00:00Z', status: 'deactivated' } });
  });
  await page.route('**/products/inactive*', (r) => {
    if (r.request().resourceType() === 'document') return r.continue();
    return r.fulfill({ json: [DEACTIVATED_PRODUCT] });
  });
}

export async function mockCustomerApi(page: Page): Promise<void> {
  // Register most-general patterns first; Playwright uses LIFO so the last
  // registered handler wins — most-specific must be registered last.
  await page.route('**/customers*', (r) => {
    if (r.request().resourceType() === 'document') return r.continue();
    if (r.request().method() === 'POST') return r.fulfill({ status: 201, json: CUSTOMER });
    return r.fulfill({ json: [CUSTOMER] });
  });
  await page.route('**/customers/*', (r) => {
    if (r.request().resourceType() === 'document') return r.continue();
    return r.fulfill({ json: CUSTOMER });
  });
  await page.route('**/customers/*/deactivate', (r) => {
    if (r.request().resourceType() === 'document') return r.continue();
    return r.fulfill({ json: { ...CUSTOMER, status: 'deactivated', deactivated_at: '2024-01-20T10:00:00Z' } });
  });
  await page.route('**/customers/inactive*', (r) => {
    if (r.request().resourceType() === 'document') return r.continue();
    return r.fulfill({ json: [{ ...CUSTOMER, status: 'deactivated', deactivated_at: '2024-01-20T10:00:00Z' }] });
  });
}

export async function mockOrderApi(page: Page): Promise<void> {
  // Register most-general patterns first; Playwright uses LIFO so the last
  // registered handler wins — most-specific must be registered last.
  // Use **/... globs (browser URL is localhost:4200, not localhost:8080).
  await page.route('**/orders*', (r) => {
    if (r.request().resourceType() === 'document') return r.continue();
    if (r.request().method() === 'POST') return r.fulfill({ status: 201, json: ORDER });
    return r.fulfill({ json: [ORDER] });
  });
  await page.route('**/orders/*', (r) => {
    if (r.request().resourceType() === 'document') return r.continue();
    return r.fulfill({ json: ORDER });
  });
  await page.route('**/orders/*/complete-payment', (r) => {
    if (r.request().resourceType() === 'document') return r.continue();
    return r.fulfill({ status: 201, json: COMPLETED_ORDER });
  });
  await page.route('**/orders/*/deactivate', (r) => {
    if (r.request().resourceType() === 'document') return r.continue();
    return r.fulfill({ json: { ...ORDER, deactivated_at: '2024-01-20T10:00:00Z' } });
  });
  await page.route('**/orders/inactive*', (r) => {
    if (r.request().resourceType() === 'document') return r.continue();
    return r.fulfill({ json: [{ ...ORDER, deactivated_at: '2024-01-20T10:00:00Z' }] });
  });
  await page.route('**/orders/payment-completed*', (r) => {
    if (r.request().resourceType() === 'document') return r.continue();
    return r.fulfill({ json: [COMPLETED_ORDER] });
  });
  await page.route('**/orders/products*', (r) => {
    if (r.request().resourceType() === 'document') return r.continue();
    return r.fulfill({ json: [CATALOG_PRODUCT] });
  });
  await page.route('**/orders/customers*', (r) => {
    if (r.request().resourceType() === 'document') return r.continue();
    return r.fulfill({ json: [CATALOG_CUSTOMER] });
  });
}
