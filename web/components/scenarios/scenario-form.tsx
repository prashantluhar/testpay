'use client';
import { useRouter } from 'next/navigation';
import { useForm, Controller } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { toast } from 'sonner';
import {
  Box,
  Button,
  Card,
  Checkbox,
  Flex,
  Heading,
  Select,
  Text,
  TextField,
} from '@radix-ui/themes';
import { ScenarioStepEditor, type Step } from './scenario-step-editor';
import { JsonViewer } from '@/components/common/json-viewer';
import { CopyButton } from '@/components/common/copy-button';
import { scenarioSchema, type ScenarioInput } from '@/lib/schemas';
import { api } from '@/lib/api';
import { useGateways } from '@/lib/hooks';
import type { Scenario } from '@/lib/types';
import { mutate } from 'swr';

export function ScenarioForm({ initial }: { initial?: Scenario }) {
  const router = useRouter();
  const isEdit = !!initial;
  const { data: gateways = [] } = useGateways();

  const form = useForm<ScenarioInput>({
    resolver: zodResolver(scenarioSchema),
    defaultValues: initial
      ? {
          name: initial.name,
          description: initial.description,
          gateway: initial.gateway,
          steps: initial.steps?.length
            ? (initial.steps as Step[])
            : [{ event: 'charge', outcome: 'success' }],
          webhook_delay_ms: initial.webhook_delay_ms,
          is_default: initial.is_default,
        }
      : {
          name: '',
          description: '',
          gateway: 'stripe',
          steps: [{ event: 'charge', outcome: 'success' }],
          webhook_delay_ms: 0,
          is_default: false,
        },
  });

  const values = form.watch();

  async function onSubmit(data: ScenarioInput) {
    try {
      if (isEdit && initial) {
        await api(`/api/scenarios/${initial.id}`, { method: 'PUT', body: JSON.stringify(data) });
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

  const previewJson = { ...values, id: initial?.id ?? 'new' };

  return (
    <form onSubmit={form.handleSubmit(onSubmit)} className="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <Card>
        <Box p="2">
          <Heading size="4" mb="4">
            {isEdit ? 'Edit scenario' : 'New scenario'}
          </Heading>
          <Flex direction="column" gap="4">
            <Box>
              <Text as="label" size="2" weight="medium" htmlFor="name">
                Name
              </Text>
              <TextField.Root id="name" mt="1" {...form.register('name')} />
            </Box>
            <Box>
              <Text as="label" size="2" weight="medium" htmlFor="desc">
                Description
              </Text>
              <TextField.Root id="desc" mt="1" {...form.register('description')} />
            </Box>
            <Box>
              <Text as="label" size="2" weight="medium">
                Gateway
              </Text>
              <Controller
                control={form.control}
                name="gateway"
                render={({ field }) => (
                  <Select.Root value={field.value} onValueChange={field.onChange}>
                    <Select.Trigger className="mt-1 w-full" />
                    <Select.Content>
                      {gateways.length === 0 ? (
                        <Select.Item value={field.value || 'stripe'}>
                          Loading…
                        </Select.Item>
                      ) : (
                        gateways.map((g) => (
                          <Select.Item key={g} value={g}>
                            {g}
                          </Select.Item>
                        ))
                      )}
                    </Select.Content>
                  </Select.Root>
                )}
              />
            </Box>
            <Box>
              <Text as="label" size="2" weight="medium">
                Steps
              </Text>
              <Box mt="1">
                <Controller
                  control={form.control}
                  name="steps"
                  render={({ field }) => (
                    <ScenarioStepEditor
                      steps={field.value as Step[]}
                      onChange={(next) => field.onChange(next)}
                    />
                  )}
                />
              </Box>
            </Box>
            <Box>
              <Text as="label" size="2" weight="medium" htmlFor="delay">
                Webhook delay (ms)
              </Text>
              <TextField.Root
                id="delay"
                type="number"
                mt="1"
                {...form.register('webhook_delay_ms', { valueAsNumber: true })}
              />
            </Box>
            <Flex align="center" gap="2">
              <Controller
                control={form.control}
                name="is_default"
                render={({ field }) => (
                  <Checkbox
                    checked={field.value}
                    onCheckedChange={(c) => field.onChange(!!c)}
                    id="default"
                  />
                )}
              />
              <Text as="label" size="2" htmlFor="default">
                Set as default for this workspace
              </Text>
            </Flex>
            <Flex gap="2">
              <Button type="submit" disabled={form.formState.isSubmitting}>
                {isEdit ? 'Save changes' : 'Create'}
              </Button>
              <Button
                type="button"
                variant="soft"
                color="gray"
                onClick={() => router.push('/scenarios')}
              >
                Cancel
              </Button>
            </Flex>
          </Flex>
        </Box>
      </Card>

      <Card>
        <Box p="2">
          <Flex align="center" justify="between" mb="4">
            <Heading size="4">Preview</Heading>
            <CopyButton value={JSON.stringify(previewJson, null, 2)} />
          </Flex>
          <JsonViewer value={previewJson} />
        </Box>
      </Card>
    </form>
  );
}
