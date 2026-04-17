'use client';
import { useEffect, useMemo, useState } from 'react';
import {
  ChevronDownIcon,
  ChevronRightIcon,
  CheckCircledIcon,
} from '@radix-ui/react-icons';
import { Box, Button, Card, Flex, Heading, Text, TextField } from '@radix-ui/themes';
import { toast } from 'sonner';
import { useMe, useGateways } from '@/lib/hooks';
import { ApiKeyReveal } from '@/components/common/api-key-reveal';
import { useTheme } from '@/components/common/theme-provider';
import { useThemePreset } from '@/components/common/theme-preset-provider';
import { CopyButton } from '@/components/common/copy-button';
import { Spinner } from '@/components/common/spinner';
import { THEME_PRESETS } from '@/lib/themes';
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
  const { presetId, setPresetId } = useThemePreset();

  const [defaultUrl, setDefaultUrl] = useState('');
  const [overrides, setOverrides] = useState<Record<string, string>>({});
  const [showOverrides, setShowOverrides] = useState(false);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    const urls = me?.workspace?.webhook_urls ?? {};
    setDefaultUrl(urls[DEFAULT_KEY] ?? '');
    const ov: Record<string, string> = {};
    for (const g of gateways) {
      if (urls[g]) ov[g] = urls[g];
    }
    setOverrides(ov);
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
    MODE === 'local' ? 'http://localhost:7700' : (process.env.NEXT_PUBLIC_API_BASE || '');

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
    <div className="max-w-3xl space-y-6 animate-in fade-in duration-300">
      <Heading size="6">Settings</Heading>

      <Card>
        <Box p="2">
          <Heading size="3" mb="3">
            Workspace
          </Heading>
          <Flex direction="column" gap="3">
            <div>
              <Text as="label" size="2" weight="medium">
                Slug
              </Text>
              <code className="block font-mono text-sm bg-muted px-3 py-2 rounded-md mt-1">
                {workspace.slug}
              </code>
            </div>
            <div>
              <Text as="label" size="2" weight="medium">
                API key
              </Text>
              <div className="mt-1">
                <ApiKeyReveal value={workspace.api_key} />
              </div>
              <Text size="1" color="gray" mt="1" as="p">
                Send as <code className="font-mono">Authorization: Bearer …</code> on mock requests
                to attribute them to this workspace.
              </Text>
            </div>
          </Flex>
        </Box>
      </Card>

      <Card>
        <Box p="2">
          <Flex align="center" justify="between" mb="4">
            <Heading size="3">Webhook destinations</Heading>
            <Text size="1" color="gray">
              {configuredCount === 0 ? 'none configured' : `${configuredCount} configured`}
            </Text>
          </Flex>
          <Flex direction="column" gap="5">
            <div>
              <Text
                as="label"
                size="2"
                weight="medium"
                htmlFor="default-webhook"
                className="flex items-center gap-2"
              >
                Default URL
                {defaultUrl && (
                  <CheckCircledIcon className="h-3.5 w-3.5 text-emerald-500" />
                )}
              </Text>
              <TextField.Root
                id="default-webhook"
                type="url"
                placeholder="https://your-app.example.com/webhook"
                value={defaultUrl}
                onChange={(e) => setDefaultUrl(e.target.value)}
                mt="1"
                className="font-mono"
              />
              <Text size="1" color="gray" mt="1" as="p">
                Used for every gateway unless you set a specific override below. Per-request
                <code className="font-mono mx-1">X-Webhook-URL</code>header still overrides both.
              </Text>
            </div>

            <div className="border-t pt-4">
              <button
                type="button"
                onClick={() => setShowOverrides((v) => !v)}
                className="flex items-center gap-2 text-sm text-muted-foreground hover:text-foreground transition-colors"
              >
                {showOverrides ? <ChevronDownIcon /> : <ChevronRightIcon />}
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
                      <Spinner size="small" />
                      Loading gateway list…
                    </div>
                  ) : (
                    gateways.map((g) => (
                      <GatewayOverrideRow
                        key={g}
                        gateway={g}
                        value={overrides[g] ?? ''}
                        placeholder="Uses default URL"
                        onChange={(v) => setOverrides((prev) => ({ ...prev, [g]: v }))}
                      />
                    ))
                  )}
                </div>
              )}
            </div>

            <Flex gap="2" pt="2" className="border-t">
              <Button
                onClick={saveWebhooks}
                disabled={saving}
                loading={saving}
                className="transition-transform hover:-translate-y-px"
              >
                {saving ? 'Saving…' : 'Save'}
              </Button>
              <Button
                variant="soft"
                color="gray"
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
            </Flex>
          </Flex>
        </Box>
      </Card>

      <EndpointsCard gateways={gateways} baseUrl={baseUrl} />

      <Card>
        <Box p="2">
          <Heading size="3" mb="3">
            Appearance
          </Heading>
          <Flex direction="column" gap="5">
            <div>
              <Text size="2" weight="medium" as="div" mb="2">
                Mode
              </Text>
              <Flex gap="2">
                {(['light', 'dark', 'system'] as const).map((t) => (
                  <Button
                    key={t}
                    variant={theme === t ? 'solid' : 'outline'}
                    size="2"
                    onClick={() => setTheme(t)}
                  >
                    {t}
                  </Button>
                ))}
              </Flex>
            </div>

            <div>
              <Text size="2" weight="medium" as="div" mb="2">
                Color theme
              </Text>
              <Text size="1" color="gray" as="p" mb="3">
                Pick an accent + gray palette. Applies instantly and is saved to
                this browser.
              </Text>
              <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-5 gap-2">
                {THEME_PRESETS.map((p) => {
                  const active = presetId === p.id;
                  return (
                    <button
                      key={p.id}
                      type="button"
                      onClick={() => setPresetId(p.id)}
                      className="text-left rounded-md border bg-card p-3 transition-all hover:-translate-y-px hover:border-[var(--accent-8)]"
                      style={
                        active
                          ? { boxShadow: 'inset 0 0 0 2px var(--accent-9)' }
                          : undefined
                      }
                      aria-pressed={active}
                    >
                      <div className="flex items-center gap-2 mb-1">
                        <span
                          className="h-4 w-4 rounded-full border border-black/10 dark:border-white/10"
                          style={{ background: `var(--${p.accentColor}-9)` }}
                        />
                        <span
                          className="h-4 w-4 rounded-full border border-black/10 dark:border-white/10"
                          style={{ background: `var(--${p.grayColor}-9)` }}
                        />
                        <Text size="2" weight="medium">
                          {p.name}
                        </Text>
                      </div>
                      <Text size="1" color="gray" as="div">
                        {p.description}
                      </Text>
                    </button>
                  );
                })}
              </div>
            </div>
          </Flex>
        </Box>
      </Card>

      {user && (
        <Card>
          <Box p="2">
            <Heading size="3" mb="2">
              Account
            </Heading>
            <Text size="2" color="gray">
              {user.email}
            </Text>
          </Box>
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
      <Box p="2">
        <Flex align="center" justify="between" mb="3">
          <Heading size="3">Endpoints</Heading>
          <Text size="1" color="gray">
            {gateways.length} gateways
          </Text>
        </Flex>
        <Flex direction="column" gap="4">
          <Text size="1" color="gray" as="p">
            Point your app at <code className="font-mono">{baseUrl}/{'{gateway}'}</code>. Click a
            chip below to copy its full URL.
          </Text>
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
        </Flex>
      </Box>
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
      <TextField.Root
        type="url"
        value={value}
        placeholder={placeholder}
        onChange={(e) => onChange(e.target.value)}
        className="font-mono text-xs flex-1"
      />
    </div>
  );
}
