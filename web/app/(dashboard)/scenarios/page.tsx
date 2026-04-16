'use client';
import { useState } from 'react';
import Link from 'next/link';
import { Plus, Play, Pencil, Trash2, CheckCircle2, Circle } from 'lucide-react';
import { toast } from 'sonner';
import { Button } from '@/components/ui/button';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
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

  // Flip is_default on/off. When on, this scenario applies to every mock
  // request in the workspace until flipped back.
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
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold">Scenarios</h1>
          <p className="text-sm text-muted-foreground">
            Named, replayable failure sequences. Click ▶ to activate one for 5 minutes, or toggle
            the ● to pin it as the workspace default.
          </p>
        </div>
        <Button asChild>
          <Link href="/scenarios/new">
            <Plus className="h-4 w-4 mr-2" />
            New scenario
          </Link>
        </Button>
      </div>

      {error && <div className="text-destructive text-sm">Failed to load scenarios.</div>}

      <div className="border rounded-md">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>Gateway</TableHead>
              <TableHead>Steps</TableHead>
              <TableHead className="w-24">Default</TableHead>
              <TableHead className="text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {data?.map((s) => (
              <TableRow key={s.id}>
                <TableCell className="font-medium">{s.name}</TableCell>
                <TableCell>
                  <GatewayBadge gateway={s.gateway} />
                </TableCell>
                <TableCell className="font-mono text-xs">{s.steps?.length ?? 0}</TableCell>
                <TableCell>
                  <Button
                    size="sm"
                    variant="ghost"
                    onClick={() => toggleDefault(s)}
                    title={s.is_default ? 'Clear default' : 'Set as default'}
                  >
                    {s.is_default ? (
                      <CheckCircle2 className="h-4 w-4 text-emerald-500" />
                    ) : (
                      <Circle className="h-4 w-4 text-muted-foreground" />
                    )}
                  </Button>
                </TableCell>
                <TableCell className="text-right space-x-1">
                  <Button
                    size="sm"
                    variant="ghost"
                    onClick={() => activate(s.id, s.name)}
                    title="Activate for 5 minutes"
                  >
                    <Play className="h-4 w-4" />
                  </Button>
                  <Button size="sm" variant="ghost" asChild title="Edit">
                    <Link href={`/scenarios/edit?id=${s.id}`}>
                      <Pencil className="h-4 w-4" />
                    </Link>
                  </Button>
                  <Button
                    size="sm"
                    variant="ghost"
                    onClick={() => setDeleteId(s.id)}
                    title="Delete"
                  >
                    <Trash2 className="h-4 w-4" />
                  </Button>
                </TableCell>
              </TableRow>
            ))}
            {data && data.length === 0 && (
              <TableRow>
                <TableCell colSpan={5} className="text-center text-muted-foreground py-8">
                  No scenarios yet. Create one to get started.
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
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
