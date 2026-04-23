import { expect, test } from '@playwright/test';

import { CUSTOMER, mockCustomerApi } from './helpers/mock-api';

test.describe('Customers — list', () => {
  test('shows active customers', async ({ page }) => {
    await mockCustomerApi(page);
    await page.goto('/customers');
    await expect(page.getByText('Alice Smith')).toBeVisible();
    await expect(page.getByText('alice@example.com')).toBeVisible();
  });

  test('navigates to detail on card click', async ({ page }) => {
    await mockCustomerApi(page);
    await page.goto('/customers');
    await page.locator('.card').first().click();
    await expect(page).toHaveURL(`/customers/${CUSTOMER.id}`);
  });
});

test.describe('Customers — create modal', () => {
  test('opens modal on Add customer click', async ({ page }) => {
    await mockCustomerApi(page);
    await page.goto('/customers');
    await page.getByRole('button', { name: 'Add customer' }).click();
    await expect(page.getByRole('dialog')).toBeVisible();
  });

  test('shows validation errors on empty submit', async ({ page }) => {
    await mockCustomerApi(page);
    await page.goto('/customers');
    await page.getByRole('button', { name: 'Add customer' }).click();
    await page.getByRole('button', { name: 'Create customer' }).click();
    await expect(page.getByText('Name is required.')).toBeVisible();
    await expect(page.getByText('Enter a valid email.')).toBeVisible();
  });

  test('shows email validation error for invalid email', async ({ page }) => {
    await mockCustomerApi(page);
    await page.goto('/customers');
    await page.getByRole('button', { name: 'Add customer' }).click();
    await page.getByLabel('Name').fill('Alice Smith');
    await page.getByLabel('Email').fill('not-an-email');
    await page.getByRole('button', { name: 'Create customer' }).click();
    await expect(page.getByText('Enter a valid email.')).toBeVisible();
  });

  test('submits valid form with phone and navigates to detail', async ({ page }) => {
    await mockCustomerApi(page);
    await page.goto('/customers');
    await page.getByRole('button', { name: 'Add customer' }).click();
    await page.getByLabel('Name').fill('Alice Smith');
    await page.getByLabel('Email').fill('alice@example.com');
    await page.getByLabel('Mobile Phone').fill('+1 555 000 0001');
    await page.getByRole('button', { name: 'Create customer' }).click();
    await expect(page).toHaveURL(`/customers/${CUSTOMER.id}`);
  });

  test('submits without optional phone', async ({ page }) => {
    await mockCustomerApi(page);
    await page.goto('/customers');
    await page.getByRole('button', { name: 'Add customer' }).click();
    await page.getByLabel('Name').fill('Alice Smith');
    await page.getByLabel('Email').fill('alice@example.com');
    await page.getByRole('button', { name: 'Create customer' }).click();
    await expect(page).toHaveURL(`/customers/${CUSTOMER.id}`);
  });
});

test.describe('Customers — detail', () => {
  test('shows customer info', async ({ page }) => {
    await mockCustomerApi(page);
    await page.goto(`/customers/${CUSTOMER.id}`);
    await expect(page.getByText('Alice Smith')).toBeVisible();
    await expect(page.getByText('alice@example.com')).toBeVisible();
    await expect(page.getByText('+1 555 000 0001')).toBeVisible();
  });

  test('deactivates customer on confirm', async ({ page }) => {
    await mockCustomerApi(page);
    await page.goto(`/customers/${CUSTOMER.id}`);
    page.on('dialog', (d) => d.accept());
    await page.getByRole('button', { name: 'Deactivate customer' }).click();
    await expect(page.locator('.banner-inactive')).toBeVisible();
  });

  test('cancels deactivate on dismiss', async ({ page }) => {
    await mockCustomerApi(page);
    await page.goto(`/customers/${CUSTOMER.id}`);
    page.on('dialog', (d) => d.dismiss());
    await page.getByRole('button', { name: 'Deactivate customer' }).click();
    await expect(page.locator('.banner-inactive')).not.toBeVisible();
  });
});

test.describe('Customers — inactive list', () => {
  test('shows inactive customers', async ({ page }) => {
    await mockCustomerApi(page);
    await page.goto('/customers/inactive');
    await expect(page.getByText('Alice Smith')).toBeVisible();
    await expect(page.locator('.card-grid .badge').first()).toHaveText('Inactive');
  });
});
