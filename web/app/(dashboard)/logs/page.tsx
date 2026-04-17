'use client';
import { useMemo, useState } from 'react';
import { useLogs } from '@/lib/hooks';
import { LogFilters, type LogFiltersValue } from '@/components/logs/log-filters';
import { LogsTable } from '@/components/logs/logs-table';
import { LogDetailDrawer } from '@/components/logs/log-detail-drawer';

export default function LogsPage() {
  const [filters, setFilters] = useState<LogFiltersValue>({
    gateway: 'all',
    statusClass: 'all',
    search: '',
  });
  const [selected, setSelected] = useState<string | null>(null);
  const { data } = useLogs({ limit: 200 });
  const loading = data === undefined;

  const rows = useMemo(() => {
    if (!data) return [];
    return data.filter((l) => {
      if (filters.gateway !== 'all' && l.gateway !== filters.gateway) return false;
      if (filters.statusClass !== 'all') {
        const s = l.response_status;
        if (filters.statusClass === '2xx' && !(s >= 200 && s < 300)) return false;
        if (filters.statusClass === '4xx' && !(s >= 400 && s < 500)) return false;
        if (filters.statusClass === '5xx' && !(s >= 500 && s < 600)) return false;
      }
      if (filters.search && !l.path.toLowerCase().includes(filters.search.toLowerCase()))
        return false;
      return true;
    });
  }, [data, filters]);

  const filterKey = `${filters.gateway}|${filters.statusClass}|${filters.search}`;

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
          onSelect={setSelected}
          loading={loading}
          filterKey={filterKey}
        />
      </div>
      <LogDetailDrawer id={selected} onClose={() => setSelected(null)} />
    </div>
  );
}
