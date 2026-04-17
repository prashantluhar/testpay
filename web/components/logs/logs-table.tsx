'use client';
import { Table } from '@radix-ui/themes';
import { StatusChip } from '@/components/common/status-chip';
import { GatewayBadge } from '@/components/common/gateway-badge';
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
}: {
  rows: RequestLog[];
  onSelect: (id: string) => void;
}) {
  return (
    <Table.Root size="1">
      <Table.Header className="sticky top-0 bg-card/95 backdrop-blur z-10">
        <Table.Row>
          <Table.ColumnHeaderCell className="w-36 text-xs uppercase tracking-wider">
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
          <Table.ColumnHeaderCell className="w-28 text-xs uppercase tracking-wider">
            Gateway
          </Table.ColumnHeaderCell>
          <Table.ColumnHeaderCell className="w-24 text-right text-xs uppercase tracking-wider">
            Duration
          </Table.ColumnHeaderCell>
        </Table.Row>
      </Table.Header>
      <Table.Body>
        {rows.map((l) => (
          <Table.Row
            key={l.id}
            className="row-accent cursor-pointer font-mono text-xs transition-colors"
            onClick={() => onSelect(l.id)}
          >
            <Table.Cell className="text-muted-foreground">
              <span title={new Date(l.created_at).toLocaleString()}>{shortTime(l.created_at)}</span>
            </Table.Cell>
            <Table.Cell>
              <StatusChip status={l.response_status} />
            </Table.Cell>
            <Table.Cell className="text-muted-foreground font-semibold">{l.method}</Table.Cell>
            <Table.Cell className="truncate max-w-xs">{l.path}</Table.Cell>
            <Table.Cell>
              <GatewayBadge gateway={l.gateway} />
            </Table.Cell>
            <Table.Cell className="text-right tabular-nums text-muted-foreground">
              {l.duration_ms}ms
            </Table.Cell>
          </Table.Row>
        ))}
        {rows.length === 0 && (
          <Table.Row>
            <Table.Cell colSpan={6} className="text-center text-muted-foreground py-12">
              <div className="text-sm">No logs yet</div>
              <div className="text-xs mt-1">
                Point your app at a mock endpoint and requests will appear here.
              </div>
            </Table.Cell>
          </Table.Row>
        )}
      </Table.Body>
    </Table.Root>
  );
}
