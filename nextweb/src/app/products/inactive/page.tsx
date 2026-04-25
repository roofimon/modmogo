'use client';

import { useQuery } from '@tanstack/react-query';
import { listInactiveProducts } from '@/lib/api/products';
import { Card } from '@/components/Card';
import { ProductActions } from '@/components/products/ProductActions';

export default function InactiveProductsPage() {
  const { data: products, isLoading } = useQuery({
    queryKey: ['inactiveProducts'],
    queryFn: () => listInactiveProducts(100),
  });

  if (isLoading) return <div className="text-gray-500">Loading...</div>;

  return (
    <>
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-3xl font-bold text-[#0a0b0d]">Inactive Products</h1>
      </div>

      <div className="grid grid-cols-[repeat(auto-fill,minmax(260px,1fr))] gap-4">
        {products?.map((product) => (
          <div key={product.id} className="relative">
            <Card
              href={`/products/${product.id}`}
              title={product.name}
              subtitle={product.sku}
              price={product.price}
              status="deactivated"
            />
            <div className="mt-3">
              <ProductActions productId={product.id} isActive={false} />
            </div>
          </div>
        ))}
      </div>
    </>
  );
}
