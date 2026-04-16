'use client';
import { useRouter } from 'next/navigation';
import { useForm, Controller } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { toast } from 'sonner';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Checkbox } from '@/components/ui/checkbox';
import { ScenarioStepEditor, type Step } from './scenario-step-editor';
import { JsonViewer } from '@/components/common/json-viewer';
import { CopyButton } from '@/components/common/copy-button';
import { scenarioSchema, type ScenarioInput } from '@/lib/schemas';
import { api } from '@/lib/api';
import type { Scenario } from '@/lib/types';
import { mutate } from 'swr';

export function ScenarioForm({ initial }: { initial?: Scenario }) {
  const router = useRouter();
  const isEdit = !!initial;

  const form = useForm<ScenarioInput>({
    resolver: zodResolver(scenarioSchema),
    defaultValues: initial
      ? {
          Name: initial.Name,
          Description: initial.Description,
          Gateway: initial.Gateway,
          Steps: initial.Steps.length
            ? (initial.Steps as Step[])
            : [{ event: 'charge', outcome: 'success' }],
          WebhookDelayMs: initial.WebhookDelayMs,
          IsDefault: initial.IsDefault,
        }
      : {
          Name: '',
          Description: '',
          Gateway: 'stripe',
          Steps: [{ event: 'charge', outcome: 'success' }],
          WebhookDelayMs: 0,
          IsDefault: false,
        },
  });

  const values = form.watch();

  async function onSubmit(data: ScenarioInput) {
    try {
      if (isEdit && initial) {
        await api(`/api/scenarios/${initial.ID}`, { method: 'PUT', body: JSON.stringify(data) });
        toast.success('Scenario saved');
      } else {
        await api(`/api/scenarios`, { method: 'POST', body: JSON.stringify(data) });
        toast.success('Scenario created');
      }
      mutate('/api/scenarios');
      router.push('/scenarios');
    } catch (e) {
      toast.error(e instanceof Error ? e.message : 'save failed');
    }
  }

  const previewJson = { ...values, ID: initial?.ID ?? 'new' };

  return (
    <form onSubmit={form.handleSubmit(onSubmit)} className="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <Card>
        <CardHeader>
          <CardTitle className="text-base">{isEdit ? 'Edit scenario' : 'New scenario'}</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div>
            <Label htmlFor="name">Name</Label>
            <Input id="name" {...form.register('Name')} />
          </div>
          <div>
            <Label htmlFor="desc">Description</Label>
            <Input id="desc" {...form.register('Description')} />
          </div>
          <div>
            <Label>Gateway</Label>
            <Controller
              control={form.control}
              name="Gateway"
              render={({ field }) => (
                <Select value={field.value} onValueChange={field.onChange}>
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="stripe">stripe</SelectItem>
                    <SelectItem value="razorpay">razorpay</SelectItem>
                    <SelectItem value="agnostic">agnostic</SelectItem>
                  </SelectContent>
                </Select>
              )}
            />
          </div>
          <div>
            <Label>Steps</Label>
            <Controller
              control={form.control}
              name="Steps"
              render={({ field }) => (
                <ScenarioStepEditor
                  steps={field.value as Step[]}
                  onChange={(next) => field.onChange(next)}
                />
              )}
            />
          </div>
          <div>
            <Label htmlFor="delay">Webhook delay (ms)</Label>
            <Input
              id="delay"
              type="number"
              {...form.register('WebhookDelayMs', { valueAsNumber: true })}
            />
          </div>
          <div className="flex items-center gap-2">
            <Controller
              control={form.control}
              name="IsDefault"
              render={({ field }) => (
                <Checkbox checked={field.value} onCheckedChange={field.onChange} id="default" />
              )}
            />
            <Label htmlFor="default">Set as default for this workspace</Label>
          </div>
          <div className="flex gap-2">
            <Button type="submit" disabled={form.formState.isSubmitting}>
              {isEdit ? 'Save changes' : 'Create'}
            </Button>
            <Button type="button" variant="ghost" onClick={() => router.push('/scenarios')}>
              Cancel
            </Button>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <CardTitle className="text-base">Preview</CardTitle>
          <CopyButton value={JSON.stringify(previewJson, null, 2)} />
        </CardHeader>
        <CardContent>
          <JsonViewer value={previewJson} />
        </CardContent>
      </Card>
    </form>
  );
}
