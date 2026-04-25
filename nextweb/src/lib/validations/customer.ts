import { z } from 'zod';

export const createCustomerSchema = z.object({
  name: z.string().min(1, 'Name is required.').max(255),
  email: z.string().email('Invalid email address.'),
  phone: z.string().max(50).optional().or(z.literal('')),
});

export type CreateCustomerFormData = z.infer<typeof createCustomerSchema>;
