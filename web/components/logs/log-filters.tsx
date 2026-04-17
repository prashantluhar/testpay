'use client';
import { Select, TextField } from '@radix-ui/themes';

export interface LogFiltersValue {
  gateway: 'all' | 'stripe' | 'razorpay' | 'agnostic';
  statusClass: 'all' | '2xx' | '4xx' | '5xx';
  search: string;
}

export function LogFilters({
  value,
  onChange,
}: {
  value: LogFiltersValue;
  onChange: (v: LogFiltersValue) => void;
}) {
  return (
    <div className="flex items-center gap-2 mb-3">
      <Select.Root
        value={value.gateway}
        onValueChange={(v) => onChange({ ...value, gateway: v as LogFiltersValue['gateway'] })}
      >
        <Select.Trigger className="w-32" />
        <Select.Content>
          <Select.Item value="all">All gateways</Select.Item>
          <Select.Item value="stripe">stripe</Select.Item>
          <Select.Item value="razorpay">razorpay</Select.Item>
          <Select.Item value="agnostic">agnostic</Select.Item>
        </Select.Content>
      </Select.Root>
      <Select.Root
        value={value.statusClass}
        onValueChange={(v) =>
          onChange({ ...value, statusClass: v as LogFiltersValue['statusClass'] })
        }
      >
        <Select.Trigger className="w-32" />
        <Select.Content>
          <Select.Item value="all">All statuses</Select.Item>
          <Select.Item value="2xx">2xx</Select.Item>
          <Select.Item value="4xx">4xx</Select.Item>
          <Select.Item value="5xx">5xx</Select.Item>
        </Select.Content>
      </Select.Root>
      <TextField.Root
        placeholder="Search path…"
        className="w-64"
        value={value.search}
        onChange={(e) => onChange({ ...value, search: e.target.value })}
      />
    </div>
  );
}
