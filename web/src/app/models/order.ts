export interface OrderItem {
  sku: string;
  quantity: number;
  unit_price: number;
}

export interface Order {
  id: string;
  customer_id?: string | null;
  items: OrderItem[];
  total: number;
  created_at: string;
  deactivated_at?: string | null;
}

export interface CreateOrderRequest {
  customer_id?: string | null;
  items: OrderItem[];
}

export interface CatalogProduct {
  sku: string;
  name: string;
  price: number;
}
