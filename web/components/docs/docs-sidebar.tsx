'use client';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { useState } from 'react';
import { ChevronDownIcon, ChevronRightIcon } from '@radix-ui/react-icons';

const TOP_LINKS = [
  { href: '/docs/about', label: 'About TestPay' },
  { href: '/docs', label: 'Getting started' },
  { href: '/docs/point-your-sdk', label: 'Point your SDK at the mock' },
  { href: '/docs/scenarios', label: 'Scenarios' },
  { href: '/docs/failure-modes', label: 'Failure modes reference' },
  { href: '/docs/api', label: 'API reference' },
  { href: '/docs/webhooks', label: 'Webhook spec' },
];

// Adapter order matches the Overview hero chip order: rich shapes first
// then the agnostic fallback at the bottom.
const ADAPTERS = [
  { slug: 'stripe', label: 'Stripe' },
  { slug: 'razorpay', label: 'Razorpay' },
  { slug: 'adyen', label: 'Adyen' },
  { slug: 'mastercard', label: 'Mastercard (MPGS)' },
  { slug: 'tillpay', label: 'TillPayment' },
  { slug: 'ecpay', label: 'ECPay' },
  { slug: 'espay', label: 'ESPay' },
  { slug: 'instamojo', label: 'Instamojo' },
  { slug: 'komoju', label: 'Komoju' },
  { slug: 'paynamics', label: 'Paynamics' },
  { slug: 'tappay', label: 'TapPay' },
  { slug: 'agnostic', label: 'Agnostic (/v1)' },
];

export function DocsSidebar() {
  const pathname = usePathname();
  const [adaptersOpen, setAdaptersOpen] = useState(
    pathname.startsWith('/docs/adapters'),
  );

  return (
    <aside className="w-60 shrink-0 sticky top-0 self-start border-r pr-4 h-[calc(100vh-4rem)] overflow-y-auto py-6">
      <div className="text-[10px] uppercase tracking-wider text-muted-foreground mb-2 px-3">
        Documentation
      </div>
      <nav className="space-y-0.5">
        {TOP_LINKS.map((it) => {
          const active = pathname === it.href;
          return (
            <Link
              key={it.href}
              href={it.href}
              className={`block px-3 py-1.5 rounded-md text-sm transition-colors ${
                active
                  ? 'bg-accent text-accent-foreground'
                  : 'text-muted-foreground hover:bg-accent/50 hover:text-foreground'
              }`}
            >
              {it.label}
            </Link>
          );
        })}

        <button
          type="button"
          onClick={() => setAdaptersOpen((o) => !o)}
          className="w-full flex items-center gap-1 px-3 py-1.5 rounded-md text-sm text-muted-foreground hover:bg-accent/50 hover:text-foreground transition-colors mt-2"
        >
          {adaptersOpen ? (
            <ChevronDownIcon className="h-3.5 w-3.5" />
          ) : (
            <ChevronRightIcon className="h-3.5 w-3.5" />
          )}
          <span className="text-[10px] uppercase tracking-wider">Gateway adapters</span>
        </button>
        {adaptersOpen ? (
          <div className="space-y-0.5 pl-2">
            {ADAPTERS.map((a) => {
              const href = `/docs/adapters/${a.slug}`;
              const active = pathname === href;
              return (
                <Link
                  key={a.slug}
                  href={href}
                  className={`block px-3 py-1.5 rounded-md text-sm transition-colors ${
                    active
                      ? 'bg-accent text-accent-foreground'
                      : 'text-muted-foreground hover:bg-accent/50 hover:text-foreground'
                  }`}
                >
                  {a.label}
                </Link>
              );
            })}
          </div>
        ) : null}
      </nav>
    </aside>
  );
}
