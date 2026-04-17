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
} from '@radix-ui/react-icons';
import type { User, Workspace } from '@/lib/types';

const items = [
  { href: '/', label: 'Overview', icon: DashboardIcon, hint: 'Live activity + stats' },
  { href: '/scenarios', label: 'Scenarios', icon: ListBulletIcon, hint: 'Failure-mode sequences' },
  { href: '/logs', label: 'Logs', icon: FileTextIcon, hint: 'Every mock request' },
  { href: '/webhooks', label: 'Webhooks', icon: PaperPlaneIcon, hint: 'Outbound deliveries' },
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
        {items.map((it) => {
          const active = pathname === it.href || (it.href !== '/' && pathname.startsWith(it.href));
          const Icon = it.icon;
          return (
            <Link
              key={it.href}
              href={it.href}
              className={`flex items-center gap-3 px-3 py-2 rounded-md text-sm transition-colors ${
                active
                  ? 'bg-accent text-accent-foreground'
                  : 'text-muted-foreground hover:bg-accent/50 hover:text-foreground'
              }`}
              title={it.hint}
            >
              <Icon className="h-4 w-4" />
              {it.label}
            </Link>
          );
        })}
      </nav>
      <div className="p-3 border-t text-[10px] text-muted-foreground/70 uppercase tracking-wider">
        TestPay
      </div>
    </aside>
  );
}
