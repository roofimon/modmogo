'use client';

import { useQuery } from '@tanstack/react-query';
import { getProductById } from '@/lib/api/products';
import { ProductActions } from '@/components/products/ProductActions';

interface ProductDetailClientProps {
  id: string;
}

export function ProductDetailClient({ id }: ProductDetailClientProps) {
  const { data: product, isLoading } = useQuery({
    queryKey: ['product', id],
    queryFn: () => getProductById(id),
    staleTime: 1000 * 60 * 5,
  });

  if (isLoading) return <div className="text-gray-500">Loading...</div>;
  if (!product) return <div className="text-gray-500">Product not found</div>;

  const isActive = !product.deactivated_at;

  return (
    <div>
      <div className="flex justify-between items-start mb-6">
        <h1 className="text-3xl font-bold text-[#0a0b0d]">{product.name}</h1>
        {isActive && <ProductActions productId={product.id} isActive={true} />}
      </div>

      <dl className="bg-white rounded-xl border border-gray-200 p-6 space-y-4">
        <div className="flex justify-between">
          <dt className="text-gray-500">SKU</dt>
          <dd className="font-medium">{product.sku}</dd>
        </div>
        <div className="flex justify-between">
          <dt className="text-gray-500">Price</dt>
          <dd className="font-medium">${product.price.toFixed(2)}</dd>
        </div>
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
          <dd className="font-medium">{new Date(product.created_at).toLocaleDateString()}</dd>
        </div>
      </dl>
    </div>
  );
}
