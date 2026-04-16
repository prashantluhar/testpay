'use client';
import { Sheet, SheetContent, SheetHeader, SheetTitle } from '@/components/ui/sheet';
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs';
import { Button } from '@/components/ui/button';
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
    <Sheet open={open} onOpenChange={(o) => !o && onClose()}>
      <SheetContent className="w-[640px] sm:max-w-[640px]">
        <SheetHeader>
          <SheetTitle className="flex items-center gap-2">
            Request detail
            {data?.request && <StatusChip status={data.request.response_status} />}
            {data?.request && <GatewayBadge gateway={data.request.gateway} />}
          </SheetTitle>
        </SheetHeader>
        {!data ? (
          <div className="py-6 text-sm text-muted-foreground">Loading…</div>
        ) : (
          <div className="mt-4 space-y-4">
            <div className="flex gap-2">
              <Button size="sm" onClick={replay}>
                Replay
              </Button>
            </div>
            <Tabs defaultValue="request">
              <TabsList>
                <TabsTrigger value="request">Request</TabsTrigger>
                <TabsTrigger value="response">Response</TabsTrigger>
                <TabsTrigger value="webhook">Webhook</TabsTrigger>
              </TabsList>
              <TabsContent value="request" className="space-y-2">
                <div className="text-xs text-muted-foreground">Headers</div>
                <JsonViewer value={data.request.request_headers} />
                <div className="text-xs text-muted-foreground mt-2">Body</div>
                <JsonViewer value={data.request.request_body} />
              </TabsContent>
              <TabsContent value="response" className="space-y-2">
                <div className="text-xs text-muted-foreground">Headers</div>
                <JsonViewer value={data.request.response_headers} />
                <div className="text-xs text-muted-foreground mt-2">Body</div>
                <JsonViewer value={data.request.response_body} />
              </TabsContent>
              <TabsContent value="webhook">
                <JsonViewer value={data.webhook ?? { note: 'no webhook for this request' }} />
              </TabsContent>
            </Tabs>
          </div>
        )}
      </SheetContent>
    </Sheet>
  );
}
