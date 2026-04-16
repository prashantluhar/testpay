'use client';
import { useState } from 'react';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { Badge } from '@/components/ui/badge';
import { Sheet, SheetContent, SheetHeader, SheetTitle } from '@/components/ui/sheet';
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs';
import { JsonViewer } from '@/components/common/json-viewer';
import { StatusChip } from '@/components/common/status-chip';
import { useWebhooks, useWebhook } from '@/lib/hooks';
import type { WebhookLog } from '@/lib/types';

function statusVariant(status: string) {
  switch (status) {
    case 'delivered':
      return 'bg-emerald-500/10 text-emerald-500';
    case 'failed':
      return 'bg-red-500/10 text-red-500';
    case 'duplicate':
      return 'bg-amber-500/10 text-amber-500';
    case 'pending':
    default:
      return 'bg-muted text-muted-foreground';
  }
}

export default function WebhooksPage() {
  const { data } = useWebhooks({ limit: 200, pollInterval: 3000 });
  const [selected, setSelected] = useState<string | null>(null);

  return (
    <div className="flex flex-col h-[calc(100vh-6rem)]">
      <div className="flex items-end justify-between mb-3 shrink-0">
        <div>
          <h1 className="text-2xl font-semibold">Webhooks</h1>
          <p className="text-sm text-muted-foreground">
            Outbound webhook deliveries. Polls every 3s.
          </p>
        </div>
        <span className="text-xs text-muted-foreground">{data?.length ?? 0} rows</span>
      </div>
      <div className="flex-1 min-h-0 overflow-y-auto border rounded-md">
        <Table>
          <TableHeader className="sticky top-0 bg-card z-10">
            <TableRow>
              <TableHead className="w-40">Time</TableHead>
              <TableHead className="w-24">Status</TableHead>
              <TableHead className="w-16">Attempts</TableHead>
              <TableHead>Target URL</TableHead>
              <TableHead className="w-40">Request ID</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {data?.map((w) => (
              <TableRow
                key={w.id}
                className="cursor-pointer font-mono text-xs"
                onClick={() => setSelected(w.id)}
              >
                <TableCell className="text-muted-foreground">
                  {new Date(w.created_at).toLocaleString()}
                </TableCell>
                <TableCell>
                  <span className={`inline-flex rounded px-2 py-0.5 text-xs ${statusVariant(w.delivery_status)}`}>
                    {w.delivery_status}
                  </span>
                </TableCell>
                <TableCell>
                  <Badge variant="outline">{w.attempts}</Badge>
                </TableCell>
                <TableCell className="truncate max-w-sm">{w.target_url}</TableCell>
                <TableCell className="text-muted-foreground truncate">
                  {w.request_log_id.slice(0, 8)}
                </TableCell>
              </TableRow>
            ))}
            {data && data.length === 0 && (
              <TableRow>
                <TableCell colSpan={5} className="text-center text-muted-foreground py-8">
                  No webhooks yet. Configure webhook URLs in Settings and fire a mock charge.
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>
      <WebhookDetailDrawer id={selected} onClose={() => setSelected(null)} />
    </div>
  );
}



function WebhookDetailDrawer({ id, onClose }: { id: string | null; onClose: () => void }) {
  const { data } = useWebhook(id);
  const open = !!id;
  return (
    <Sheet open={open} onOpenChange={(o) => !o && onClose()}>
      <SheetContent className="w-[700px] sm:max-w-[700px]">
        <SheetHeader>
          <SheetTitle className="flex items-center gap-2">
            Webhook delivery
            {data && (
              <span
                className={`inline-flex rounded px-2 py-0.5 text-xs ${statusVariant(data.delivery_status)}`}
              >
                {data.delivery_status}
              </span>
            )}
          </SheetTitle>
        </SheetHeader>
        {!data ? (
          <div className="py-6 text-sm text-muted-foreground">Loading…</div>
        ) : (
          <div className="mt-4 space-y-4">
            <div className="grid grid-cols-2 gap-3 text-xs font-mono">
              <div>
                <div className="text-muted-foreground uppercase tracking-wider text-[10px]">
                  Target URL
                </div>
                <div className="break-all">{data.target_url}</div>
              </div>
              <div>
                <div className="text-muted-foreground uppercase tracking-wider text-[10px]">
                  Attempts
                </div>
                <div>{data.attempts}</div>
              </div>
              <div>
                <div className="text-muted-foreground uppercase tracking-wider text-[10px]">
                  Request log id
                </div>
                <div className="break-all">{data.request_log_id}</div>
              </div>
              <div>
                <div className="text-muted-foreground uppercase tracking-wider text-[10px]">
                  Delivered at
                </div>
                <div>{data.delivered_at ?? '—'}</div>
              </div>
            </div>

            <Tabs defaultValue="attempts">
              <TabsList>
                <TabsTrigger value="attempts">Attempts</TabsTrigger>
                <TabsTrigger value="payload">Payload</TabsTrigger>
              </TabsList>
              <TabsContent value="attempts" className="space-y-2">
                <AttemptsTable attempts={data.attempt_logs ?? []} />
              </TabsContent>
              <TabsContent value="payload">
                <JsonViewer value={data.payload} />
              </TabsContent>
            </Tabs>
          </div>
        )}
      </SheetContent>
    </Sheet>
  );
}

function AttemptsTable({ attempts }: { attempts: WebhookLog['attempt_logs'] }) {
  if (!attempts || attempts.length === 0) {
    return <div className="text-xs text-muted-foreground">No attempt details recorded.</div>;
  }
  return (
    <div className="border rounded-md">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead className="w-12">#</TableHead>
            <TableHead className="w-24">Status</TableHead>
            <TableHead className="w-28">Duration</TableHead>
            <TableHead>Attempted at</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {attempts.map((a, i) => (
            <TableRow key={i} className="font-mono text-xs">
              <TableCell className="text-muted-foreground">#{i + 1}</TableCell>
              <TableCell>
                {a.status > 0 ? <StatusChip status={a.status} /> : <span className="text-red-500">network</span>}
              </TableCell>
              <TableCell>{a.duration_ms}ms</TableCell>
              <TableCell className="text-muted-foreground">{a.attempted_at}</TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  );
}
