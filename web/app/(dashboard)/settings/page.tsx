'use client';
import { useEffect, useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Label } from '@/components/ui/label';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { toast } from 'sonner';
import { useMe } from '@/lib/hooks';
import { ApiKeyReveal } from '@/components/common/api-key-reveal';
import { useTheme } from '@/components/common/theme-provider';
import { CopyButton } from '@/components/common/copy-button';
import { MODE } from '@/lib/types';
import { api, ApiError } from '@/lib/api';
import { mutate } from 'swr';

const GATEWAYS = ['stripe', 'razorpay', 'agnostic'] as const;
type Gateway = (typeof GATEWAYS)[number];

export default function SettingsPage() {
  const { data } = useMe();
  const { theme, setTheme } = useTheme();
  const [urls, setUrls] = useState<Record<Gateway, string>>({
    stripe: '',
    razorpay: '',
    agnostic: '',
  });
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    if (data?.workspace?.webhook_urls) {
      setUrls({
        stripe: data.workspace.webhook_urls.stripe ?? '',
        razorpay: data.workspace.webhook_urls.razorpay ?? '',
        agnostic: data.workspace.webhook_urls.agnostic ?? '',
      });
    }
  }, [data?.workspace?.webhook_urls]);

  if (!data) return null;
  const { workspace, user } = data;

  const baseUrl =
    MODE === 'local' ? 'http://localhost:7700' : `https://api.testpay.dev/ws_${workspace.slug}`;

  async function saveWebhooks() {
    setSaving(true);
    try {
      await api('/api/workspace', {
        method: 'PUT',
        body: JSON.stringify({ webhook_urls: urls }),
      });
      toast.success('Webhook URLs saved');
      mutate('/api/auth/me');
    } catch (e) {
      toast.error(e instanceof ApiError ? e.message : 'save failed');
    } finally {
      setSaving(false);
    }
  }

  return (
    <div className="max-w-2xl space-y-6">
      <h1 className="text-2xl font-semibold">Settings</h1>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Workspace</CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          <div>
            <Label>Slug</Label>
            <code className="block font-mono text-sm bg-muted px-3 py-2 rounded-md mt-1">
              {workspace.slug}
            </code>
          </div>
          <div>
            <Label>API key</Label>
            <div className="mt-1">
              <ApiKeyReveal value={workspace.api_key} />
            </div>
            <p className="text-xs text-muted-foreground mt-1">
              Send as <code className="font-mono">Authorization: Bearer …</code> on your mock
              requests to attribute them to this workspace.
            </p>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Webhook destinations</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <p className="text-xs text-muted-foreground">
            Each gateway gets its own webhook URL. The per-request{' '}
            <code className="font-mono">X-Webhook-URL</code> header overrides this for one call.
          </p>
          {GATEWAYS.map((g) => (
            <div key={g}>
              <Label htmlFor={`webhook-${g}`} className="capitalize">
                {g}
              </Label>
              <Input
                id={`webhook-${g}`}
                type="url"
                placeholder={`https://your-app.example.com/webhooks/${g}`}
                value={urls[g]}
                onChange={(e) => setUrls((prev) => ({ ...prev, [g]: e.target.value }))}
                className="font-mono mt-1"
              />
            </div>
          ))}
          <Button onClick={saveWebhooks} disabled={saving}>
            {saving ? 'Saving…' : 'Save webhook URLs'}
          </Button>
          <div className="border-t pt-3 space-y-1">
            <p className="text-xs font-semibold">Echoed fields in the webhook payload:</p>
            <ul className="text-xs text-muted-foreground list-disc list-inside space-y-0.5">
              <li>
                <b>Stripe:</b> <code className="font-mono">metadata</code> →{' '}
                <code className="font-mono">data.object.metadata</code>
              </li>
              <li>
                <b>Razorpay:</b> <code className="font-mono">notes</code> →{' '}
                <code className="font-mono">payload.payment.entity.notes</code>
              </li>
              <li>
                <b>Agnostic:</b> full request body →{' '}
                <code className="font-mono">request_echo</code>
              </li>
            </ul>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Endpoints</CardTitle>
        </CardHeader>
        <CardContent className="space-y-2">
          {GATEWAYS.map((g) => {
            const url = g === 'agnostic' ? `${baseUrl}/v1` : `${baseUrl}/${g}`;
            return (
              <div key={g} className="flex items-center gap-3">
                <span className="w-20 text-sm uppercase text-muted-foreground">{g}</span>
                <code className="flex-1 font-mono text-sm bg-muted px-3 py-2 rounded-md">
                  {url}
                </code>
                <CopyButton value={url} label="" />
              </div>
            );
          })}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Appearance</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex gap-2">
            {(['light', 'dark', 'system'] as const).map((t) => (
              <Button
                key={t}
                variant={theme === t ? 'default' : 'outline'}
                size="sm"
                onClick={() => setTheme(t)}
              >
                {t}
              </Button>
            ))}
          </div>
        </CardContent>
      </Card>

      {user && (
        <Card>
          <CardHeader>
            <CardTitle className="text-base">Account</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-sm text-muted-foreground">{user.email}</div>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
