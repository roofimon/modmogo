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
import { createProduct } from '@/lib/api/products';
import type { CreateProductRequest } from '@/lib/api/types';
import { createProductSchema, type CreateProductFormData } from '@/lib/validations/product';

interface ProductCreateModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function ProductCreateModal({ open, onOpenChange }: ProductCreateModalProps) {
  const queryClient = useQueryClient();
  const router = useRouter();
  const [error, setError] = useState<string | null>(null);

  const form = useForm<CreateProductFormData>({
    resolver: zodResolver(createProductSchema),
    defaultValues: { sku: '', name: '', price: 0 },
    mode: 'onChange',
  });

  const mutation = useMutation({
    mutationFn: (data: CreateProductRequest) =>
      createProduct({ sku: data.sku, name: data.name, price: Number(data.price) }),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['products'] });
      queryClient.invalidateQueries({ queryKey: ['inactiveProducts'] });
      onOpenChange(false);
      form.reset();
      router.push(`/products/${variables.sku}`);
    },
    onError: (err) => setError(err.message),
  });

  const onSubmit = (data: CreateProductFormData) => {
    setError(null);
    mutation.mutate(data);
  };

  return (
    <Modal open={open} onOpenChange={onOpenChange} title="Add product">
      <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
        <div>
          <Label htmlFor="sku">SKU</Label>
          <Input
            id="sku"
            {...form.register('sku')}
            error={!!form.formState.errors.sku}
            placeholder="Enter SKU"
          />
          {form.formState.errors.sku && (
            <p className="text-sm text-[#cf202f] mt-1">{form.formState.errors.sku.message}</p>
          )}
        </div>
        <div>
          <Label htmlFor="name">Name</Label>
          <Input
            id="name"
            {...form.register('name')}
            error={!!form.formState.errors.name}
            placeholder="Enter product name"
          />
          {form.formState.errors.name && (
            <p className="text-sm text-[#cf202f] mt-1">{form.formState.errors.name.message}</p>
          )}
        </div>
        <div>
          <Label htmlFor="price">Price (USD)</Label>
          <Input
            id="price"
            type="number"
            step="0.01"
            min="0"
            {...form.register('price', { valueAsNumber: true })}
            error={!!form.formState.errors.price}
            placeholder="0.00"
          />
          {form.formState.errors.price && (
            <p className="text-sm text-[#cf202f] mt-1">{form.formState.errors.price.message}</p>
          )}
        </div>
        {error && <p className="text-sm text-[#cf202f]">{error}</p>}
        <div className="flex justify-end gap-3 pt-4">
          <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button type="submit" disabled={mutation.isPending}>
            {mutation.isPending ? 'Creating...' : 'Create product'}
          </Button>
        </div>
      </form>
    </Modal>
  );
}
