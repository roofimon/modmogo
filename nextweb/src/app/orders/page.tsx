'use client';

import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { listOrders } from '@/lib/api/orders';
import { Card } from '@/components/Card';
import { OrderCreateModal } from '@/components/orders/OrderCreateModal';

export default function OrdersPage() {
  const [showCreateModal, setShowCreateModal] = useState(false);

  const { data: orders, isLoading } = useQuery({
    queryKey: ['orders'],
    queryFn: () => listOrders(100),
  });

  if (isLoading) return <div className="text-gray-500">Loading...</div>;

  return (
    <>
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-3xl font-bold text-[#0a0b0d]">Orders</h1>
        <button
          onClick={() => setShowCreateModal(true)}
          className="bg-[#0052ff] text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition-colors"
        >
          Add order
        </button>
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

      <OrderCreateModal open={showCreateModal} onOpenChange={setShowCreateModal} />
    </>
  );
}
