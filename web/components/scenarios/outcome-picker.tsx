'use client';
import { FAILURE_MODES, type FailureMode } from '@/lib/failure-modes';
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectLabel,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';

export function OutcomePicker({
  value,
  onChange,
}: {
  value: FailureMode;
  onChange: (v: FailureMode) => void;
}) {
  const groups = FAILURE_MODES.reduce<Record<string, (typeof FAILURE_MODES)[number][]>>(
    (acc, m) => {
      (acc[m.group] ||= []).push(m);
      return acc;
    },
    {},
  );

  return (
    <Select value={value} onValueChange={(v) => onChange(v as FailureMode)}>
      <SelectTrigger className="font-mono text-xs">
        <SelectValue />
      </SelectTrigger>
      <SelectContent>
        {Object.entries(groups).map(([group, items]) => (
          <SelectGroup key={group}>
            <SelectLabel>{group}</SelectLabel>
            {items.map((m) => (
              <SelectItem key={m.value} value={m.value} className="font-mono text-xs">
                {m.value}
              </SelectItem>
            ))}
          </SelectGroup>
        ))}
      </SelectContent>
    </Select>
  );
}
