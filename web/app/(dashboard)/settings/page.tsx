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

export default function SettingsPage() {
  const { data } = useMe();
  const { theme, setTheme } = useTheme();
  const [webhookUrl, setWebhookUrl] = useState('');
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    if (data?.workspace?.webhook_url !== undefined) {
      setWebhookUrl(data.workspace.webhook_url);
    }
  }, [data?.workspace?.webhook_url]);

  if (!data) return null;
  const { workspace, user } = data;

  const baseUrl =
    MODE === 'local' ? 'http://localhost:7700' : `https://api.testpay.dev/ws_${workspace.slug}`;

  async function saveWebhook() {
    setSaving(true);
    try {
      await api('/api/workspace', {
        method: 'PUT',
        body: JSON.stringify({ webhook_url: webhookUrl }),
      });
      toast.success('Webhook URL saved');
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
              Send as <code className="font-mono">Authorization: Bearer …</code> on your mock requests
              to attribute them to this workspace.
            </p>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Webhook destination</CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          <div>
            <Label htmlFor="webhook-url">Default webhook URL</Label>
            <div className="flex gap-2 mt-1">
              <Input
                id="webhook-url"
                type="url"
                placeholder="https://your-app.example.com/webhook"
                value={webhookUrl}
                onChange={(e) => setWebhookUrl(e.target.value)}
                className="font-mono"
              />
              <Button onClick={saveWebhook} disabled={saving}>
                {saving ? 'Saving…' : 'Save'}
              </Button>
            </div>
            <p className="text-xs text-muted-foreground mt-2">
              Webhooks for this workspace are POSTed here. Per-request{' '}
              <code className="font-mono">X-Webhook-URL</code> header overrides this for one call.
            </p>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Endpoints</CardTitle>
        </CardHeader>
        <CardContent className="space-y-2">
          {['stripe', 'razorpay', 'agnostic'].map((g) => {
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
