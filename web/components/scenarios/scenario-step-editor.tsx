'use client';
import { Button, Select, TextField } from '@radix-ui/themes';
import { ArrowUpIcon, ArrowDownIcon, TrashIcon, PlusIcon } from '@radix-ui/react-icons';
import { OutcomePicker } from './outcome-picker';
import { EVENT_TYPES, type EventType, type FailureMode } from '@/lib/failure-modes';

export interface Step {
  event: EventType;
  outcome: FailureMode;
  code?: string;
}

export function ScenarioStepEditor({
  steps,
  onChange,
}: {
  steps: Step[];
  onChange: (next: Step[]) => void;
}) {
  function update(i: number, patch: Partial<Step>) {
    const next = [...steps];
    next[i] = { ...next[i], ...patch };
    onChange(next);
  }
  function remove(i: number) {
    onChange(steps.filter((_, idx) => idx !== i));
  }
  function swap(i: number, j: number) {
    if (j < 0 || j >= steps.length) return;
    const next = [...steps];
    [next[i], next[j]] = [next[j], next[i]];
    onChange(next);
  }
  function add() {
    onChange([...steps, { event: 'charge', outcome: 'success' }]);
  }

  return (
    <div className="space-y-2">
      {steps.map((s, i) => (
        <div key={i} className="flex items-center gap-2 border rounded-md p-2">
          <span className="text-xs text-muted-foreground font-mono w-6">#{i + 1}</span>
          <Select.Root
            value={s.event}
            onValueChange={(v) => update(i, { event: v as EventType })}
          >
            <Select.Trigger className="w-28 text-xs" />
            <Select.Content>
              {EVENT_TYPES.map((e) => (
                <Select.Item key={e} value={e} className="text-xs">
                  {e}
                </Select.Item>
              ))}
            </Select.Content>
          </Select.Root>
          <div className="flex-1">
            <OutcomePicker value={s.outcome} onChange={(v) => update(i, { outcome: v })} />
          </div>
          <TextField.Root
            placeholder="code (opt)"
            value={s.code ?? ''}
            onChange={(e) => update(i, { code: e.target.value || undefined })}
            className="w-40 text-xs font-mono"
          />
          <Button size="1" variant="ghost" color="gray" onClick={() => swap(i, i - 1)} type="button">
            <ArrowUpIcon />
          </Button>
          <Button size="1" variant="ghost" color="gray" onClick={() => swap(i, i + 1)} type="button">
            <ArrowDownIcon />
          </Button>
          <Button size="1" variant="ghost" color="gray" onClick={() => remove(i)} type="button">
            <TrashIcon />
          </Button>
        </div>
      ))}
      <Button variant="outline" size="2" onClick={add} type="button">
        <PlusIcon />
        Add step
      </Button>
    </div>
  );
}
