import { expect, test } from '@playwright/test';

import { CATALOG_CUSTOMER, COMPLETED_ORDER, ORDER, mockCustomerApi, mockOrderApi, mockProductApi } from './helpers/mock-api';

test.describe('Orders — list', () => {
  test('shows active orders', async ({ page }) => {
    await mockOrderApi(page);
    await page.goto('/orders');
    await expect(page.getByText(ORDER.id)).toBeVisible();
    await expect(page.locator('.card-total').first()).toContainText('19.98');
  });

  test('navigates to detail on card click', async ({ page }) => {
    await mockOrderApi(page);
    await page.goto('/orders');
    await expect(page.locator('.card').first()).toBeVisible();
    await page.locator('.card').first().click();
    await expect(page).toHaveURL(`/orders/${ORDER.id}`);
  });
});

test.describe('Orders — create modal', () => {
  test('opens modal on Add order click', async ({ page }) => {
    await mockOrderApi(page);
    await page.goto('/orders');
    await page.getByRole('button', { name: 'Add order' }).click();
    await expect(page.getByRole('dialog')).toBeVisible();
  });

  test('shows validation error on empty submit', async ({ page }) => {
    await mockOrderApi(page);
    await page.goto('/orders');
    await page.getByRole('button', { name: 'Add order' }).click();
    await page.getByRole('button', { name: 'Create order' }).click();
    await expect(page.getByText('Required.')).toBeVisible();
  });

  test('creates order with SKU autocomplete', async ({ page }) => {
    await mockOrderApi(page);
    await page.goto('/orders');
    await page.getByRole('button', { name: 'Add order' }).click();

    // Type SKU prefix to trigger autocomplete
    await page.locator('input[formControlName="sku"]').fill('SKU');
    await page.locator('.sku-dropdown li').first().click();

    // Verify price was auto-filled
    await expect(page.locator('input[formControlName="unit_price"]').first()).toHaveValue('9.99');

    await page.locator('input[formControlName="quantity"]').fill('2');
    await page.getByRole('button', { name: 'Create order' }).click();
    await expect(page).toHaveURL(`/orders/${ORDER.id}`);
  });

  test('creates order with customer autocomplete', async ({ page }) => {
    await mockOrderApi(page);
    await page.goto('/orders');
    await page.getByRole('button', { name: 'Add order' }).click();

    // Search for customer by name
    await page.locator('input[placeholder="Search by name…"]').fill('Alice');
    await page.locator('.customer-dropdown li').first().click();

    // Verify customer name appears in the search box
    await expect(page.locator('input[placeholder="Search by name…"]')).toHaveValue(CATALOG_CUSTOMER.name);

    // Fill line item
    await page.locator('input[formControlName="sku"]').fill('SKU-001');
    await page.locator('.sku-dropdown li').first().click();
    await page.getByRole('button', { name: 'Create order' }).click();
    await expect(page).toHaveURL(`/orders/${ORDER.id}`);
  });

  test('adds and removes line items', async ({ page }) => {
    await mockOrderApi(page);
    await page.goto('/orders');
    await page.getByRole('button', { name: 'Add order' }).click();

    // Add a second item row
    await page.getByRole('button', { name: '+ Add item' }).click();
    await expect(page.locator('.item-row')).toHaveCount(2);

    // Remove it
    await page.locator('.btn-remove-item').first().click();
    await expect(page.locator('.item-row')).toHaveCount(1);
  });
});

test.describe('Orders — detail', () => {
  test('shows order info with line items', async ({ page }) => {
    await mockOrderApi(page);
    await page.goto(`/orders/${ORDER.id}`);
    await expect(page.getByText(ORDER.id)).toBeVisible();
    await expect(page.getByText('Widget Pro')).toBeVisible();
    await expect(page.locator('.total-value')).toContainText('19.98');
  });

  test('completes payment and redirects to new order', async ({ page }) => {
    await mockOrderApi(page);
    await page.goto(`/orders/${ORDER.id}`);
    await page.getByRole('button', { name: 'Complete Payment' }).click();
    await expect(page).toHaveURL(`/orders/${COMPLETED_ORDER.id}`);
  });

  test('deactivates order on confirm', async ({ page }) => {
    await mockOrderApi(page);
    await page.goto(`/orders/${ORDER.id}`);
    page.on('dialog', (d) => d.accept());
    await page.getByRole('button', { name: 'Deactivate order' }).click();
    await expect(page.locator('.banner-inactive')).toBeVisible();
  });

  test('cancels deactivate on dismiss', async ({ page }) => {
    await mockOrderApi(page);
    await page.goto(`/orders/${ORDER.id}`);
    page.on('dialog', (d) => d.dismiss());
    await page.getByRole('button', { name: 'Deactivate order' }).click();
    await expect(page.locator('.banner-inactive')).not.toBeVisible();
  });
});

test.describe('Orders — inactive list', () => {
  test('shows inactive orders', async ({ page }) => {
    await mockOrderApi(page);
    await page.goto('/orders/inactive');
    await expect(page.locator('.card.card-inactive').first()).toBeVisible();
    await expect(page.locator('.card-total').first()).toContainText('19.98');
  });
});

test.describe('Orders — payment-completed list', () => {
  test('shows completed orders', async ({ page }) => {
    await mockOrderApi(page);
    await page.goto('/orders/payment-completed');
    await expect(page.locator('.card.card-payment-completed').first()).toBeVisible();
    await expect(page.getByText(`Original: ${ORDER.id}`)).toBeVisible();
  });
});

test.describe('Navigation', () => {
  test('top nav links to all domains', async ({ page }) => {
    await mockProductApi(page);
    await mockCustomerApi(page);
    await mockOrderApi(page);
    await page.goto('/products');

    await page.getByRole('link', { name: 'Customers' }).click();
    await expect(page).toHaveURL('/customers');

    await page.getByRole('link', { name: 'Orders' }).click();
    await expect(page).toHaveURL('/orders');

    await page.getByRole('link', { name: 'Products' }).click();
    await expect(page).toHaveURL('/products');
  });
});
