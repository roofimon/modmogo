'use client';

import { useQuery } from '@tanstack/react-query';
import { getOrderById } from '@/lib/api/orders';
import { OrderActions } from '@/components/orders/OrderActions';
import {
  Table,
  TableBody,
  TableCell,
  TableFooter,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';

interface OrderDetailClientProps {
  id: string;
}

export function OrderDetailClient({ id }: OrderDetailClientProps) {
  const { data: order, isLoading } = useQuery({
    queryKey: ['order', id],
    queryFn: () => getOrderById(id),
    staleTime: 1000 * 60 * 5,
  });

  if (isLoading) return <div className="text-gray-500">Loading...</div>;
  if (!order) return <div className="text-gray-500">Order not found</div>;

  const isActive = !order.deactivated_at;

  return (
    <div>
      <div className="flex justify-between items-start mb-6">
        <h1 className="text-3xl font-bold text-[#0a0b0d]">Order #{order.id.slice(0, 8)}</h1>
        {isActive && <OrderActions orderId={order.id} status={order.status} isActive={true} />}
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <dl className="bg-white rounded-xl border border-gray-200 p-6 space-y-4">
          <div className="flex justify-between">
            <dt className="text-gray-500">Customer</dt>
            <dd className="font-medium">{order.customer_name || 'Guest'}</dd>
          </div>
          <div className="flex justify-between">
            <dt className="text-gray-500">Status</dt>
            <dd className="font-medium">
              <span className={`px-2 py-1 rounded-full text-sm ${order.status === 'payment_completed' ? 'bg-green-100 text-green-800' : 'bg-yellow-100 text-yellow-800'}`}>
                {order.status === 'payment_completed' ? 'Payment Completed' : 'Pending'}
              </span>
            </dd>
          </div>
          <div className="flex justify-between">
            <dt className="text-gray-500">Created</dt>
            <dd className="font-medium">{new Date(order.created_at).toLocaleDateString()}</dd>
          </div>
          {order.original_order_id && (
            <div className="flex justify-between">
              <dt className="text-gray-500">Original Order</dt>
              <dd className="font-medium">{order.original_order_id.slice(0, 8)}</dd>
            </div>
          )}
        </dl>

        <div className="bg-white rounded-xl border border-gray-200 p-6">
          <h2 className="text-lg font-semibold mb-4">Order Items</h2>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>SKU</TableHead>
                <TableHead>Quantity</TableHead>
                <TableHead>Unit Price</TableHead>
                <TableHead className="text-right">Subtotal</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {order.items.map((item, index) => (
                <TableRow key={index}>
                  <TableCell>
                    {item.sku}
                    {item.product_name && (
                      <div className="text-sm text-gray-500">{item.product_name}</div>
                    )}
                  </TableCell>
                  <TableCell>{item.quantity}</TableCell>
                  <TableCell>${item.unit_price.toFixed(2)}</TableCell>
                  <TableCell className="text-right">
                    ${(item.quantity * item.unit_price).toFixed(2)}
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
            <TableFooter>
              <TableRow>
                <TableCell colSpan={3} className="text-right font-semibold">Total</TableCell>
                <TableCell className="text-right font-bold text-xl text-[#0052ff]">
                  ${order.total.toFixed(2)}
                </TableCell>
              </TableRow>
            </TableFooter>
          </Table>
        </div>
      </div>
    </div>
  );
}
