import { Order, CreateOrderRequest, CatalogProduct, CatalogCustomer } from './types';

const API_BASE = process.env.NEXT_PUBLIC_API_URL || '/api';

export async function listOrders(limit = 50): Promise<Order[]> {
  const res = await fetch(`${API_BASE}/orders?limit=${limit}`, { cache: 'no-store' });
  if (!res.ok) throw new Error('Failed to fetch orders');
  return res.json();
}

export async function listInactiveOrders(limit = 50): Promise<Order[]> {
  const res = await fetch(`${API_BASE}/orders/inactive?limit=${limit}`, { cache: 'no-store' });
  if (!res.ok) throw new Error('Failed to fetch orders');
  return res.json();
}

export async function listPaymentCompletedOrders(limit = 50): Promise<Order[]> {
  const res = await fetch(`${API_BASE}/orders/payment-completed?limit=${limit}`, { cache: 'no-store' });
  if (!res.ok) throw new Error('Failed to fetch orders');
  return res.json();
}

export async function getOrderById(id: string): Promise<Order> {
  const res = await fetch(`${API_BASE}/orders/${encodeURIComponent(id)}`, { cache: 'no-store' });
  if (!res.ok) throw new Error('Failed to fetch order');
  return res.json();
}

export async function createOrder(data: CreateOrderRequest): Promise<Order> {
  const res = await fetch(`${API_BASE}/orders`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  });
  if (!res.ok) throw new Error('Failed to create order');
  return res.json();
}

export async function completePayment(id: string): Promise<Order> {
  const res = await fetch(`${API_BASE}/orders/${encodeURIComponent(id)}/complete-payment`, {
    method: 'POST',
  });
  if (!res.ok) throw new Error('Failed to complete payment');
  return res.json();
}

export async function deactivateOrder(id: string): Promise<Order> {
  const res = await fetch(`${API_BASE}/orders/${encodeURIComponent(id)}/deactivate`, {
    method: 'POST',
  });
  if (!res.ok) throw new Error('Failed to deactivate order');
  return res.json();
}

export async function searchProducts(query: string): Promise<CatalogProduct[]> {
  const res = await fetch(`${API_BASE}/orders/products?query=${encodeURIComponent(query)}`, { cache: 'no-store' });
  if (!res.ok) throw new Error('Failed to search products');
  return res.json();
}

export async function searchCustomers(query: string): Promise<CatalogCustomer[]> {
  const res = await fetch(`${API_BASE}/orders/customers?query=${encodeURIComponent(query)}`, { cache: 'no-store' });
  if (!res.ok) throw new Error('Failed to search customers');
  return res.json();
}
