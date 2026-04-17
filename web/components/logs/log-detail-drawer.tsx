'use client';
import { Button, Dialog, Flex, Tabs, Text } from '@radix-ui/themes';
import { toast } from 'sonner';
import { StatusChip } from '@/components/common/status-chip';
import { GatewayBadge } from '@/components/common/gateway-badge';
import { JsonViewer } from '@/components/common/json-viewer';
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
    <Dialog.Root open={open} onOpenChange={(o) => !o && onClose()}>
      <Dialog.Content maxWidth="640px">
        <Dialog.Title>
          <Flex align="center" gap="2">
            Request detail
            {data?.request && <StatusChip status={data.request.response_status} />}
            {data?.request && <GatewayBadge gateway={data.request.gateway} />}
          </Flex>
        </Dialog.Title>
        {!data ? (
          <Text size="2" color="gray" as="p" mt="4">
            Loading…
          </Text>
        ) : (
          <Flex direction="column" gap="4" mt="4">
            <Flex gap="2">
              <Button size="2" onClick={replay}>
                Replay
              </Button>
            </Flex>
            <Tabs.Root defaultValue="request">
              <Tabs.List>
                <Tabs.Trigger value="request">Request</Tabs.Trigger>
                <Tabs.Trigger value="response">Response</Tabs.Trigger>
                <Tabs.Trigger value="webhook">Webhook</Tabs.Trigger>
              </Tabs.List>
              <Tabs.Content value="request" className="space-y-2 pt-3">
                <div className="text-xs text-muted-foreground">Headers</div>
                <JsonViewer value={data.request.request_headers} />
                <div className="text-xs text-muted-foreground mt-2">Body</div>
                <JsonViewer value={data.request.request_body} />
              </Tabs.Content>
              <Tabs.Content value="response" className="space-y-2 pt-3">
                <div className="text-xs text-muted-foreground">Headers</div>
                <JsonViewer value={data.request.response_headers} />
                <div className="text-xs text-muted-foreground mt-2">Body</div>
                <JsonViewer value={data.request.response_body} />
              </Tabs.Content>
              <Tabs.Content value="webhook" className="pt-3">
                <JsonViewer value={data.webhook ?? { note: 'no webhook for this request' }} />
              </Tabs.Content>
            </Tabs.Root>
          </Flex>
        )}
      </Dialog.Content>
    </Dialog.Root>
  );
}
