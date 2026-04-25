'use client';

import { useQuery } from '@tanstack/react-query';
import { listInactiveCustomers } from '@/lib/api/customers';
import { Card } from '@/components/Card';

export default function InactiveCustomersPage() {
  const { data: customers, isLoading } = useQuery({
    queryKey: ['inactiveCustomers'],
    queryFn: () => listInactiveCustomers(100),
  });

  if (isLoading) return <div className="text-gray-500">Loading...</div>;

  return (
    <>
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-3xl font-bold text-[#0a0b0d]">Inactive Customers</h1>
      </div>

      <div className="grid grid-cols-[repeat(auto-fill,minmax(260px,1fr))] gap-4">
        {customers?.map((customer) => (
          <Card
            key={customer.id}
            href={`/customers/${customer.id}`}
            title={customer.name}
            subtitle={customer.email}
            status="deactivated"
          />
        ))}
      </div>
    </>
  );
}
