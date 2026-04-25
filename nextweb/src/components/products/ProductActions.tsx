'use client';

import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useRouter } from 'next/navigation';
import { Button } from '@/components/ui/button';
import { deactivateProduct, activateProduct } from '@/lib/api/products';

interface ProductActionsProps {
  productId: string;
  isActive: boolean;
}

export function ProductActions({ productId, isActive }: ProductActionsProps) {
  const queryClient = useQueryClient();
  const router = useRouter();

  const deactivateMutation = useMutation({
    mutationFn: () => deactivateProduct(productId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['products'] });
      queryClient.invalidateQueries({ queryKey: ['product', productId] });
    },
  });

  const activateMutation = useMutation({
    mutationFn: () => activateProduct(productId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['inactiveProducts'] });
      queryClient.invalidateQueries({ queryKey: ['product', productId] });
      router.push('/products');
    },
  });

  if (isActive) {
    return (
      <Button
        onClick={() => {
          if (confirm('Are you sure you want to deactivate this product?')) {
            deactivateMutation.mutate();
          }
        }}
        disabled={deactivateMutation.isPending}
        className="bg-[#cf202f] hover:bg-red-700"
      >
        {deactivateMutation.isPending ? 'Deactivating...' : 'Deactivate product'}
      </Button>
    );
  }

  return (
    <Button
      onClick={() => activateMutation.mutate()}
      disabled={activateMutation.isPending}
      className="bg-[#0052ff] hover:bg-blue-700"
    >
      {activateMutation.isPending ? 'Activating...' : 'Activate'}
    </Button>
  );
}
