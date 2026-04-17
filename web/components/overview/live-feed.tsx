'use client';
import { useEffect, useRef, useState } from 'react';
import Link from 'next/link';
import { useLogs } from '@/lib/hooks';
import { StatusChip } from '@/components/common/status-chip';
import { GatewayBadge } from '@/components/common/gateway-badge';
import { ErrorState } from '@/components/common/error-state';
import { Spinner } from '@/components/common/spinner';

const HIGHLIGHT_MS = 1000;

export function LiveFeed() {
  const { data, error, mutate } = useLogs({ limit: 50, pollInterval: 2000 });
  const seenRef = useRef<Set<string>>(new Set());
  const firstRenderRef = useRef(true);
  const [highlighted, setHighlighted] = useState<Set<string>>(new Set());

  useEffect(() => {
    if (!data) return;
    const currentIds = data.map((l) => l.id);
    const newlyAdded: string[] = [];
    for (const id of currentIds) {
      if (!seenRef.current.has(id)) {
        // Skip the first data arrival — treat all as already-seen so the
        // initial page load doesn't light up the whole list.
        if (!firstRenderRef.current) newlyAdded.push(id);
        seenRef.current.add(id);
      }
    }
    firstRenderRef.current = false;

    if (newlyAdded.length > 0) {
      setHighlighted((prev) => {
        const next = new Set(prev);
        newlyAdded.forEach((id) => next.add(id));
        return next;
      });
      const timers = newlyAdded.map((id) =>
        setTimeout(() => {
          setHighlighted((prev) => {
            if (!prev.has(id)) return prev;
            const next = new Set(prev);
            next.delete(id);
            return next;
          });
        }, HIGHLIGHT_MS),
      );
      return () => {
        timers.forEach((t) => clearTimeout(t));
      };
    }
  }, [data]);

  if (error) return <ErrorState message="Failed to load live feed" onRetry={() => mutate()} />;
  if (data === undefined)
    return (
      <div className="text-sm text-muted-foreground p-4 flex items-center gap-2">
        <Spinner size="small" />
        <span>Loading…</span>
      </div>
    );
  if (!data || data.length === 0)
    return (
      <div className="p-6 text-center text-sm text-muted-foreground">
        No requests yet — point your app at the mock endpoint above.
      </div>
    );

  return (
    <div className="divide-y">
      {data.map((l) => {
        const isNew = highlighted.has(l.id);
        const shortId = l.id.slice(0, 8);
        return (
          <Link
            key={l.id}
            href={`/logs?id=${l.id}`}
            prefetch={false}
            className={`px-4 py-2 flex items-center gap-3 text-sm font-mono transition-colors hover:bg-[var(--gray-a3)] ${
              isNew
                ? 'animate-in fade-in slide-in-from-top-1 duration-300 bg-[var(--accent-4)]'
                : 'bg-transparent'
            }`}
            style={isNew ? { transition: 'background-color 1s ease-out' } : undefined}
            title={l.id}
          >
            <span className="text-[var(--gray-10)] text-xs tabular-nums w-[70px] shrink-0">
              {shortId}
            </span>
            <StatusChip status={l.response_status} />
            <span className="text-muted-foreground w-14 shrink-0">{l.method}</span>
            <GatewayBadge gateway={l.gateway} />
            <span className="flex-1 truncate">{l.path}</span>
            <span className="text-muted-foreground text-xs">{l.duration_ms}ms</span>
            <span className="text-muted-foreground text-xs">
              {new Date(l.created_at).toLocaleTimeString()}
            </span>
          </Link>
        );
      })}
    </div>
  );
}
