'use client';

import { useQuery } from '@tanstack/react-query';
import { listPaymentCompletedOrders } from '@/lib/api/orders';
import { Card } from '@/components/Card';

export default function PaymentCompletedOrdersPage() {
  const { data: orders, isLoading } = useQuery({
    queryKey: ['paymentCompletedOrders'],
    queryFn: () => listPaymentCompletedOrders(100),
  });

  if (isLoading) return <div className="text-gray-500">Loading...</div>;

  return (
    <>
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-3xl font-bold text-[#0a0b0d]">Payment Completed</h1>
      </div>

      <div className="grid grid-cols-[repeat(auto-fill,minmax(260px,1fr))] gap-4">
        {orders?.map((order) => (
          <Card
            key={order.id}
            href={`/orders/${order.id}`}
            title={order.customer_name || 'Guest Order'}
            subtitle={`Order #${order.id.slice(0, 8)}...`}
            price={order.total}
          />
        ))}
      </div>
    </>
  );
}
