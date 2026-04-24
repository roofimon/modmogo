import { Product, CreateProductRequest } from './types';

const API_BASE = process.env.NEXT_PUBLIC_API_URL || '/api';

export async function listProducts(limit = 50): Promise<Product[]> {
  const res = await fetch(`${API_BASE}/products?limit=${limit}`, { cache: 'no-store' });
  if (!res.ok) throw new Error('Failed to fetch products');
  return res.json();
}

export async function listInactiveProducts(limit = 50): Promise<Product[]> {
  const res = await fetch(`${API_BASE}/products/inactive?limit=${limit}`, { cache: 'no-store' });
  if (!res.ok) throw new Error('Failed to fetch products');
  return res.json();
}

export async function getProductById(id: string): Promise<Product> {
  const res = await fetch(`${API_BASE}/products/${encodeURIComponent(id)}`, { cache: 'no-store' });
  if (!res.ok) throw new Error('Failed to fetch product');
  return res.json();
}

export async function createProduct(data: CreateProductRequest): Promise<Product> {
  const res = await fetch(`${API_BASE}/products`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  });
  if (!res.ok) throw new Error('Failed to create product');
  return res.json();
}

export async function deactivateProduct(id: string): Promise<Product> {
  const res = await fetch(`${API_BASE}/products/${encodeURIComponent(id)}/deactivate`, {
    method: 'POST',
  });
  if (!res.ok) throw new Error('Failed to deactivate product');
  return res.json();
}

export async function activateProduct(id: string): Promise<Product> {
  const res = await fetch(`${API_BASE}/products/${encodeURIComponent(id)}/activate`, {
    method: 'POST',
  });
  if (!res.ok) throw new Error('Failed to activate product');
  return res.json();
}
