'use client';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import {
  DashboardIcon,
  ListBulletIcon,
  FileTextIcon,
  GearIcon,
  LightningBoltIcon,
  PaperPlaneIcon,
  ReaderIcon,
} from '@radix-ui/react-icons';
import type { User, Workspace } from '@/lib/types';

// Primary nav — the product surfaces.
const items = [
  { href: '/', label: 'Overview', icon: DashboardIcon, hint: 'Live activity + stats' },
  { href: '/scenarios', label: 'Scenarios', icon: ListBulletIcon, hint: 'Failure-mode sequences' },
  { href: '/logs', label: 'Logs', icon: FileTextIcon, hint: 'Every mock request' },
  { href: '/webhooks', label: 'Webhooks', icon: PaperPlaneIcon, hint: 'Outbound deliveries' },
];

// Secondary nav — Docs + Settings pinned to the sidebar footer, matching
// the Linear / Supabase pattern (reference-style links separated from
// day-to-day product surfaces).
const footerItems = [
  { href: '/docs', label: 'Docs', icon: ReaderIcon, hint: 'How to use TestPay, per-gateway' },
  { href: '/settings', label: 'Settings', icon: GearIcon, hint: 'Keys, endpoints, theme' },
];

// User menu lives in the topbar (standard SaaS pattern); sidebar is purely nav.
export function Sidebar({ workspace }: { user: User; workspace: Workspace }) {
  const pathname = usePathname();

  return (
    <aside className="w-60 shrink-0 border-r flex flex-col bg-card h-screen sticky top-0">
      <div className="p-4 border-b flex items-center gap-2 font-semibold">
        <LightningBoltIcon className="h-5 w-5 text-emerald-500" />
        <span>TestPay</span>
      </div>
      <div className="p-4 text-xs text-muted-foreground border-b">
        <div className="uppercase tracking-wider mb-1">Workspace</div>
        <div className="font-mono text-foreground truncate">{workspace.slug}</div>
      </div>
      <nav className="flex-1 p-2 space-y-0.5 overflow-y-auto">
        {items.map((it) => (
          <SidebarLink key={it.href} item={it} pathname={pathname} />
        ))}
      </nav>
      <div className="p-2 border-t space-y-0.5">
        {footerItems.map((it) => (
          <SidebarLink key={it.href} item={it} pathname={pathname} />
        ))}
      </div>
      <div className="px-3 py-2 border-t text-[10px] text-muted-foreground/70 uppercase tracking-wider">
        TestPay
      </div>
    </aside>
  );
}

type NavItem = {
  href: string;
  label: string;
  icon: React.ComponentType<{ className?: string }>;
  hint: string;
};

function SidebarLink({ item, pathname }: { item: NavItem; pathname: string }) {
  const active = pathname === item.href || (item.href !== '/' && pathname.startsWith(item.href));
  const Icon = item.icon;
  return (
    <Link
      href={item.href}
      className={`flex items-center gap-3 px-3 py-2 rounded-md text-sm transition-colors ${
        active
          ? 'bg-accent text-accent-foreground'
          : 'text-muted-foreground hover:bg-accent/50 hover:text-foreground'
      }`}
      title={item.hint}
    >
      <Icon className="h-4 w-4" />
      {item.label}
    </Link>
  );
}
