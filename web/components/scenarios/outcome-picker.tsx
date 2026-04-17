'use client';
import { FAILURE_MODES, type FailureMode } from '@/lib/failure-modes';
import { Select } from '@radix-ui/themes';

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
    <Select.Root value={value} onValueChange={(v) => onChange(v as FailureMode)}>
      <Select.Trigger className="font-mono text-xs" />
      <Select.Content>
        {Object.entries(groups).map(([group, items]) => (
          <Select.Group key={group}>
            <Select.Label>{group}</Select.Label>
            {items.map((m) => (
              <Select.Item key={m.value} value={m.value} className="font-mono text-xs">
                {m.value}
              </Select.Item>
            ))}
          </Select.Group>
        ))}
      </Select.Content>
    </Select.Root>
  );
}
