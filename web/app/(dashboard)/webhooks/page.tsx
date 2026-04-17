'use client';
import { useState } from 'react';
import { Badge, Dialog, Flex, Heading, Table, Tabs, Text } from '@radix-ui/themes';
import { JsonViewer } from '@/components/common/json-viewer';
import { StatusChip } from '@/components/common/status-chip';
import { useWebhooks, useWebhook } from '@/lib/hooks';
import type { WebhookLog } from '@/lib/types';

// Short relative-time like "4s", "2m", "1h", else an absolute short timestamp.
function shortTime(iso: string) {
  const then = new Date(iso).getTime();
  const diff = Date.now() - then;
  if (diff < 60_000) return `${Math.max(1, Math.floor(diff / 1000))}s ago`;
  if (diff < 3_600_000) return `${Math.floor(diff / 60_000)}m ago`;
  if (diff < 86_400_000) return `${Math.floor(diff / 3_600_000)}h ago`;
  return new Date(iso).toLocaleDateString();
}

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
      <Flex align="end" justify="between" mb="3" className="shrink-0">
        <div>
          <Heading size="6">Webhooks</Heading>
          <Text size="2" color="gray" as="p">
            Outbound webhook deliveries. Polls every 3s.
          </Text>
        </div>
        <Text size="1" color="gray">
          {data?.length ?? 0} rows
        </Text>
      </Flex>
      <div className="flex-1 min-h-0 overflow-y-auto border rounded-md">
        <Table.Root size="1">
          <Table.Header className="sticky top-0 bg-card/95 backdrop-blur z-10">
            <Table.Row>
              <Table.ColumnHeaderCell className="w-36 text-xs uppercase tracking-wider">
                Time
              </Table.ColumnHeaderCell>
              <Table.ColumnHeaderCell className="w-24 text-xs uppercase tracking-wider">
                Status
              </Table.ColumnHeaderCell>
              <Table.ColumnHeaderCell className="w-20 text-xs uppercase tracking-wider text-center">
                Attempts
              </Table.ColumnHeaderCell>
              <Table.ColumnHeaderCell className="text-xs uppercase tracking-wider">
                Target URL
              </Table.ColumnHeaderCell>
              <Table.ColumnHeaderCell className="w-32 text-xs uppercase tracking-wider">
                Request ID
              </Table.ColumnHeaderCell>
            </Table.Row>
          </Table.Header>
          <Table.Body>
            {data?.map((w) => (
              <Table.Row
                key={w.id}
                className="row-accent cursor-pointer font-mono text-xs transition-colors"
                onClick={() => setSelected(w.id)}
              >
                <Table.Cell className="text-muted-foreground">
                  <span title={new Date(w.created_at).toLocaleString()}>
                    {shortTime(w.created_at)}
                  </span>
                </Table.Cell>
                <Table.Cell>
                  <span
                    className={`inline-flex items-center gap-1 rounded px-2 py-0.5 text-xs ${statusVariant(w.delivery_status)}`}
                  >
                    {w.delivery_status === 'delivered' && (
                      <span className="h-1.5 w-1.5 rounded-full bg-emerald-500" />
                    )}
                    {w.delivery_status === 'failed' && (
                      <span className="h-1.5 w-1.5 rounded-full bg-red-500" />
                    )}
                    {w.delivery_status}
                  </span>
                </Table.Cell>
                <Table.Cell className="text-center">
                  <Badge variant="outline" color="gray" className="tabular-nums">
                    {w.attempts}
                  </Badge>
                </Table.Cell>
                <Table.Cell className="truncate max-w-sm">{w.target_url}</Table.Cell>
                <Table.Cell className="text-muted-foreground truncate">
                  {w.request_log_id.slice(0, 8)}
                </Table.Cell>
              </Table.Row>
            ))}
            {data && data.length === 0 && (
              <Table.Row>
                <Table.Cell colSpan={5} className="text-center text-muted-foreground py-12">
                  <div className="text-sm">No webhooks yet</div>
                  <div className="text-xs mt-1">
                    Configure URLs in Settings, then fire a mock charge.
                  </div>
                </Table.Cell>
              </Table.Row>
            )}
          </Table.Body>
        </Table.Root>
      </div>
      <WebhookDetailDrawer id={selected} onClose={() => setSelected(null)} />
    </div>
  );
}

function WebhookDetailDrawer({ id, onClose }: { id: string | null; onClose: () => void }) {
  const { data } = useWebhook(id);
  const open = !!id;
  return (
    <Dialog.Root open={open} onOpenChange={(o) => !o && onClose()}>
      <Dialog.Content maxWidth="700px">
        <Dialog.Title>
          <Flex align="center" gap="2">
            Webhook delivery
            {data && (
              <span
                className={`inline-flex rounded px-2 py-0.5 text-xs ${statusVariant(data.delivery_status)}`}
              >
                {data.delivery_status}
              </span>
            )}
          </Flex>
        </Dialog.Title>
        {!data ? (
          <Text size="2" color="gray" as="p" mt="4">
            Loading…
          </Text>
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

            <Tabs.Root defaultValue="attempts">
              <Tabs.List>
                <Tabs.Trigger value="attempts">Attempts</Tabs.Trigger>
                <Tabs.Trigger value="payload">Payload</Tabs.Trigger>
              </Tabs.List>
              <Tabs.Content value="attempts" className="space-y-2 pt-3">
                <AttemptsTable attempts={data.attempt_logs ?? []} />
              </Tabs.Content>
              <Tabs.Content value="payload" className="pt-3">
                <JsonViewer value={data.payload} />
              </Tabs.Content>
            </Tabs.Root>
          </div>
        )}
      </Dialog.Content>
    </Dialog.Root>
  );
}

function AttemptsTable({ attempts }: { attempts: WebhookLog['attempt_logs'] }) {
  if (!attempts || attempts.length === 0) {
    return (
      <Text size="1" color="gray">
        No attempt details recorded.
      </Text>
    );
  }
  return (
    <div className="border rounded-md">
      <Table.Root size="1">
        <Table.Header>
          <Table.Row>
            <Table.ColumnHeaderCell className="w-12">#</Table.ColumnHeaderCell>
            <Table.ColumnHeaderCell className="w-24">Status</Table.ColumnHeaderCell>
            <Table.ColumnHeaderCell className="w-28">Duration</Table.ColumnHeaderCell>
            <Table.ColumnHeaderCell>Attempted at</Table.ColumnHeaderCell>
          </Table.Row>
        </Table.Header>
        <Table.Body>
          {attempts.map((a, i) => (
            <Table.Row key={i} className="font-mono text-xs">
              <Table.Cell className="text-muted-foreground">#{i + 1}</Table.Cell>
              <Table.Cell>
                {a.status > 0 ? (
                  <StatusChip status={a.status} />
                ) : (
                  <span className="text-red-500">network</span>
                )}
              </Table.Cell>
              <Table.Cell>{a.duration_ms}ms</Table.Cell>
              <Table.Cell className="text-muted-foreground">{a.attempted_at}</Table.Cell>
            </Table.Row>
          ))}
        </Table.Body>
      </Table.Root>
    </div>
  );
}
