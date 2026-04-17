'use client';
import { Select, TextField } from '@radix-ui/themes';
import { MagnifyingGlassIcon, Cross2Icon } from '@radix-ui/react-icons';
import { useGateways, useScenarios } from '@/lib/hooks';

export interface LogFiltersValue {
  gateway: string;
  statusClass: 'all' | '2xx' | '4xx' | '5xx';
  scenarioId: string; // 'all' | 'none' | <scenario uuid>
  search: string;
}

export function LogFilters({
  value,
  onChange,
}: {
  value: LogFiltersValue;
  onChange: (v: LogFiltersValue) => void;
}) {
  const { data: gateways = [] } = useGateways();
  const { data: scenarios = [] } = useScenarios();

  return (
    <div className="flex items-center gap-2 mb-3 flex-wrap">
      <Select.Root
        value={value.gateway}
        onValueChange={(v) => onChange({ ...value, gateway: v })}
      >
        <Select.Trigger placeholder="Gateway" className="min-w-[140px]" />
        <Select.Content>
          <Select.Item value="all">All gateways</Select.Item>
          <Select.Separator />
          {gateways.map((g) => (
            <Select.Item key={g} value={g}>
              {g}
            </Select.Item>
          ))}
        </Select.Content>
      </Select.Root>

      <Select.Root
        value={value.statusClass}
        onValueChange={(v) =>
          onChange({ ...value, statusClass: v as LogFiltersValue['statusClass'] })
        }
      >
        <Select.Trigger placeholder="Status" className="min-w-[128px]" />
        <Select.Content>
          <Select.Item value="all">All statuses</Select.Item>
          <Select.Separator />
          <Select.Item value="2xx">2xx success</Select.Item>
          <Select.Item value="4xx">4xx client</Select.Item>
          <Select.Item value="5xx">5xx server</Select.Item>
        </Select.Content>
      </Select.Root>

      <Select.Root
        value={value.scenarioId}
        onValueChange={(v) => onChange({ ...value, scenarioId: v })}
      >
        <Select.Trigger placeholder="Scenario" className="min-w-[180px]" />
        <Select.Content>
          <Select.Item value="all">All scenarios</Select.Item>
          <Select.Item value="none">— No scenario —</Select.Item>
          {scenarios.length > 0 && <Select.Separator />}
          {scenarios.map((s) => (
            <Select.Item key={s.id} value={s.id}>
              {s.name}
            </Select.Item>
          ))}
        </Select.Content>
      </Select.Root>

      <TextField.Root
        placeholder="Search id, path, order ref…"
        className="flex-1 min-w-[240px] max-w-[520px]"
        value={value.search}
        onChange={(e) => onChange({ ...value, search: e.target.value })}
      >
        <TextField.Slot>
          <MagnifyingGlassIcon height="14" width="14" />
        </TextField.Slot>
        {value.search && (
          <TextField.Slot pr="1">
            <button
              type="button"
              onClick={() => onChange({ ...value, search: '' })}
              className="text-[var(--gray-10)] hover:text-[var(--gray-12)] transition-colors"
              aria-label="Clear search"
            >
              <Cross2Icon height="14" width="14" />
            </button>
          </TextField.Slot>
        )}
      </TextField.Root>
    </div>
  );
}
