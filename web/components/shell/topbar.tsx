'use client';
import { usePathname } from 'next/navigation';
import { Badge } from '@radix-ui/themes';
import { CopyButton } from '@/components/common/copy-button';
import { MODE } from '@/lib/types';
import { UserMenu } from './user-menu';
import type { User } from '@/lib/types';

export function Topbar({ user }: { user: User }) {
  const pathname = usePathname();
  const crumb = pathname === '/' ? 'Overview' : pathname.split('/').filter(Boolean).join(' / ');

  const baseUrl =
    MODE === 'local' ? 'http://localhost:7700' : (process.env.NEXT_PUBLIC_API_BASE || '');

  return (
    <header className="sticky top-0 z-30 h-14 border-b px-6 flex items-center justify-between bg-card/95 backdrop-blur supports-[backdrop-filter]:bg-card/80">
      <div className="text-sm text-muted-foreground capitalize truncate">{crumb}</div>
      <div className="flex items-center gap-3">
        <Badge variant="outline" color="gray" className="font-mono">
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
