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

  return (
    <div className="space-y-3">
      <h1 className="text-2xl font-semibold">Logs</h1>
      <LogFilters value={filters} onChange={setFilters} />
      <LogsTable rows={rows} onSelect={setSelected} />
      <LogDetailDrawer id={selected} onClose={() => setSelected(null)} />
    </div>
  );
}
