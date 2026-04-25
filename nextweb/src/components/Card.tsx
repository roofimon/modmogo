import Link from 'next/link';
import { cn } from '@/lib/utils';

interface CardProps {
  href: string;
  title: string;
  subtitle?: string;
  price?: number;
  status?: string;
  className?: string;
}

export function Card({ href, title, subtitle, price, status, className }: CardProps) {
  return (
    <Link href={href} className={cn('card', className)}>
      <div className="bg-white rounded-xl border border-gray-200 p-6 hover:shadow-lg hover:-translate-y-0.5 transition-all duration-200 h-full flex flex-col">
        {status === 'deactivated' && (
          <span className="absolute top-4 right-4 px-2 py-1 text-xs font-medium bg-gray-100 text-gray-600 rounded-full">
            Inactive
          </span>
        )}
        <h3 className="text-lg font-semibold text-[#0a0b0d] mb-1">{title}</h3>
        {subtitle && <p className="text-sm text-gray-500 mb-3">{subtitle}</p>}
        {price !== undefined && (
          <p className="card-price text-xl font-bold text-[#0052ff] mt-auto">${price.toFixed(2)}</p>
        )}
      </div>
    </Link>
  );
}
