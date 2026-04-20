export interface Product {
  id: string;
  sku: string;
  name: string;
  price: number;
  created_at: string;
}

export interface CreateProductRequest {
  sku: string;
  name: string;
  price: number;
}
