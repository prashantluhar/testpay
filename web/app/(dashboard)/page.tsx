'use client';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { StatCard } from '@/components/overview/stat-card';
import { LiveFeed } from '@/components/overview/live-feed';
import { useLogs, useScenarios } from '@/lib/hooks';

export default function OverviewPage() {
  const { data: logs } = useLogs({ limit: 500 });
  const { data: scenarios } = useScenarios();

  const total = logs?.length ?? 0;
  const errors = logs?.filter((l) => l.response_status >= 400).length ?? 0;
  const success = total > 0 ? (((total - errors) / total) * 100).toFixed(1) + '%' : '—';

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold">Overview</h1>
        <p className="text-sm text-muted-foreground">Your mock gateway, live.</p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <StatCard label="Requests today" value={total} />
        <StatCard label="Active scenarios" value={scenarios?.length ?? '—'} />
        <StatCard label="Success rate" value={success} accent="good" />
      </div>

      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <CardTitle className="text-base">Live feed</CardTitle>
          <span className="text-xs text-emerald-500">● live</span>
        </CardHeader>
        <CardContent className="p-0">
          <LiveFeed />
        </CardContent>
      </Card>
    </div>
  );
}
