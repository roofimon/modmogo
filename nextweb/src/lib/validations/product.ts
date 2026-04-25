import { z } from 'zod';

export const createProductSchema = z.object({
  sku: z.string().min(1, 'SKU is required.').max(64),
  name: z.string().min(1, 'Name is required.').max(500),
  price: z.number().min(0, 'Price must be non-negative.'),
});

export type CreateProductFormData = z.infer<typeof createProductSchema>;
