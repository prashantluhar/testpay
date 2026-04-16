'use client';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Input } from '@/components/ui/input';

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
      <Select
        value={value.gateway}
        onValueChange={(v) => onChange({ ...value, gateway: v as LogFiltersValue['gateway'] })}
      >
        <SelectTrigger className="w-32">
          <SelectValue />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="all">All gateways</SelectItem>
          <SelectItem value="stripe">stripe</SelectItem>
          <SelectItem value="razorpay">razorpay</SelectItem>
          <SelectItem value="agnostic">agnostic</SelectItem>
        </SelectContent>
      </Select>
      <Select
        value={value.statusClass}
        onValueChange={(v) =>
          onChange({ ...value, statusClass: v as LogFiltersValue['statusClass'] })
        }
      >
        <SelectTrigger className="w-32">
          <SelectValue />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="all">All statuses</SelectItem>
          <SelectItem value="2xx">2xx</SelectItem>
          <SelectItem value="4xx">4xx</SelectItem>
          <SelectItem value="5xx">5xx</SelectItem>
        </SelectContent>
      </Select>
      <Input
        placeholder="Search path…"
        className="w-64"
        value={value.search}
        onChange={(e) => onChange({ ...value, search: e.target.value })}
      />
    </div>
  );
}
