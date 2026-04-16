'use client';
import { useSearchParams } from 'next/navigation';
import { Suspense } from 'react';
import { useScenario } from '@/lib/hooks';
import { ScenarioForm } from '@/components/scenarios/scenario-form';

function EditScenarioInner() {
  const sp = useSearchParams();
  const id = sp.get('id');
  const { data, error } = useScenario(id);

  if (!id) return <div className="text-muted-foreground">No scenario id in URL.</div>;
  if (error) return <div className="text-destructive">Not found.</div>;
  if (!data) return <div className="text-muted-foreground">Loading…</div>;
  return <ScenarioForm initial={data} />;
}

export default function EditScenarioPage() {
  return (
    <Suspense fallback={<div className="text-muted-foreground">Loading…</div>}>
      <EditScenarioInner />
    </Suspense>
  );
}
