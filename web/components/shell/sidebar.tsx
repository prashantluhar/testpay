'use client';
import Link from 'next/link';
import { usePathname, useRouter } from 'next/navigation';
import { LayoutDashboard, ListTodo, ScrollText, Settings, Zap, LogOut, Send } from 'lucide-react';
import { api } from '@/lib/api';
import type { User, Workspace } from '@/lib/types';

const items = [
  { href: '/', label: 'Overview', icon: LayoutDashboard, hint: 'Live activity + stats' },
  { href: '/scenarios', label: 'Scenarios', icon: ListTodo, hint: 'Failure-mode sequences' },
  { href: '/logs', label: 'Logs', icon: ScrollText, hint: 'Every mock request' },
  { href: '/webhooks', label: 'Webhooks', icon: Send, hint: 'Outbound deliveries' },
  { href: '/settings', label: 'Settings', icon: Settings, hint: 'Keys, endpoints, theme' },
];

export function Sidebar({ user, workspace }: { user: User; workspace: Workspace }) {
  const pathname = usePathname();
  const router = useRouter();

  async function signOut() {
    try {
      await api('/api/auth/logout', { method: 'POST' });
    } catch {
      /* ignore */
    }
    router.push('/login');
  }

  return (
    <aside className="w-60 shrink-0 border-r flex flex-col bg-card">
      <div className="p-4 border-b flex items-center gap-2 font-semibold">
        <Zap className="h-5 w-5 text-emerald-500" />
        <span>TestPay</span>
      </div>
      <div className="p-4 text-xs text-muted-foreground border-b">
        <div className="uppercase tracking-wider mb-1">Workspace</div>
        <div className="font-mono text-foreground truncate">{workspace.slug}</div>
      </div>
      <nav className="flex-1 p-2 space-y-0.5">
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
      <div className="p-4 border-t text-xs space-y-2">
        <div className="flex items-center justify-between">
          <span className="text-muted-foreground truncate">{user.email}</span>
          <button
            onClick={signOut}
            className="text-muted-foreground hover:text-foreground"
            aria-label="Sign out"
            title="Sign out"
          >
            <LogOut className="h-4 w-4" />
          </button>
        </div>
      </div>
    </aside>
  );
}
