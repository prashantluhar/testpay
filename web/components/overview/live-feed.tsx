'use client';
import { useLogs } from '@/lib/hooks';
import { StatusChip } from '@/components/common/status-chip';
import { GatewayBadge } from '@/components/common/gateway-badge';
import { ErrorState } from '@/components/common/error-state';

export function LiveFeed() {
  const { data, error, mutate } = useLogs({ limit: 50, pollInterval: 2000 });

  if (error) return <ErrorState message="Failed to load live feed" onRetry={() => mutate()} />;
  if (!data) return <div className="text-sm text-muted-foreground p-4">Loading…</div>;
  if (data.length === 0)
    return (
      <div className="p-6 text-center text-sm text-muted-foreground">
        No requests yet — point your app at the mock endpoint above.
      </div>
    );

  return (
    <div className="divide-y">
      {data.map((l) => (
        <div key={l.ID} className="px-4 py-2 flex items-center gap-3 text-sm font-mono">
          <StatusChip status={l.ResponseStatus} />
          <span className="text-muted-foreground w-14">{l.Method}</span>
          <GatewayBadge gateway={l.Gateway} />
          <span className="flex-1 truncate">{l.Path}</span>
          <span className="text-muted-foreground text-xs">{l.DurationMs}ms</span>
          <span className="text-muted-foreground text-xs">
            {new Date(l.CreatedAt).toLocaleTimeString()}
          </span>
        </div>
      ))}
    </div>
  );
}
