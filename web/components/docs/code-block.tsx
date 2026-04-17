'use client';
import { CopyButton } from '@/components/common/copy-button';

// Minimal monospace code block. Keeps the same tokens JsonViewer uses so
// docs pages look consistent with the existing Log/Webhook drawers.
export function CodeBlock({
  children,
  language,
  copyable = true,
}: {
  children: string;
  language?: string;
  copyable?: boolean;
}) {
  return (
    <div className="relative group">
      {language ? (
        <div className="absolute top-2 right-12 text-[10px] uppercase tracking-wider text-[var(--gray-11)] font-mono pointer-events-none">
          {language}
        </div>
      ) : null}
      {copyable ? (
        <div className="absolute top-1.5 right-1.5 opacity-60 hover:opacity-100 transition-opacity">
          <CopyButton value={children} label="" />
        </div>
      ) : null}
      <pre className="font-mono text-[13px] leading-relaxed text-[var(--gray-12)] bg-[var(--gray-a3)] border border-[var(--gray-a5)] p-4 rounded-md overflow-auto max-h-[620px] whitespace-pre">
        {children}
      </pre>
    </div>
  );
}
