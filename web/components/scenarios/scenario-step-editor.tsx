'use client';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { OutcomePicker } from './outcome-picker';
import { ArrowUp, ArrowDown, Trash2, Plus } from 'lucide-react';
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
          <Select value={s.event} onValueChange={(v) => update(i, { event: v as EventType })}>
            <SelectTrigger className="w-28 text-xs">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {EVENT_TYPES.map((e) => (
                <SelectItem key={e} value={e} className="text-xs">
                  {e}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          <div className="flex-1">
            <OutcomePicker value={s.outcome} onChange={(v) => update(i, { outcome: v })} />
          </div>
          <Input
            placeholder="code (opt)"
            value={s.code ?? ''}
            onChange={(e) => update(i, { code: e.target.value || undefined })}
            className="w-40 text-xs font-mono"
          />
          <Button size="sm" variant="ghost" onClick={() => swap(i, i - 1)} type="button">
            <ArrowUp className="h-4 w-4" />
          </Button>
          <Button size="sm" variant="ghost" onClick={() => swap(i, i + 1)} type="button">
            <ArrowDown className="h-4 w-4" />
          </Button>
          <Button size="sm" variant="ghost" onClick={() => remove(i)} type="button">
            <Trash2 className="h-4 w-4" />
          </Button>
        </div>
      ))}
      <Button variant="outline" size="sm" onClick={add} type="button">
        <Plus className="h-4 w-4 mr-2" />
        Add step
      </Button>
    </div>
  );
}
