import { expect, test } from '@playwright/test';

import { DEACTIVATED_PRODUCT, PRODUCT, mockProductApi } from './helpers/mock-api';

test.describe('Products — list', () => {
  test('shows active products', async ({ page }) => {
    await mockProductApi(page);
    await page.goto('/products');
    await expect(page.getByText('Widget Pro')).toBeVisible();
    await expect(page.getByText('SKU-001')).toBeVisible();
    await expect(page.locator('.card-price').first()).toContainText('9.99');
  });

  test('navigates to detail on card click', async ({ page }) => {
    await mockProductApi(page);
    await page.goto('/products');
    await expect(page.locator('.card').first()).toBeVisible();
    await page.locator('.card').first().click();
    await expect(page).toHaveURL(`/products/${PRODUCT.id}`);
  });
});

test.describe('Products — create modal', () => {
  test('opens modal on Add product click', async ({ page }) => {
    await mockProductApi(page);
    await page.goto('/products');
    await page.getByRole('button', { name: 'Add product' }).click();
    await expect(page.getByRole('dialog')).toBeVisible();
    await expect(page.getByRole('heading', { name: 'Add product' })).toBeVisible();
  });

  test('shows validation errors on empty submit', async ({ page }) => {
    await mockProductApi(page);
    await page.goto('/products');
    await page.getByRole('button', { name: 'Add product' }).click();
    await page.getByRole('button', { name: 'Create product' }).click();
    await expect(page.getByText('SKU is required.')).toBeVisible();
    await expect(page.getByText('Name is required.')).toBeVisible();
  });

  test('submits and navigates to detail', async ({ page }) => {
    await mockProductApi(page);
    await page.goto('/products');
    await page.getByRole('button', { name: 'Add product' }).click();
    await page.getByLabel('SKU').fill('SKU-001');
    await page.getByLabel('Name').fill('Widget Pro');
    await page.getByLabel('Price (USD)').fill('9.99');
    await page.getByRole('button', { name: 'Create product' }).click();
    await expect(page).toHaveURL(`/products/${PRODUCT.id}`);
  });

  test('closes on cancel', async ({ page }) => {
    await mockProductApi(page);
    await page.goto('/products');
    await page.getByRole('button', { name: 'Add product' }).click();
    await page.getByRole('button', { name: 'Cancel' }).click();
    await expect(page.getByRole('dialog')).not.toBeVisible();
  });
});

test.describe('Products — detail', () => {
  test('shows product info', async ({ page }) => {
    await mockProductApi(page);
    await page.goto(`/products/${PRODUCT.id}`);
    await expect(page.getByText('Widget Pro')).toBeVisible();
    await expect(page.getByText('SKU-001')).toBeVisible();
    await expect(page.locator('dd').filter({ hasText: '9.99' })).toBeVisible();
  });

  test('deactivates product on confirm', async ({ page }) => {
    await mockProductApi(page);
    await page.goto(`/products/${PRODUCT.id}`);
    page.on('dialog', (d) => d.accept());
    await page.getByRole('button', { name: 'Deactivate product' }).click();
    await expect(page.locator('.banner-inactive')).toBeVisible();
    await expect(page.getByText('inactive')).toBeVisible();
  });

  test('cancels deactivate on dismiss', async ({ page }) => {
    await mockProductApi(page);
    await page.goto(`/products/${PRODUCT.id}`);
    page.on('dialog', (d) => d.dismiss());
    await page.getByRole('button', { name: 'Deactivate product' }).click();
    await expect(page.locator('.banner-inactive')).not.toBeVisible();
  });
});

test.describe('Products — inactive list', () => {
  test('shows inactive products', async ({ page }) => {
    await mockProductApi(page);
    await page.goto('/products/inactive');
    await expect(page.getByText('Vintage Gadget')).toBeVisible();
    await expect(page.locator('.card-grid .badge').first()).toHaveText('Inactive');
  });

  test('activates product and removes from list', async ({ page }) => {
    await mockProductApi(page);
    await page.goto('/products/inactive');
    await expect(page.getByText(DEACTIVATED_PRODUCT.name)).toBeVisible();
    await page.getByRole('button', { name: 'Activate' }).click();
    await expect(page.getByText(DEACTIVATED_PRODUCT.name)).not.toBeVisible();
  });
});
