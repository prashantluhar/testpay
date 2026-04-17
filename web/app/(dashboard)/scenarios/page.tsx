'use client';
import { useState } from 'react';
import Link from 'next/link';
import {
  PlusIcon,
  PlayIcon,
  Pencil1Icon,
  TrashIcon,
  CheckCircledIcon,
  CircleIcon,
} from '@radix-ui/react-icons';
import { toast } from 'sonner';
import { Button, Flex, Heading, Table, Text } from '@radix-ui/themes';
import { GatewayBadge } from '@/components/common/gateway-badge';
import { ConfirmModal } from '@/components/common/confirm-modal';
import { useScenarios } from '@/lib/hooks';
import { api, ApiError } from '@/lib/api';
import { mutate } from 'swr';
import type { Scenario } from '@/lib/types';

const SESSION_TTL_SECONDS = 300; // 5-minute activation

export default function ScenariosPage() {
  const { data, error } = useScenarios();
  const [deleteId, setDeleteId] = useState<string | null>(null);

  // "Run" = pin the scenario to this workspace's mock endpoint for 5 minutes.
  // All incoming mock requests during that window execute this scenario.
  async function activate(id: string, name: string) {
    try {
      await api('/api/sessions', {
        method: 'POST',
        body: JSON.stringify({ scenario_id: id, ttl_seconds: SESSION_TTL_SECONDS }),
      });
      toast.success(
        `"${name}" is now active for ${SESSION_TTL_SECONDS / 60} minutes — mock requests will use this scenario.`,
      );
    } catch (e) {
      toast.error(e instanceof ApiError ? e.message : 'Activation failed');
    }
  }

  // Flip is_default on/off.
  async function toggleDefault(s: Scenario) {
    try {
      await api(`/api/scenarios/${s.id}`, {
        method: 'PUT',
        body: JSON.stringify({ ...s, is_default: !s.is_default }),
      });
      toast.success(s.is_default ? 'Cleared default' : `Default is now "${s.name}"`);
      mutate('/api/scenarios');
    } catch (e) {
      toast.error(e instanceof ApiError ? e.message : 'Failed to toggle default');
    }
  }

  async function deleteScenario(id: string) {
    try {
      await api(`/api/scenarios/${id}`, { method: 'DELETE' });
      toast.success('Scenario deleted');
      mutate('/api/scenarios');
    } catch {
      toast.error('Failed to delete');
    }
    setDeleteId(null);
  }

  return (
    <div className="space-y-6">
      <Flex align="center" justify="between">
        <div>
          <Heading size="6">Scenarios</Heading>
          <Text size="2" color="gray" as="p">
            Named, replayable failure sequences. Click ▶ to activate one for 5 minutes, or toggle
            the ● to pin it as the workspace default.
          </Text>
        </div>
        <Button asChild>
          <Link href="/scenarios/new">
            <PlusIcon />
            New scenario
          </Link>
        </Button>
      </Flex>

      {error && (
        <Text color="red" size="2">
          Failed to load scenarios.
        </Text>
      )}

      <div className="border rounded-md">
        <Table.Root>
          <Table.Header>
            <Table.Row>
              <Table.ColumnHeaderCell>Name</Table.ColumnHeaderCell>
              <Table.ColumnHeaderCell>Gateway</Table.ColumnHeaderCell>
              <Table.ColumnHeaderCell>Steps</Table.ColumnHeaderCell>
              <Table.ColumnHeaderCell className="w-24">Default</Table.ColumnHeaderCell>
              <Table.ColumnHeaderCell className="text-right">Actions</Table.ColumnHeaderCell>
            </Table.Row>
          </Table.Header>
          <Table.Body>
            {data?.map((s) => (
              <Table.Row key={s.id}>
                <Table.Cell className="font-medium">{s.name}</Table.Cell>
                <Table.Cell>
                  <GatewayBadge gateway={s.gateway} />
                </Table.Cell>
                <Table.Cell className="font-mono text-xs">{s.steps?.length ?? 0}</Table.Cell>
                <Table.Cell>
                  <Button
                    size="1"
                    variant="ghost"
                    color="gray"
                    onClick={() => toggleDefault(s)}
                    title={s.is_default ? 'Clear default' : 'Set as default'}
                  >
                    {s.is_default ? (
                      <CheckCircledIcon className="text-emerald-500" />
                    ) : (
                      <CircleIcon className="text-muted-foreground" />
                    )}
                  </Button>
                </Table.Cell>
                <Table.Cell className="text-right">
                  <Flex gap="1" justify="end">
                    <Button
                      size="1"
                      variant="ghost"
                      color="gray"
                      onClick={() => activate(s.id, s.name)}
                      title="Activate for 5 minutes"
                    >
                      <PlayIcon />
                    </Button>
                    <Button size="1" variant="ghost" color="gray" asChild title="Edit">
                      <Link href={`/scenarios/edit?id=${s.id}`}>
                        <Pencil1Icon />
                      </Link>
                    </Button>
                    <Button
                      size="1"
                      variant="ghost"
                      color="gray"
                      onClick={() => setDeleteId(s.id)}
                      title="Delete"
                    >
                      <TrashIcon />
                    </Button>
                  </Flex>
                </Table.Cell>
              </Table.Row>
            ))}
            {data && data.length === 0 && (
              <Table.Row>
                <Table.Cell colSpan={5} className="text-center text-muted-foreground py-8">
                  No scenarios yet. Create one to get started.
                </Table.Cell>
              </Table.Row>
            )}
          </Table.Body>
        </Table.Root>
      </div>

      <ConfirmModal
        open={!!deleteId}
        title="Delete scenario?"
        description="This cannot be undone."
        confirmLabel="Delete"
        destructive
        onConfirm={() => {
          if (deleteId) deleteScenario(deleteId);
        }}
        onClose={() => setDeleteId(null)}
      />
    </div>
  );
}
