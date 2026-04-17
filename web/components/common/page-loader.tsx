'use client';
import { Spinner } from './spinner';

// Full-container centered spinner. Used as a Suspense fallback or for
// route-level loading states.
export function PageLoader({ label = 'Loading…' }: { label?: string }) {
  return (
    <div className="min-h-[40vh] w-full grid place-items-center animate-in fade-in duration-300">
      <div className="flex flex-col items-center gap-3 text-muted-foreground">
        <Spinner size="large" />
        {label && <div className="text-sm">{label}</div>}
      </div>
    </div>
  );
}
