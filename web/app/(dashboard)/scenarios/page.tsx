'use client';
import { useState } from 'react';
import Link from 'next/link';
import { Plus, Play, Pencil, Trash2 } from 'lucide-react';
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
import { api } from '@/lib/api';
import { mutate } from 'swr';

export default function ScenariosPage() {
  const { data, error } = useScenarios();
  const [deleteId, setDeleteId] = useState<string | null>(null);

  async function runScenario(id: string) {
    try {
      await api(`/api/scenarios/${id}/run`, { method: 'POST' });
      toast.success('Scenario running');
      mutate('/api/scenarios');
    } catch {
      toast.error('Failed to run scenario');
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
          <p className="text-sm text-muted-foreground">Named, replayable failure sequences.</p>
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
              <TableHead>Default</TableHead>
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
                <TableCell className="text-xs">{s.is_default ? 'yes' : ''}</TableCell>
                <TableCell className="text-right space-x-1">
                  <Button size="sm" variant="ghost" onClick={() => runScenario(s.id)}>
                    <Play className="h-4 w-4" />
                  </Button>
                  <Button size="sm" variant="ghost" asChild>
                    <Link href={`/scenarios/edit?id=${s.id}`}>
                      <Pencil className="h-4 w-4" />
                    </Link>
                  </Button>
                  <Button size="sm" variant="ghost" onClick={() => setDeleteId(s.id)}>
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
