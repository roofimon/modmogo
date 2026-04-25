import { z } from 'zod';

export const orderItemSchema = z.object({
  sku: z.string().min(1, 'SKU is required.'),
  quantity: z.number().min(1, 'Quantity must be at least 1.'),
  unit_price: z.number().min(0, 'Price must be non-negative.'),
});

export const createOrderSchema = z.object({
  customer_id: z.string().optional().or(z.literal(null)).or(z.literal('')),
  items: z.array(orderItemSchema).min(1, 'At least one item is required.'),
});

export type CreateOrderFormData = z.infer<typeof createOrderSchema>;
export type OrderItemFormData = z.infer<typeof orderItemSchema>;
