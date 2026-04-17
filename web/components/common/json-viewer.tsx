export function JsonViewer({ value }: { value: unknown }) {
  const text = JSON.stringify(value, null, 2);
  return (
    <pre className="font-mono text-[13px] leading-relaxed text-[var(--gray-12)] bg-[var(--gray-a3)] border border-[var(--gray-a5)] p-4 rounded-md overflow-auto max-h-[620px] whitespace-pre">
      {text}
    </pre>
  );
}

// Compact key/value list — used by drawers to show metadata above the JSON
// tabs. Labels are small-caps gray; values are mono and readable.
export function KeyValueGrid({
  items,
}: {
  items: Array<{ label: string; value: React.ReactNode }>;
}) {
  return (
    <div className="grid grid-cols-2 gap-x-6 gap-y-4">
      {items.map(({ label, value }) => (
        <div key={label} className="min-w-0">
          <div className="text-[11px] uppercase tracking-wider text-[var(--gray-11)] mb-1">
            {label}
          </div>
          <div className="font-mono text-sm text-[var(--gray-12)] break-all">{value}</div>
        </div>
      ))}
    </div>
  );
}
