'use client';

import { useQuery } from '@tanstack/react-query';
import { getCustomerById } from '@/lib/api/customers';
import { CustomerActions } from '@/components/customers/CustomerActions';

interface CustomerDetailClientProps {
  id: string;
}

export function CustomerDetailClient({ id }: CustomerDetailClientProps) {
  const { data: customer, isLoading } = useQuery({
    queryKey: ['customer', id],
    queryFn: () => getCustomerById(id),
    staleTime: 1000 * 60 * 5,
  });

  if (isLoading) return <div className="text-gray-500">Loading...</div>;
  if (!customer) return <div className="text-gray-500">Customer not found</div>;

  const isActive = !customer.deactivated_at;

  return (
    <div>
      <div className="flex justify-between items-start mb-6">
        <h1 className="text-3xl font-bold text-[#0a0b0d]">{customer.name}</h1>
        {isActive && <CustomerActions customerId={customer.id} />}
      </div>

      <dl className="bg-white rounded-xl border border-gray-200 p-6 space-y-4">
        <div className="flex justify-between">
          <dt className="text-gray-500">Name</dt>
          <dd className="font-medium">{customer.name}</dd>
        </div>
        <div className="flex justify-between">
          <dt className="text-gray-500">Email</dt>
          <dd className="font-medium">{customer.email}</dd>
        </div>
        {customer.phone && (
          <div className="flex justify-between">
            <dt className="text-gray-500">Phone</dt>
            <dd className="font-medium">{customer.phone}</dd>
          </div>
        )}
        <div className="flex justify-between">
          <dt className="text-gray-500">Status</dt>
          <dd className="font-medium">
            <span className={`px-2 py-1 rounded-full text-sm ${isActive ? 'bg-green-100 text-green-800' : 'bg-gray-100 text-gray-600'}`}>
              {isActive ? 'Active' : 'Inactive'}
            </span>
          </dd>
        </div>
        <div className="flex justify-between">
          <dt className="text-gray-500">Created</dt>
          <dd className="font-medium">{new Date(customer.created_at).toLocaleDateString()}</dd>
        </div>
      </dl>
    </div>
  );
}
