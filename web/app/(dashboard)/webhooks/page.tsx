'use client';
import { Fragment, useState } from 'react';
import { ChevronDownIcon } from '@radix-ui/react-icons';
import { Badge, Flex, Heading, Separator, Table, Tabs, Text } from '@radix-ui/themes';
import { JsonViewer, KeyValueGrid } from '@/components/common/json-viewer';
import { StatusChip } from '@/components/common/status-chip';
import { TableSkeleton } from '@/components/common/table-skeleton';
import { Spinner } from '@/components/common/spinner';
import { Sheet, SheetContent, SheetTitle } from '@/components/common/sheet';
import { useWebhooks, useWebhook, useLog } from '@/lib/hooks';
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
    <div className="flex flex-col h-[calc(100vh-6rem)] animate-in fade-in duration-300">
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
          {data === undefined ? (
            <TableSkeleton rows={8} columns={5} />
          ) : (
          <Table.Body>
            {data?.map((w) => (
              <Table.Row
                key={w.id}
                className="row-accent cursor-pointer font-mono text-xs transition-colors animate-in fade-in duration-200"
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
          )}
        </Table.Root>
      </div>
      <WebhookDetailDrawer id={selected} onClose={() => setSelected(null)} />
    </div>
  );
}

function WebhookDetailDrawer({ id, onClose }: { id: string | null; onClose: () => void }) {
  const { data } = useWebhook(id);
  // Pull in the originating request so the drawer can surface request/response
  // alongside the webhook — saves the user from bouncing to the Logs tab.
  const { data: origin } = useLog(data?.request_log_id ?? null);
  const open = !!id;
  return (
    <Sheet open={open} onOpenChange={(o) => !o && onClose()}>
      <SheetContent width={780}>
        <SheetTitle asChild>
          <Flex align="center" gap="2" wrap="wrap" className="pr-10">
            <Text size="5" weight="bold">
              Webhook delivery
            </Text>
            {data && (
              <span
                className={`inline-flex rounded px-2 py-0.5 text-xs font-medium ${statusVariant(data.delivery_status)}`}
              >
                {data.delivery_status}
              </span>
            )}
          </Flex>
        </SheetTitle>
        {!data ? (
          <Flex align="center" gap="2" mt="5">
            <Spinner size="small" />
            <Text size="2" color="gray">
              Loading…
            </Text>
          </Flex>
        ) : (
          <Flex direction="column" gap="5" mt="5">
            <KeyValueGrid
              items={[
                { label: 'Target URL', value: data.target_url },
                { label: 'Attempts', value: String(data.attempts) },
                {
                  label: 'Delivered at',
                  value: data.delivered_at
                    ? new Date(data.delivered_at).toLocaleString()
                    : '—',
                },
                { label: 'Request log', value: data.request_log_id },
              ]}
            />

            <Separator size="4" />

            <Tabs.Root defaultValue="payload">
              <Tabs.List size="2">
                <Tabs.Trigger value="payload">Webhook payload</Tabs.Trigger>
                <Tabs.Trigger value="attempts">Attempts ({data.attempts})</Tabs.Trigger>
                <Tabs.Trigger value="request">Request</Tabs.Trigger>
                <Tabs.Trigger value="response">Response</Tabs.Trigger>
              </Tabs.List>
              <Tabs.Content value="payload" className="pt-4">
                <JsonViewer value={data.payload} />
              </Tabs.Content>
              <Tabs.Content value="attempts" className="pt-4">
                <AttemptsTable attempts={data.attempt_logs ?? []} />
              </Tabs.Content>
              <Tabs.Content value="request" className="pt-4">
                {origin?.request ? (
                  <Flex direction="column" gap="4">
                    <LinkedHeader
                      label={`${origin.request.method} ${origin.request.path}`}
                      statusLabel={origin.request.gateway}
                    />
                    <SectionLabel label="Headers" />
                    <JsonViewer value={origin.request.request_headers} />
                    <SectionLabel label="Body" />
                    <JsonViewer value={origin.request.request_body} />
                  </Flex>
                ) : (
                  <InlineSpinner label="Loading originating request…" />
                )}
              </Tabs.Content>
              <Tabs.Content value="response" className="pt-4">
                {origin?.request ? (
                  <Flex direction="column" gap="4">
                    <LinkedHeader
                      label={`${origin.request.response_status}`}
                      statusLabel={`${origin.request.duration_ms} ms`}
                    />
                    <SectionLabel label="Headers" />
                    <JsonViewer value={origin.request.response_headers} />
                    <SectionLabel label="Body" />
                    <JsonViewer value={origin.request.response_body} />
                  </Flex>
                ) : (
                  <InlineSpinner label="Loading originating response…" />
                )}
              </Tabs.Content>
            </Tabs.Root>
          </Flex>
        )}
      </SheetContent>
    </Sheet>
  );
}

function SectionLabel({ label }: { label: string }) {
  return <div className="text-[11px] uppercase tracking-wider text-[var(--gray-11)]">{label}</div>;
}

function LinkedHeader({ label, statusLabel }: { label: string; statusLabel: string }) {
  return (
    <Flex align="center" gap="3" wrap="wrap">
      <Text size="2" weight="medium" className="font-mono">
        {label}
      </Text>
      <Text size="1" color="gray" className="font-mono">
        {statusLabel}
      </Text>
    </Flex>
  );
}

function InlineSpinner({ label }: { label: string }) {
  return (
    <Flex align="center" gap="2">
      <Spinner size="small" />
      <Text size="2" color="gray">
        {label}
      </Text>
    </Flex>
  );
}

function AttemptsTable({ attempts }: { attempts: WebhookLog['attempt_logs'] }) {
  const [expanded, setExpanded] = useState<number | null>(null);
  if (!attempts || attempts.length === 0) {
    return (
      <Text size="2" color="gray">
        No attempt details recorded.
      </Text>
    );
  }
  return (
    <div className="border border-[var(--gray-a5)] rounded-md overflow-hidden">
      <Table.Root size="2">
        <Table.Header>
          <Table.Row>
            <Table.ColumnHeaderCell className="w-12">#</Table.ColumnHeaderCell>
            <Table.ColumnHeaderCell className="w-24">Status</Table.ColumnHeaderCell>
            <Table.ColumnHeaderCell className="w-28">Duration</Table.ColumnHeaderCell>
            <Table.ColumnHeaderCell>Attempted at</Table.ColumnHeaderCell>
            <Table.ColumnHeaderCell className="w-10" />
          </Table.Row>
        </Table.Header>
        <Table.Body>
          {attempts.map((a, i) => {
            const isOpen = expanded === i;
            const hasDetail = a.response || a.error;
            return (
              <Fragment key={i}>
                <Table.Row
                  onClick={() => hasDetail && setExpanded(isOpen ? null : i)}
                  className={`font-mono text-sm ${hasDetail ? 'cursor-pointer hover:bg-[var(--gray-a3)] transition-colors' : ''}`}
                >
                  <Table.Cell className="text-[var(--gray-11)]">#{i + 1}</Table.Cell>
                  <Table.Cell>
                    {a.status > 0 ? (
                      <StatusChip status={a.status} />
                    ) : (
                      <span className="text-red-500">network</span>
                    )}
                  </Table.Cell>
                  <Table.Cell>{a.duration_ms} ms</Table.Cell>
                  <Table.Cell className="text-[var(--gray-11)]">{a.attempted_at}</Table.Cell>
                  <Table.Cell>
                    {hasDetail && (
                      <ChevronDownIcon
                        className={`h-4 w-4 text-[var(--gray-11)] transition-transform duration-200 ${isOpen ? 'rotate-180' : ''}`}
                      />
                    )}
                  </Table.Cell>
                </Table.Row>
                {isOpen && (
                  <Table.Row>
                    <Table.Cell colSpan={5} className="bg-[var(--gray-a2)] p-0">
                      <div className="p-4 space-y-3 animate-in fade-in duration-200">
                        {a.error && (
                          <div>
                            <div className="text-[11px] uppercase tracking-wider text-[var(--red-11)] mb-1">
                              Transport error
                            </div>
                            <pre className="font-mono text-[13px] leading-relaxed text-[var(--red-12)] bg-[var(--red-a3)] border border-[var(--red-a5)] p-3 rounded-md whitespace-pre-wrap break-words">
                              {a.error}
                            </pre>
                          </div>
                        )}
                        {a.response ? (
                          <div>
                            <div className="text-[11px] uppercase tracking-wider text-[var(--gray-11)] mb-1">
                              Response body
                            </div>
                            <pre className="font-mono text-[13px] leading-relaxed text-[var(--gray-12)] bg-[var(--gray-a3)] border border-[var(--gray-a5)] p-3 rounded-md overflow-auto max-h-[320px] whitespace-pre-wrap break-words">
                              {a.response}
                            </pre>
                          </div>
                        ) : (
                          !a.error && (
                            <Text size="2" color="gray">
                              (empty response body)
                            </Text>
                          )
                        )}
                      </div>
                    </Table.Cell>
                  </Table.Row>
                )}
              </Fragment>
            );
          })}
        </Table.Body>
      </Table.Root>
    </div>
  );
}
