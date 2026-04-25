'use client';

import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useRouter } from 'next/navigation';
import { Button } from '@/components/ui/button';
import { completePayment, deactivateOrder } from '@/lib/api/orders';

interface OrderActionsProps {
  orderId: string;
  status: string | null | undefined;
  isActive: boolean;
}

export function OrderActions({ orderId, status, isActive }: OrderActionsProps) {
  const queryClient = useQueryClient();
  const router = useRouter();

  const completePaymentMutation = useMutation({
    mutationFn: () => completePayment(orderId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['orders'] });
      queryClient.invalidateQueries({ queryKey: ['order', orderId] });
      queryClient.invalidateQueries({ queryKey: ['paymentCompletedOrders'] });
    },
  });

  const deactivateMutation = useMutation({
    mutationFn: () => deactivateOrder(orderId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['orders'] });
      queryClient.invalidateQueries({ queryKey: ['paymentCompletedOrders'] });
      queryClient.invalidateQueries({ queryKey: ['order', orderId] });
    },
  });

  if (!isActive) {
    return null;
  }

  if (status === 'payment_completed') {
    return (
      <Button
        onClick={() => {
          if (confirm('Are you sure you want to deactivate this order?')) {
            deactivateMutation.mutate();
          }
        }}
        disabled={deactivateMutation.isPending}
        className="bg-[#cf202f] hover:bg-red-700"
      >
        {deactivateMutation.isPending ? 'Deactivating...' : 'Deactivate order'}
      </Button>
    );
  }

  return (
    <Button
      onClick={() => completePaymentMutation.mutate()}
      disabled={completePaymentMutation.isPending}
      className="bg-[#0052ff] hover:bg-blue-700"
    >
      {completePaymentMutation.isPending ? 'Processing...' : 'Complete payment'}
    </Button>
  );
}
