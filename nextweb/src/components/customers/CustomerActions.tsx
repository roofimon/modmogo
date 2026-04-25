'use client';

import { useMutation, useQueryClient } from '@tanstack/react-query';
import { Button } from '@/components/ui/button';
import { deactivateCustomer } from '@/lib/api/customers';

interface CustomerActionsProps {
  customerId: string;
}

export function CustomerActions({ customerId }: CustomerActionsProps) {
  const queryClient = useQueryClient();

  const deactivateMutation = useMutation({
    mutationFn: () => deactivateCustomer(customerId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['customers'] });
      queryClient.invalidateQueries({ queryKey: ['customer', customerId] });
    },
  });

  return (
    <Button
      onClick={() => {
        if (confirm('Are you sure you want to deactivate this customer?')) {
          deactivateMutation.mutate();
        }
      }}
      disabled={deactivateMutation.isPending}
      className="bg-[#cf202f] hover:bg-red-700"
    >
      {deactivateMutation.isPending ? 'Deactivating...' : 'Deactivate customer'}
    </Button>
  );
}
