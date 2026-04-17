'use client';
import { useState } from 'react';
import { Box, Button, Card, Flex, Heading, Text } from '@radix-ui/themes';
import { StatCard } from '@/components/overview/stat-card';
import { LiveFeed } from '@/components/overview/live-feed';
import { useLogs, useScenarios, useWebhooks, useGateways } from '@/lib/hooks';
import { MODE } from '@/lib/types';
import { CopyButton } from '@/components/common/copy-button';
import { Spinner } from '@/components/common/spinner';
import Link from 'next/link';
import { ListBulletIcon, PaperPlaneIcon, FileTextIcon } from '@radix-ui/react-icons';

export default function OverviewPage() {
  const { data: logs } = useLogs({ limit: 500, pollInterval: 3000 });
  const { data: scenarios } = useScenarios();
  const { data: webhooks } = useWebhooks({ limit: 500 });
  const { data: gateways = [] } = useGateways();
  const [copiedGateway, setCopiedGateway] = useState<string | null>(null);

  const total = logs?.length ?? 0;
  const errors = logs?.filter((l) => l.response_status >= 400).length ?? 0;
  const success = total > 0 ? (((total - errors) / total) * 100).toFixed(1) + '%' : '—';
  const delivered = webhooks?.filter((w) => w.delivery_status === 'delivered').length ?? 0;
  const failed = webhooks?.filter((w) => w.delivery_status === 'failed').length ?? 0;

  const baseUrl =
    MODE === 'local' ? 'http://localhost:7700' : (process.env.NEXT_PUBLIC_API_BASE || '');

  function copyGatewayUrl(g: string) {
    const url = g === 'agnostic' ? `${baseUrl}/v1` : `${baseUrl}/${g}`;
    navigator.clipboard.writeText(url);
    setCopiedGateway(g);
    setTimeout(() => setCopiedGateway((c) => (c === g ? null : c)), 1500);
  }

  return (
    <div className="space-y-6 animate-in fade-in duration-300">
      {/* Hero: workspace endpoints */}
      <div className="rounded-xl border bg-gradient-to-br from-card to-accent/30 p-6">
        <div className="flex items-start justify-between gap-4 flex-wrap">
          <div>
            <Heading size="6">Welcome back</Heading>
            <Text size="2" color="gray" className="mt-1 block">
              Point your app at the base URL, appended with any gateway below. Every request is
              logged; scenarios shape the responses.
            </Text>
          </div>
          <Flex gap="2">
            <Button asChild variant="outline" size="2">
              <Link href="/scenarios/new">
                <ListBulletIcon />
                New scenario
              </Link>
            </Button>
            <Button asChild size="2">
              <Link href="/settings">Settings</Link>
            </Button>
          </Flex>
        </div>

        {/* Base URL */}
        <div className="mt-5 flex items-center gap-2 bg-background/80 border rounded-md px-3 py-2">
          <span className="text-xs uppercase tracking-wider text-muted-foreground shrink-0">
            base
          </span>
          <code className="flex-1 truncate font-mono text-sm">{baseUrl}</code>
          <CopyButton value={baseUrl} label="" />
          <Link
            href="/docs"
            className="text-xs text-muted-foreground hover:text-foreground underline shrink-0 ml-1"
          >
            Read the docs →
          </Link>
        </div>

        {/* Gateway chips — click to copy full URL */}
        <div className="mt-3 flex flex-wrap gap-1.5">
          {gateways.length === 0 ? (
            <span className="text-xs text-muted-foreground inline-flex items-center gap-2">
              <Spinner size="small" />
              Loading gateways…
            </span>
          ) : (
            gateways.map((g) => {
              const active = copiedGateway === g;
              return (
                <button
                  key={g}
                  type="button"
                  onClick={() => copyGatewayUrl(g)}
                  className={`px-2.5 py-1 rounded border text-xs font-mono transition-colors ${
                    active
                      ? 'bg-emerald-500/10 border-emerald-500/40 text-emerald-500'
                      : 'bg-background/70 hover:bg-accent hover:border-border'
                  }`}
                  title={g === 'agnostic' ? `${baseUrl}/v1` : `${baseUrl}/${g}`}
                >
                  {active ? 'copied!' : g}
                </button>
              );
            })
          )}
        </div>
      </div>

      {/* Stat cards */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <StatCard label="Requests" value={total} />
        <StatCard label="Success rate" value={success} accent="good" />
        <StatCard label="Scenarios" value={scenarios?.length ?? 0} />
        <StatCard
          label="Webhooks"
          value={`${delivered}/${delivered + failed}`}
          accent={failed > 0 ? 'bad' : 'good'}
        />
      </div>

      {/* Live feed + quick links */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <Card className="lg:col-span-2">
          <Box p="2">
            <Flex align="center" justify="between" mb="2">
              <Heading size="3" as="h3">
                <Flex align="center" gap="2">
                  <FileTextIcon className="text-muted-foreground" />
                  Live feed
                </Flex>
              </Heading>
              <span className="text-xs text-emerald-500 flex items-center gap-1">
                <span className="h-2 w-2 rounded-full bg-emerald-500 animate-pulse" />
                live
              </span>
            </Flex>
          </Box>
          <LiveFeed />
        </Card>

        <div className="space-y-4">
          <Card>
            <Box p="2">
              <Heading size="3" mb="2">
                <Flex align="center" gap="2">
                  <ListBulletIcon className="text-muted-foreground" />
                  Scenarios
                </Flex>
              </Heading>
              <div className="text-sm space-y-2">
                {scenarios && scenarios.length > 0 ? (
                  <>
                    <div className="text-muted-foreground">
                      {scenarios.length} configured,{' '}
                      {scenarios.filter((s) => s.is_default).length} default
                    </div>
                    <ul className="space-y-1 text-xs">
                      {scenarios.slice(0, 5).map((s) => (
                        <li key={s.id} className="flex items-center justify-between gap-2">
                          <Link
                            href={`/scenarios/edit?id=${s.id}`}
                            className="truncate hover:underline"
                          >
                            {s.name}
                          </Link>
                          {s.is_default && (
                            <span className="text-emerald-500 text-[10px]">DEFAULT</span>
                          )}
                        </li>
                      ))}
                    </ul>
                  </>
                ) : (
                  <div className="text-muted-foreground">
                    None yet.{' '}
                    <Link href="/scenarios/new" className="underline text-foreground">
                      Create one
                    </Link>
                    .
                  </div>
                )}
              </div>
            </Box>
          </Card>

          <Card>
            <Box p="2">
              <Heading size="3" mb="2">
                <Flex align="center" gap="2">
                  <PaperPlaneIcon className="text-muted-foreground" />
                  Webhooks
                </Flex>
              </Heading>
              <div className="text-sm space-y-2">
                {delivered + failed > 0 ? (
                  <>
                    <div className="flex justify-between">
                      <span className="text-muted-foreground">Delivered</span>
                      <span className="text-emerald-500">{delivered}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-muted-foreground">Failed</span>
                      <span className={failed > 0 ? 'text-red-500' : 'text-muted-foreground'}>
                        {failed}
                      </span>
                    </div>
                    <Link href="/webhooks" className="block text-xs underline mt-2">
                      View all
                    </Link>
                  </>
                ) : (
                  <div className="text-muted-foreground">
                    No deliveries yet. Set webhook URLs in{' '}
                    <Link href="/settings" className="underline text-foreground">
                      Settings
                    </Link>
                    .
                  </div>
                )}
              </div>
            </Box>
          </Card>
        </div>
      </div>
    </div>
  );
}
