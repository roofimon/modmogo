import { Customer, CreateCustomerRequest, CatalogCustomer } from './types';

const API_BASE = process.env.NEXT_PUBLIC_API_URL || '/api';

export async function listCustomers(limit = 50): Promise<Customer[]> {
  const res = await fetch(`${API_BASE}/customers?limit=${limit}`, { cache: 'no-store' });
  if (!res.ok) throw new Error('Failed to fetch customers');
  return res.json();
}

export async function listInactiveCustomers(limit = 50): Promise<Customer[]> {
  const res = await fetch(`${API_BASE}/customers/inactive?limit=${limit}`, { cache: 'no-store' });
  if (!res.ok) throw new Error('Failed to fetch customers');
  return res.json();
}

export async function getCustomerById(id: string): Promise<Customer> {
  const res = await fetch(`${API_BASE}/customers/${encodeURIComponent(id)}`, { cache: 'no-store' });
  if (!res.ok) throw new Error('Failed to fetch customer');
  return res.json();
}

export async function createCustomer(data: CreateCustomerRequest): Promise<Customer> {
  const res = await fetch(`${API_BASE}/customers`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  });
  if (!res.ok) throw new Error('Failed to create customer');
  return res.json();
}

export async function deactivateCustomer(id: string): Promise<Customer> {
  const res = await fetch(`${API_BASE}/customers/${encodeURIComponent(id)}/deactivate`, {
    method: 'POST',
  });
  if (!res.ok) throw new Error('Failed to deactivate customer');
  return res.json();
}

export async function searchCustomers(query: string): Promise<CatalogCustomer[]> {
  const res = await fetch(`${API_BASE}/customers?query=${encodeURIComponent(query)}&limit=10`, { cache: 'no-store' });
  if (!res.ok) throw new Error('Failed to search customers');
  return res.json();
}
