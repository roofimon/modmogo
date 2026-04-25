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
import { OrderItemsTable } from '@/components/orders/OrderItemsTable';
import { createOrder, searchProducts, searchCustomers } from '@/lib/api/orders';
import type { CreateOrderRequest } from '@/lib/api/types';
import { createOrderSchema, type CreateOrderFormData, type OrderItemFormData } from '@/lib/validations/order';
import { useAutocomplete } from '@/lib/hooks/use-autocomplete';

interface OrderCreateModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function OrderCreateModal({ open, onOpenChange }: OrderCreateModalProps) {
  const queryClient = useQueryClient();
  const router = useRouter();
  const [error, setError] = useState<string | null>(null);
  const [customerQuery, setCustomerQuery] = useState('');

  const form = useForm<CreateOrderFormData>({
    resolver: zodResolver(createOrderSchema),
    defaultValues: { customer_id: '', items: [{ sku: '', quantity: 1, unit_price: 0 }] },
  });

  const { results: customerResults, search: searchCustomersFn } = useAutocomplete(searchCustomers, 1);
  const { results: productResults, search: searchProductsFn } = useAutocomplete(searchProducts, 1);

  const mutation = useMutation({
    mutationFn: (data: CreateOrderRequest) => createOrder(data),
    onSuccess: (order) => {
      queryClient.invalidateQueries({ queryKey: ['orders'] });
      queryClient.invalidateQueries({ queryKey: ['inactiveOrders'] });
      onOpenChange(false);
      form.reset();
      router.push(`/orders/${order.id}`);
    },
    onError: (err) => setError(err.message),
  });

  const addNewItem = () => {
    const currentItems = form.getValues('items') || [];
    form.setValue('items', [...currentItems, { sku: '', quantity: 1, unit_price: 0 }]);
  };

  const updateItem = (index: number, field: keyof OrderItemFormData, value: string | number) => {
    const currentItems = form.getValues('items') || [];
    const updated = [...currentItems];
    updated[index] = { ...updated[index], [field]: value };
    form.setValue('items', updated);
  };

  const removeItem = (index: number) => {
    const currentItems = form.getValues('items') || [];
    form.setValue('items', currentItems.filter((_, i) => i !== index));
  };

  const handleCustomerSelect = (id: string) => {
    form.setValue('customer_id', id);
    setCustomerQuery('');
  };

  const onSubmit = (data: CreateOrderFormData) => {
    setError(null);
    const validItems = data.items.filter((item) => item.sku && item.quantity > 0);
    if (validItems.length === 0) {
      setError('At least one valid item is required');
      return;
    }
    const payload: CreateOrderRequest = {
      customer_id: data.customer_id || null,
      items: validItems,
    };
    mutation.mutate(payload);
  };

  return (
    <Modal open={open} onOpenChange={onOpenChange} title="Create order" className="max-w-2xl max-h-[90vh] overflow-y-auto">
      <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
        <div>
          <Label htmlFor="customer">Customer (optional)</Label>
          <div className="relative">
            <Input
              id="customer"
              value={customerQuery}
              onChange={(e) => {
                setCustomerQuery(e.target.value);
                searchCustomersFn(e.target.value);
                form.setValue('customer_id', '');
              }}
              placeholder="Search customer..."
            />
            {customerResults.length > 0 && (
              <div className="absolute z-10 w-full mt-1 bg-white border border-gray-300 rounded-md shadow-lg max-h-48 overflow-y-auto">
                {customerResults.map((customer) => (
                  <button
                    key={customer.id}
                    type="button"
                    className="w-full px-4 py-2 text-left hover:bg-gray-100"
                    onClick={() => handleCustomerSelect(customer.id)}
                  >
                    {customer.name}
                    {customer.phone && <span className="text-gray-500 ml-2">({customer.phone})</span>}
                  </button>
                ))}
              </div>
            )}
          </div>
          {form.formState.errors.customer_id && (
            <p className="text-sm text-[#cf202f] mt-1">{form.formState.errors.customer_id.message}</p>
          )}
        </div>

        <div>
          <Label>Order Items</Label>
          <OrderItemsTable
            items={form.getValues('items') || []}
            onUpdate={updateItem}
            onRemove={removeItem}
            products={productResults}
          />
          <button
            type="button"
            onClick={addNewItem}
            className="mt-2 text-sm text-[#0052ff] hover:underline"
          >
            + Add item
          </button>
          <div className="relative">
            <Input
              value={''}
              onChange={() => {}}
              placeholder="Search products by SKU..."
              className="mt-2"
              onBlur={() => {}}
            />
            {productResults.length > 0 && (
              <div className="absolute z-10 w-full mt-1 bg-white border border-gray-300 rounded-md shadow-lg max-h-48 overflow-y-auto">
                {productResults.map((product) => (
                  <button
                    key={product.sku}
                    type="button"
                    className="w-full px-4 py-2 text-left hover:bg-gray-100"
                    onClick={() => {
                      const firstEmptyItem = form.getValues('items')?.findIndex((i) => !i.sku);
                      if (firstEmptyItem !== undefined && firstEmptyItem >= 0) {
                        updateItem(firstEmptyItem, 'sku', product.sku);
                        updateItem(firstEmptyItem, 'unit_price', product.price);
                      } else {
                        addNewItem();
                        setTimeout(() => {
                          const items = form.getValues('items') || [];
                          updateItem(items.length - 1, 'sku', product.sku);
                          updateItem(items.length - 1, 'unit_price', product.price);
                        }, 0);
                      }
                    }}
                  >
                    {product.sku} - {product.name} (${product.price.toFixed(2)})
                  </button>
                ))}
              </div>
            )}
            <div className="absolute left-3 top-1/2 -translate-y-1/2 pointer-events-none">
              <svg className="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
              </svg>
            </div>
            <input
              type="text"
              className="absolute inset-0 w-full h-full opacity-0 cursor-text"
              onChange={(e) => searchProductsFn(e.target.value)}
              aria-label="Search products"
            />
          </div>
          {form.formState.errors.items && (
            <p className="text-sm text-[#cf202f] mt-1">{form.formState.errors.items.message}</p>
          )}
        </div>

        {error && <p className="text-sm text-[#cf202f]">{error}</p>}

        <div className="flex justify-end gap-3 pt-4 border-t border-gray-200">
          <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button type="submit" disabled={mutation.isPending}>
            {mutation.isPending ? 'Creating...' : 'Create order'}
          </Button>
        </div>
      </form>
    </Modal>
  );
}
