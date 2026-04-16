'use client';
import { usePathname } from 'next/navigation';
import { Badge } from '@/components/ui/badge';
import { CopyButton } from '@/components/common/copy-button';
import { MODE } from '@/lib/types';
import { UserMenu } from './user-menu';
import type { User, Workspace } from '@/lib/types';

export function Topbar({ user, workspace }: { user: User; workspace: Workspace }) {
  const pathname = usePathname();
  const crumb = pathname === '/' ? 'Overview' : pathname.split('/').filter(Boolean).join(' / ');

  const baseUrl =
    MODE === 'local' ? 'http://localhost:7700' : `https://api.testpay.dev/ws_${workspace.slug}`;

  return (
    <header className="sticky top-0 z-30 h-14 border-b px-6 flex items-center justify-between bg-card/95 backdrop-blur supports-[backdrop-filter]:bg-card/80">
      <div className="text-sm text-muted-foreground capitalize truncate">{crumb}</div>
      <div className="flex items-center gap-3">
        <Badge variant="outline" className="font-mono text-xs">
          {MODE}
        </Badge>
        <span className="font-mono text-xs text-muted-foreground hidden lg:inline">{baseUrl}</span>
        <CopyButton value={baseUrl} label="" />
        <div className="h-6 w-px bg-border" />
        <UserMenu user={user} />
      </div>
    </header>
  );
}
