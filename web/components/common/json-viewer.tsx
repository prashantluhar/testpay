export function JsonViewer({ value }: { value: unknown }) {
  const text = JSON.stringify(value, null, 2);
  return (
    <pre className="text-xs font-mono bg-muted p-3 rounded-md overflow-auto max-h-[600px] whitespace-pre">
      {text}
    </pre>
  );
}
