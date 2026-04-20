export interface Product {
  id: string;
  sku: string;
  name: string;
  price: number;
  created_at: string;
  /** ISO-8601 when soft-deactivated; absent or null when active. */
  deactivated_at?: string | null;
}

export interface CreateProductRequest {
  sku: string;
  name: string;
  price: number;
}
