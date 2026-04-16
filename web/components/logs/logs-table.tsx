'use client';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
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
    <Table>
      <TableHeader className="sticky top-0 bg-card/95 backdrop-blur z-10">
        <TableRow className="border-b">
          <TableHead className="w-36 text-xs uppercase tracking-wider">Time</TableHead>
          <TableHead className="w-20 text-xs uppercase tracking-wider">Status</TableHead>
          <TableHead className="w-20 text-xs uppercase tracking-wider">Method</TableHead>
          <TableHead className="text-xs uppercase tracking-wider">Path</TableHead>
          <TableHead className="w-28 text-xs uppercase tracking-wider">Gateway</TableHead>
          <TableHead className="w-24 text-right text-xs uppercase tracking-wider">Duration</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {rows.map((l) => (
          <TableRow
            key={l.id}
            className="row-accent cursor-pointer font-mono text-xs border-0 transition-colors"
            onClick={() => onSelect(l.id)}
          >
            <TableCell className="text-muted-foreground">
              <span title={new Date(l.created_at).toLocaleString()}>{shortTime(l.created_at)}</span>
            </TableCell>
            <TableCell>
              <StatusChip status={l.response_status} />
            </TableCell>
            <TableCell className="text-muted-foreground font-semibold">{l.method}</TableCell>
            <TableCell className="truncate max-w-xs">{l.path}</TableCell>
            <TableCell>
              <GatewayBadge gateway={l.gateway} />
            </TableCell>
            <TableCell className="text-right tabular-nums text-muted-foreground">
              {l.duration_ms}ms
            </TableCell>
          </TableRow>
        ))}
        {rows.length === 0 && (
          <TableRow>
            <TableCell colSpan={6} className="text-center text-muted-foreground py-12">
              <div className="text-sm">No logs yet</div>
              <div className="text-xs mt-1">
                Point your app at a mock endpoint and requests will appear here.
              </div>
            </TableCell>
          </TableRow>
        )}
      </TableBody>
    </Table>
  );
}
