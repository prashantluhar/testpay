'use client';
import { useEffect, useMemo, useState } from 'react';
import { useSearchParams } from 'next/navigation';
import { useLogs } from '@/lib/hooks';
import { LogFilters, type LogFiltersValue } from '@/components/logs/log-filters';
import { LogsTable } from '@/components/logs/logs-table';
import { LogDetailDrawer } from '@/components/logs/log-detail-drawer';

export default function LogsPage() {
  const search = useSearchParams();
  const deeplinkId = search.get('id');

  const [filters, setFilters] = useState<LogFiltersValue>({
    gateway: 'all',
    statusClass: 'all',
    scenarioId: 'all',
    search: '',
  });
  const [selected, setSelected] = useState<string | null>(null);
  const { data } = useLogs({ limit: 200 });
  const loading = data === undefined;

  // Deep-link: if the URL carries ?id=..., open the drawer for that log
  // immediately on mount (and again if the param changes — e.g. user clicks
  // another live-feed row while on this page).
  useEffect(() => {
    if (deeplinkId) setSelected(deeplinkId);
  }, [deeplinkId]);

  const rows = useMemo(() => {
    if (!data) return [];
    const q = filters.search.trim().toLowerCase();
    return data.filter((l) => {
      if (filters.gateway !== 'all' && l.gateway !== filters.gateway) return false;
      if (filters.statusClass !== 'all') {
        const s = l.response_status;
        if (filters.statusClass === '2xx' && !(s >= 200 && s < 300)) return false;
        if (filters.statusClass === '4xx' && !(s >= 400 && s < 500)) return false;
        if (filters.statusClass === '5xx' && !(s >= 500 && s < 600)) return false;
      }
      if (filters.scenarioId === 'none') {
        if (l.scenario_id) return false;
      } else if (filters.scenarioId !== 'all') {
        if (l.scenario_id !== filters.scenarioId) return false;
      }
      if (q) {
        // Broadened search — match path, log ID prefix, merchant order ref,
        // and the HTTP status as a string ("429", "500", …).
        const hay = [
          l.path,
          l.id,
          l.merchant_order_id || '',
          String(l.response_status),
          l.method,
        ]
          .join(' ')
          .toLowerCase();
        if (!hay.includes(q)) return false;
      }
      return true;
    });
  }, [data, filters]);

  const filterKey = `${filters.gateway}|${filters.statusClass}|${filters.scenarioId}|${filters.search}`;

  return (
    <div className="flex flex-col h-[calc(100vh-8rem)] animate-in fade-in duration-300">
      <div className="flex items-end justify-between mb-3 shrink-0">
        <h1 className="text-2xl font-semibold">Logs</h1>
        <span className="text-xs text-muted-foreground">{rows.length} shown</span>
      </div>
      <div className="shrink-0">
        <LogFilters value={filters} onChange={setFilters} />
      </div>
      <div className="flex-1 min-h-0 overflow-y-auto border rounded-md">
        <LogsTable
          rows={rows}
          selectedId={selected}
          onSelect={setSelected}
          loading={loading}
          filterKey={filterKey}
        />
      </div>
      <LogDetailDrawer id={selected} onClose={() => setSelected(null)} />
    </div>
  );
}
