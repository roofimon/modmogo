'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { cn } from '@/lib/utils';

export function Navigation() {
  const pathname = usePathname();

  const navItems = [
    { href: '/products', label: 'Products' },
    { href: '/customers', label: 'Customers' },
    { href: '/orders', label: 'Orders' },
  ];

  return (
    <header className="bg-[#0a0b0d] text-white">
      <nav className="max-w-[1120px] mx-auto px-4">
        <div className="flex items-center h-16 gap-6">
          <Link href="/products" className="text-lg font-semibold tracking-tight">
            modmono
          </Link>
          <div className="flex gap-1">
            {navItems.map((item) => (
              <Link
                key={item.href}
                href={item.href}
                className={cn(
                  'px-4 py-2 rounded-lg transition-colors',
                  pathname?.startsWith(item.href)
                    ? 'bg-white/10 text-white'
                    : 'text-gray-400 hover:text-white hover:bg-white/5'
                )}
              >
                {item.label}
              </Link>
            ))}
          </div>
        </div>
      </nav>
    </header>
  );
}
