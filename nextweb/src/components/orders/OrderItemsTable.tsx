'use client';

import { useMemo } from 'react';
import { OrderItemFormData } from '@/lib/validations/order';

interface OrderItemsTableProps {
  items: OrderItemFormData[];
  onUpdate: (index: number, field: keyof OrderItemFormData, value: string | number) => void;
  onRemove: (index: number) => void;
  products: { sku: string; name: string; price: number }[];
}

export function OrderItemsTable({ items, onUpdate, onRemove, products }: OrderItemsTableProps) {
  const total = useMemo(() => {
    return items.reduce((sum, item) => sum + item.quantity * item.unit_price, 0);
  }, [items]);

  const handleSkuChange = (index: number, sku: string) => {
    onUpdate(index, 'sku', sku);
    const product = products.find((p) => p.sku.toLowerCase() === sku.toLowerCase());
    if (product) {
      onUpdate(index, 'unit_price', product.price);
    }
  };

  return (
    <div className="space-y-4">
      <div className="overflow-x-auto">
        <table className="w-full">
          <thead>
            <tr className="border-b border-gray-200">
              <th className="text-left py-2 px-2 text-sm font-medium text-gray-500">SKU</th>
              <th className="text-left py-2 px-2 text-sm font-medium text-gray-500">Quantity</th>
              <th className="text-left py-2 px-2 text-sm font-medium text-gray-500">Unit Price</th>
              <th className="text-right py-2 px-2 text-sm font-medium text-gray-500">Subtotal</th>
              <th className="text-right py-2 px-2"></th>
            </tr>
          </thead>
          <tbody>
            {items.map((item, index) => (
              <tr key={index} className="border-b border-gray-100">
                <td className="py-2 px-2">
                  <input
                    type="text"
                    value={item.sku}
                    onChange={(e) => handleSkuChange(index, e.target.value)}
                    className="w-full px-2 py-1 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-[#0052ff]"
                    placeholder="Enter SKU"
                    list={`sku-options-${index}`}
                  />
                  <datalist id={`sku-options-${index}`}>
                    {products.map((p) => (
                      <option key={p.sku} value={p.sku} label={`${p.sku} - ${p.name}`} />
                    ))}
                  </datalist>
                </td>
                <td className="py-2 px-2">
                  <input
                    type="number"
                    min="1"
                    value={item.quantity}
                    onChange={(e) => onUpdate(index, 'quantity', parseInt(e.target.value) || 0)}
                    className="w-20 px-2 py-1 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-[#0052ff]"
                  />
                </td>
                <td className="py-2 px-2">
                  <input
                    type="number"
                    min="0"
                    step="0.01"
                    value={item.unit_price}
                    onChange={(e) => onUpdate(index, 'unit_price', parseFloat(e.target.value) || 0)}
                    className="w-24 px-2 py-1 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-[#0052ff]"
                  />
                </td>
                <td className="py-2 px-2 text-right font-medium">
                  ${(item.quantity * item.unit_price).toFixed(2)}
                </td>
                <td className="py-2 px-2 text-right">
                  <button
                    type="button"
                    onClick={() => onRemove(index)}
                    className="text-[#cf202f] hover:text-red-700"
                  >
                    Remove
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
          <tfoot>
            <tr className="border-t-2 border-gray-200">
              <td colSpan={3} className="py-3 px-2 text-right font-semibold text-[#0a0b0d]">Total</td>
              <td className="py-3 px-2 text-right font-bold text-xl text-[#0052ff]">${total.toFixed(2)}</td>
              <td></td>
            </tr>
          </tfoot>
        </table>
      </div>
      {items.length === 0 && (
        <p className="text-sm text-gray-500 text-center py-4">No items added yet</p>
      )}
    </div>
  );
}
