'use client';

import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useRouter } from 'next/navigation';
import { Modal } from '@/components/Modal';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Button } from '@/components/ui/button';
import { createCustomer } from '@/lib/api/customers';
import type { CreateCustomerRequest } from '@/lib/api/types';
import { createCustomerSchema, type CreateCustomerFormData } from '@/lib/validations/customer';

interface CustomerCreateModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function CustomerCreateModal({ open, onOpenChange }: CustomerCreateModalProps) {
  const queryClient = useQueryClient();
  const router = useRouter();
  const [error, setError] = useState<string | null>(null);

  const form = useForm<CreateCustomerFormData>({
    resolver: zodResolver(createCustomerSchema),
    defaultValues: { name: '', email: '', phone: '' },
  });

  const mutation = useMutation({
    mutationFn: (data: CreateCustomerRequest) => createCustomer(data),
    onSuccess: (customer) => {
      queryClient.invalidateQueries({ queryKey: ['customers'] });
      queryClient.invalidateQueries({ queryKey: ['inactiveCustomers'] });
      onOpenChange(false);
      form.reset();
      router.push(`/customers/${customer.id}`);
    },
    onError: (err) => setError(err.message),
  });

  const onSubmit = (data: CreateCustomerFormData) => {
    setError(null);
    const payload: CreateCustomerRequest = {
      name: data.name,
      email: data.email,
      phone: data.phone || undefined,
    };
    mutation.mutate(payload);
  };

  return (
    <Modal open={open} onOpenChange={onOpenChange} title="Add customer">
      <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
        <div>
          <Label htmlFor="name">Name</Label>
          <Input
            id="name"
            {...form.register('name')}
            error={!!form.formState.errors.name}
            placeholder="Enter customer name"
          />
          {form.formState.errors.name && (
            <p className="text-sm text-[#cf202f] mt-1">{form.formState.errors.name.message}</p>
          )}
        </div>
        <div>
          <Label htmlFor="email">Email</Label>
          <Input
            id="email"
            type="email"
            {...form.register('email')}
            error={!!form.formState.errors.email}
            placeholder="customer@example.com"
          />
          {form.formState.errors.email && (
            <p className="text-sm text-[#cf202f] mt-1">{form.formState.errors.email.message}</p>
          )}
        </div>
        <div>
          <Label htmlFor="phone">Phone (optional)</Label>
          <Input
            id="phone"
            {...form.register('phone')}
            error={!!form.formState.errors.phone}
            placeholder="+1 555 000 0000"
          />
          {form.formState.errors.phone && (
            <p className="text-sm text-[#cf202f] mt-1">{form.formState.errors.phone.message}</p>
          )}
        </div>
        {error && <p className="text-sm text-[#cf202f]">{error}</p>}
        <div className="flex justify-end gap-3 pt-4">
          <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button type="submit" disabled={mutation.isPending}>
            {mutation.isPending ? 'Creating...' : 'Create customer'}
          </Button>
        </div>
      </form>
    </Modal>
  );
}
