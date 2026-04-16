'use client';
import { useParams } from 'next/navigation';
import { useScenario } from '@/lib/hooks';
import { ScenarioForm } from '@/components/scenarios/scenario-form';

export function EditScenarioClient() {
  const params = useParams<{ id: string }>();
  const { data, error } = useScenario(params.id && params.id !== '_' ? params.id : null);

  if (!params.id || params.id === '_') {
    return <div className="text-muted-foreground">Loading scenario…</div>;
  }
  if (error) return <div className="text-destructive">Not found.</div>;
  if (!data) return <div className="text-muted-foreground">Loading…</div>;
  return <ScenarioForm initial={data} />;
}
