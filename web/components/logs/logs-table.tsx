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

export function LogsTable({
  rows,
  onSelect,
}: {
  rows: RequestLog[];
  onSelect: (id: string) => void;
}) {
  return (
    <div className="border rounded-md">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead className="w-40">Time</TableHead>
            <TableHead className="w-16">Status</TableHead>
            <TableHead className="w-20">Method</TableHead>
            <TableHead>Path</TableHead>
            <TableHead className="w-24">Gateway</TableHead>
            <TableHead className="w-20 text-right">Duration</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {rows.map((l) => (
            <TableRow
              key={l.id}
              className="cursor-pointer font-mono text-xs"
              onClick={() => onSelect(l.id)}
            >
              <TableCell className="text-muted-foreground">
                {new Date(l.created_at).toLocaleString()}
              </TableCell>
              <TableCell>
                <StatusChip status={l.response_status} />
              </TableCell>
              <TableCell className="text-muted-foreground">{l.method}</TableCell>
              <TableCell className="truncate max-w-xs">{l.path}</TableCell>
              <TableCell>
                <GatewayBadge gateway={l.gateway} />
              </TableCell>
              <TableCell className="text-right text-muted-foreground">{l.duration_ms}ms</TableCell>
            </TableRow>
          ))}
          {rows.length === 0 && (
            <TableRow>
              <TableCell colSpan={6} className="text-center text-muted-foreground py-8">
                No logs yet.
              </TableCell>
            </TableRow>
          )}
        </TableBody>
      </Table>
    </div>
  );
}
