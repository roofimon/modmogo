'use client';

import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { listProducts } from '@/lib/api/products';
import { Card } from '@/components/Card';
import { ProductCreateModal } from '@/components/products/ProductCreateModal';

export default function ProductsPage() {
  const [showCreateModal, setShowCreateModal] = useState(false);

  const { data: products, isLoading } = useQuery({
    queryKey: ['products'],
    queryFn: () => listProducts(100),
  });

  if (isLoading) return <div className="text-gray-500">Loading...</div>;

  return (
    <>
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-3xl font-bold text-[#0a0b0d]">Products</h1>
        <button
          onClick={() => setShowCreateModal(true)}
          className="bg-[#0052ff] text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition-colors"
        >
          Add product
        </button>
      </div>

      <div className="grid grid-cols-[repeat(auto-fill,minmax(260px,1fr))] gap-4">
        {products?.map((product) => (
          <Card
            key={product.id}
            href={`/products/${product.id}`}
            title={product.name}
            subtitle={product.sku}
            price={product.price}
          />
        ))}
      </div>

      <ProductCreateModal open={showCreateModal} onOpenChange={setShowCreateModal} />
    </>
  );
}
