'use client';

import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { listCustomers } from '@/lib/api/customers';
import { Card } from '@/components/Card';
import { CustomerCreateModal } from '@/components/customers/CustomerCreateModal';

export default function CustomersPage() {
  const [showCreateModal, setShowCreateModal] = useState(false);

  const { data: customers, isLoading } = useQuery({
    queryKey: ['customers'],
    queryFn: () => listCustomers(100),
  });

  if (isLoading) return <div className="text-gray-500">Loading...</div>;

  return (
    <>
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-3xl font-bold text-[#0a0b0d]">Customers</h1>
        <button
          onClick={() => setShowCreateModal(true)}
          className="bg-[#0052ff] text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition-colors"
        >
          Add customer
        </button>
      </div>

      <div className="grid grid-cols-[repeat(auto-fill,minmax(260px,1fr))] gap-4">
        {customers?.map((customer) => (
          <Card
            key={customer.id}
            href={`/customers/${customer.id}`}
            title={customer.name}
            subtitle={customer.email}
          />
        ))}
      </div>

      <CustomerCreateModal open={showCreateModal} onOpenChange={setShowCreateModal} />
    </>
  );
}
