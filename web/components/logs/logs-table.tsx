'use client';
import { useEffect, useMemo, useRef, useState } from 'react';
import { Table } from '@radix-ui/themes';
import { StatusChip } from '@/components/common/status-chip';
import { GatewayBadge } from '@/components/common/gateway-badge';
import { TableSkeleton } from '@/components/common/table-skeleton';
import { useScenarios } from '@/lib/hooks';
import type { RequestLog } from '@/lib/types';

// Short relative-time like "4s", "2m", "1h", else an absolute short timestamp.
function shortTime(iso: string) {
  const then = new Date(iso).getTime();
  const diff = Date.now() - then;
  if (diff < 60_000) return `${Math.max(1, Math.floor(diff / 1000))}s ago`;
  if (diff < 3_600_000) return `${Math.floor(diff / 60_000)}m ago`;
  if (diff < 86_400_000) return `${Math.floor(diff / 3_600_000)}h ago`;
  return new Date(iso).toLocaleDateString();
}

export function LogsTable({
  rows,
  onSelect,
  selectedId,
  loading = false,
  filterKey,
}: {
  rows: RequestLog[];
  onSelect: (id: string) => void;
  selectedId?: string | null;
  loading?: boolean;
  // Changing this value re-keys the table body so the rows replay their
  // fade-in animation — used by the logs page when filters change.
  filterKey?: string;
}) {
  // Bump a nonce whenever the filter signature changes so row animations replay.
  const [nonce, setNonce] = useState(0);
  const lastKeyRef = useRef<string | undefined>(filterKey);
  useEffect(() => {
    if (filterKey !== lastKeyRef.current) {
      lastKeyRef.current = filterKey;
      setNonce((n) => n + 1);
    }
  }, [filterKey]);

  // Map scenario IDs → names so the Scenario column shows something useful
  // instead of a raw UUID.
  const { data: scenarios = [] } = useScenarios();
  const scenarioNames = useMemo(() => {
    const m: Record<string, string> = {};
    for (const s of scenarios) m[s.id] = s.name;
    return m;
  }, [scenarios]);

  return (
    <Table.Root size="1">
      <Table.Header className="sticky top-0 bg-card/95 backdrop-blur z-10">
        <Table.Row>
          <Table.ColumnHeaderCell className="w-24 text-xs uppercase tracking-wider">
            ID
          </Table.ColumnHeaderCell>
          <Table.ColumnHeaderCell className="w-32 text-xs uppercase tracking-wider">
            Time
          </Table.ColumnHeaderCell>
          <Table.ColumnHeaderCell className="w-20 text-xs uppercase tracking-wider">
            Status
          </Table.ColumnHeaderCell>
          <Table.ColumnHeaderCell className="w-20 text-xs uppercase tracking-wider">
            Method
          </Table.ColumnHeaderCell>
          <Table.ColumnHeaderCell className="text-xs uppercase tracking-wider">
            Path
          </Table.ColumnHeaderCell>
          <Table.ColumnHeaderCell className="w-32 text-xs uppercase tracking-wider">
            Order ref
          </Table.ColumnHeaderCell>
          <Table.ColumnHeaderCell className="w-28 text-xs uppercase tracking-wider">
            Gateway
          </Table.ColumnHeaderCell>
          <Table.ColumnHeaderCell className="w-32 text-xs uppercase tracking-wider">
            Scenario
          </Table.ColumnHeaderCell>
          <Table.ColumnHeaderCell className="w-20 text-right text-xs uppercase tracking-wider">
            Duration
          </Table.ColumnHeaderCell>
        </Table.Row>
      </Table.Header>
      {loading ? (
        <TableSkeleton rows={8} columns={9} />
      ) : (
        <Table.Body key={nonce}>
          {rows.map((l) => {
            const isSelected = selectedId === l.id;
            const scenarioName = l.scenario_id ? scenarioNames[l.scenario_id] : null;
            return (
              <Table.Row
                key={l.id}
                className={`row-accent cursor-pointer font-mono text-xs transition-colors animate-in fade-in duration-200 ${isSelected ? 'bg-[var(--accent-a3)]' : ''}`}
                onClick={() => onSelect(l.id)}
              >
                <Table.Cell className="text-[var(--gray-11)]" title={l.id}>
                  {l.id.slice(0, 8)}
                </Table.Cell>
                <Table.Cell className="text-muted-foreground">
                  <span title={new Date(l.created_at).toLocaleString()}>
                    {shortTime(l.created_at)}
                  </span>
                </Table.Cell>
                <Table.Cell>
                  <StatusChip status={l.response_status} />
                </Table.Cell>
                <Table.Cell className="text-muted-foreground font-semibold">{l.method}</Table.Cell>
                <Table.Cell className="truncate max-w-xs">{l.path}</Table.Cell>
                <Table.Cell className="truncate" title={l.merchant_order_id || '—'}>
                  {l.merchant_order_id ? (
                    <span>{l.merchant_order_id}</span>
                  ) : (
                    <span className="text-[var(--gray-9)]">—</span>
                  )}
                </Table.Cell>
                <Table.Cell>
                  <GatewayBadge gateway={l.gateway} />
                </Table.Cell>
                <Table.Cell
                  className="truncate text-[var(--gray-11)]"
                  title={
                    l.scenario_id
                      ? `${scenarioName ?? '(unknown)'} — ${l.scenario_id}`
                      : 'No scenario — used built-in fallback'
                  }
                >
                  {scenarioName ?? (l.scenario_id ? l.scenario_id.slice(0, 8) : <span className="text-[var(--gray-9)]">—</span>)}
                </Table.Cell>
                <Table.Cell className="text-right tabular-nums text-muted-foreground">
                  {l.duration_ms}ms
                </Table.Cell>
              </Table.Row>
            );
          })}
          {rows.length === 0 && (
            <Table.Row>
              <Table.Cell colSpan={9} className="text-center text-muted-foreground py-12">
                <div className="text-sm">No logs yet</div>
                <div className="text-xs mt-1">
                  Point your app at a mock endpoint and requests will appear here.
                </div>
              </Table.Cell>
            </Table.Row>
          )}
        </Table.Body>
      )}
    </Table.Root>
  );
}
