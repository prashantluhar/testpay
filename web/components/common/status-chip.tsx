export function StatusChip({ status }: { status: number }) {
  const color =
    status >= 500
      ? 'bg-red-500/10 text-red-500'
      : status >= 400
        ? 'bg-amber-500/10 text-amber-500'
        : status >= 200
          ? 'bg-emerald-500/10 text-emerald-500'
          : 'bg-muted text-muted-foreground';
  return (
    <span className={`inline-flex items-center rounded px-1.5 py-0.5 text-xs font-mono ${color}`}>
      {status}
    </span>
  );
}
