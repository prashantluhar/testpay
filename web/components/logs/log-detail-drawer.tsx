'use client';
import { Button, Flex, Separator, Tabs, Text } from '@radix-ui/themes';
import { toast } from 'sonner';
import { StatusChip } from '@/components/common/status-chip';
import { GatewayBadge } from '@/components/common/gateway-badge';
import { JsonViewer, KeyValueGrid } from '@/components/common/json-viewer';
import { Spinner } from '@/components/common/spinner';
import { Sheet, SheetContent, SheetTitle } from '@/components/common/sheet';
import { useLog } from '@/lib/hooks';
import { api } from '@/lib/api';

export function LogDetailDrawer({ id, onClose }: { id: string | null; onClose: () => void }) {
  const { data } = useLog(id);
  const open = !!id;

  async function replay() {
    if (!id) return;
    try {
      await api(`/api/logs/${id}/replay`, { method: 'POST' });
      toast.success('Replayed');
    } catch {
      toast.error('Replay failed');
    }
  }

  return (
    <Sheet open={open} onOpenChange={(o) => !o && onClose()}>
      <SheetContent width={760}>
        <SheetTitle asChild>
          <Flex align="center" gap="2" wrap="wrap" className="pr-10">
            <Text size="5" weight="bold">
              Request detail
            </Text>
            {data?.request && <StatusChip status={data.request.response_status} />}
            {data?.request && <GatewayBadge gateway={data.request.gateway} />}
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
                { label: 'Method', value: data.request.method },
                { label: 'Path', value: data.request.path },
                { label: 'Duration', value: `${data.request.duration_ms} ms` },
                { label: 'Client IP', value: data.request.client_ip || '—' },
              ]}
            />

            <Flex gap="2">
              <Button size="2" onClick={replay}>
                Replay this request
              </Button>
            </Flex>

            <Separator size="4" />

            <Tabs.Root defaultValue="request">
              <Tabs.List size="2">
                <Tabs.Trigger value="request">Request</Tabs.Trigger>
                <Tabs.Trigger value="response">Response</Tabs.Trigger>
                <Tabs.Trigger value="webhook">Webhook</Tabs.Trigger>
              </Tabs.List>
              <Tabs.Content value="request" className="pt-4">
                <Flex direction="column" gap="4">
                  <Section label="Headers">
                    <JsonViewer value={data.request.request_headers} />
                  </Section>
                  <Section label="Body">
                    <JsonViewer value={data.request.request_body} />
                  </Section>
                </Flex>
              </Tabs.Content>
              <Tabs.Content value="response" className="pt-4">
                <Flex direction="column" gap="4">
                  <Section label="Headers">
                    <JsonViewer value={data.request.response_headers} />
                  </Section>
                  <Section label="Body">
                    <JsonViewer value={data.request.response_body} />
                  </Section>
                </Flex>
              </Tabs.Content>
              <Tabs.Content value="webhook" className="pt-4">
                <Section label={data.webhook ? 'Webhook payload' : 'Webhook'}>
                  <JsonViewer value={data.webhook ?? { note: 'no webhook for this request' }} />
                </Section>
              </Tabs.Content>
            </Tabs.Root>
          </Flex>
        )}
      </SheetContent>
    </Sheet>
  );
}

function Section({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div>
      <div className="text-[11px] uppercase tracking-wider text-[var(--gray-11)] mb-2">{label}</div>
      {children}
    </div>
  );
}
