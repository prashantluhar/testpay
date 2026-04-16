'use client';
import { usePathname } from 'next/navigation';
import { Badge } from '@/components/ui/badge';
import { CopyButton } from '@/components/common/copy-button';
import { MODE } from '@/lib/types';
import type { Workspace } from '@/lib/types';

export function Topbar({ workspace }: { workspace: Workspace }) {
  const pathname = usePathname();
  const crumb = pathname === '/' ? 'Overview' : pathname.split('/').filter(Boolean).join(' / ');

  const baseUrl =
    MODE === 'local'
      ? 'http://localhost:7700'
      : `https://api.testpay.dev/ws_${workspace.slug}`;

  return (
    <header className="h-14 border-b px-6 flex items-center justify-between bg-card">
      <div className="text-sm text-muted-foreground capitalize">{crumb}</div>
      <div className="flex items-center gap-3">
        <Badge variant="outline" className="font-mono text-xs">
          {MODE}
        </Badge>
        <span className="font-mono text-xs text-muted-foreground">{baseUrl}</span>
        <CopyButton value={baseUrl} label="" />
      </div>
    </header>
  );
}
