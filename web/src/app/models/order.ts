export interface OrderItem {
  sku: string;
  product_name?: string | null;
  quantity: number;
  unit_price: number;
}

export interface Order {
  id: string;
  customer_id?: string | null;
  customer_name?: string | null;
  items: OrderItem[];
  total: number;
  status?: string | null;
  original_order_id?: string | null;
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

export interface CatalogCustomer {
  id: string;
  name: string;
  phone?: string | null;
}
