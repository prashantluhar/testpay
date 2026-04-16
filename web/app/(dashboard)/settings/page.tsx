'use client';
import { useEffect, useMemo, useState } from 'react';
import { ChevronDown, ChevronRight, AlertCircle, CheckCircle2 } from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Label } from '@/components/ui/label';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { toast } from 'sonner';
import { useMe, useGateways } from '@/lib/hooks';
import { ApiKeyReveal } from '@/components/common/api-key-reveal';
import { useTheme } from '@/components/common/theme-provider';
import { CopyButton } from '@/components/common/copy-button';
import { MODE } from '@/lib/types';
import { api, ApiError } from '@/lib/api';
import { mutate } from 'swr';

// Sentinel key in webhook_urls map that means "use this URL for any gateway
// that has no explicit override". Server-side fallback logic reads this too.
const DEFAULT_KEY = '_default';

export default function SettingsPage() {
  const { data: me } = useMe();
  const { data: gateways = [] } = useGateways();
  const { theme, setTheme } = useTheme();

  const [defaultUrl, setDefaultUrl] = useState('');
  const [overrides, setOverrides] = useState<Record<string, string>>({});
  const [showOverrides, setShowOverrides] = useState(false);
  const [saving, setSaving] = useState(false);

  // Seed form from persisted workspace webhook_urls.
  useEffect(() => {
    const urls = me?.workspace?.webhook_urls ?? {};
    setDefaultUrl(urls[DEFAULT_KEY] ?? '');
    const ov: Record<string, string> = {};
    for (const g of gateways) {
      if (urls[g]) ov[g] = urls[g];
    }
    setOverrides(ov);
    // Auto-expand the overrides section if any are non-empty.
    if (Object.keys(ov).length > 0) setShowOverrides(true);
  }, [me?.workspace?.webhook_urls, gateways]);

  const configuredCount = useMemo(() => {
    let n = defaultUrl ? 1 : 0;
    n += Object.values(overrides).filter((v) => v && v !== defaultUrl).length;
    return n;
  }, [defaultUrl, overrides]);

  if (!me) return null;
  const { workspace, user } = me;

  const baseUrl =
    MODE === 'local' ? 'http://localhost:7700' : `https://api.testpay.dev/ws_${workspace.slug}`;

  async function saveWebhooks() {
    setSaving(true);
    try {
      const payload: Record<string, string> = {};
      if (defaultUrl) payload[DEFAULT_KEY] = defaultUrl;
      for (const [g, v] of Object.entries(overrides)) {
        if (v && v !== defaultUrl) payload[g] = v;
      }
      await api('/api/workspace', {
        method: 'PUT',
        body: JSON.stringify({ webhook_urls: payload }),
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
    <div className="max-w-3xl space-y-6">
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
              Send as <code className="font-mono">Authorization: Bearer …</code> on mock requests
              to attribute them to this workspace.
            </p>
          </div>
        </CardContent>
      </Card>

      {/* Webhook destinations — redesigned */}
      <Card>
        <CardHeader>
          <CardTitle className="text-base flex items-center justify-between">
            <span>Webhook destinations</span>
            <span className="text-xs font-normal text-muted-foreground">
              {configuredCount === 0
                ? 'none configured'
                : `${configuredCount} configured`}
            </span>
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-5">
          {/* Default URL — applies to all gateways that don't have an override */}
          <div>
            <Label htmlFor="default-webhook" className="flex items-center gap-2">
              Default URL
              {defaultUrl && <CheckCircle2 className="h-3.5 w-3.5 text-emerald-500" />}
            </Label>
            <Input
              id="default-webhook"
              type="url"
              placeholder="https://your-app.example.com/webhook"
              value={defaultUrl}
              onChange={(e) => setDefaultUrl(e.target.value)}
              className="font-mono mt-1"
            />
            <p className="text-xs text-muted-foreground mt-1.5">
              Used for every gateway unless you set a specific override below. Per-request
              <code className="font-mono mx-1">X-Webhook-URL</code>header still overrides both.
            </p>
          </div>

          {/* Per-gateway overrides — collapsible */}
          <div className="border-t pt-4">
            <button
              type="button"
              onClick={() => setShowOverrides((v) => !v)}
              className="flex items-center gap-2 text-sm text-muted-foreground hover:text-foreground transition-colors"
            >
              {showOverrides ? (
                <ChevronDown className="h-4 w-4" />
              ) : (
                <ChevronRight className="h-4 w-4" />
              )}
              Per-gateway overrides
              <span className="text-xs text-muted-foreground">
                ({Object.values(overrides).filter((v) => v && v !== defaultUrl).length} set
                {gateways.length > 0 ? ` of ${gateways.length}` : ''})
              </span>
            </button>

            {showOverrides && (
              <div className="mt-3 space-y-2">
                {gateways.length === 0 ? (
                  <div className="text-xs text-muted-foreground flex items-center gap-2">
                    <AlertCircle className="h-3.5 w-3.5" />
                    Loading gateway list…
                  </div>
                ) : (
                  gateways.map((g) => (
                    <GatewayOverrideRow
                      key={g}
                      gateway={g}
                      value={overrides[g] ?? ''}
                      placeholder="Uses default URL"
                      onChange={(v) =>
                        setOverrides((prev) => ({ ...prev, [g]: v }))
                      }
                    />
                  ))
                )}
              </div>
            )}
          </div>

          <div className="flex gap-2 pt-2 border-t">
            <Button onClick={saveWebhooks} disabled={saving}>
              {saving ? 'Saving…' : 'Save'}
            </Button>
            <Button
              variant="ghost"
              onClick={() => {
                const urls = me.workspace.webhook_urls ?? {};
                setDefaultUrl(urls[DEFAULT_KEY] ?? '');
                const ov: Record<string, string> = {};
                for (const g of gateways) if (urls[g]) ov[g] = urls[g];
                setOverrides(ov);
              }}
            >
              Reset
            </Button>
          </div>
        </CardContent>
      </Card>

      <EndpointsCard gateways={gateways} baseUrl={baseUrl} />

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

function EndpointsCard({ gateways, baseUrl }: { gateways: string[]; baseUrl: string }) {
  const [selected, setSelected] = useState<string | null>(null);
  const selectedUrl =
    selected === 'agnostic' ? `${baseUrl}/v1` : selected ? `${baseUrl}/${selected}` : '';

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base flex items-center justify-between">
          <span>Endpoints</span>
          <span className="text-xs font-normal text-muted-foreground">
            {gateways.length} gateways
          </span>
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <p className="text-xs text-muted-foreground">
          Point your app at <code className="font-mono">{baseUrl}/{'{gateway}'}</code>. Click a
          chip below to copy its full URL.
        </p>
        <div className="flex flex-wrap gap-2">
          {gateways.map((g) => {
            const url = g === 'agnostic' ? `${baseUrl}/v1` : `${baseUrl}/${g}`;
            const active = selected === g;
            return (
              <button
                key={g}
                type="button"
                onClick={() => {
                  navigator.clipboard.writeText(url);
                  setSelected(g);
                  setTimeout(() => setSelected((cur) => (cur === g ? null : cur)), 1500);
                }}
                className={`px-3 py-1 rounded-md border text-xs font-mono transition-colors ${
                  active
                    ? 'bg-emerald-500/10 border-emerald-500/40 text-emerald-500'
                    : 'bg-muted hover:bg-accent hover:border-border'
                }`}
                title={url}
              >
                {active ? 'copied' : g}
              </button>
            );
          })}
        </div>
        {selected && (
          <div className="flex items-center gap-2 pt-2 border-t">
            <span className="w-20 text-xs uppercase tracking-wider text-muted-foreground">
              {selected}
            </span>
            <code className="flex-1 font-mono text-xs bg-muted px-3 py-2 rounded-md truncate">
              {selectedUrl}
            </code>
            <CopyButton value={selectedUrl} label="" />
          </div>
        )}
      </CardContent>
    </Card>
  );
}

function GatewayOverrideRow({
  gateway,
  value,
  placeholder,
  onChange,
}: {
  gateway: string;
  value: string;
  placeholder: string;
  onChange: (v: string) => void;
}) {
  return (
    <div className="flex items-center gap-2">
      <span className="w-24 text-xs uppercase tracking-wider text-muted-foreground shrink-0">
        {gateway}
      </span>
      <Input
        type="url"
        value={value}
        placeholder={placeholder}
        onChange={(e) => onChange(e.target.value)}
        className="font-mono text-xs flex-1"
      />
    </div>
  );
}
