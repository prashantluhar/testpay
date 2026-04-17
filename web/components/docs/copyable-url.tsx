'use client';
import { CopyButton } from '@/components/common/copy-button';

// Compact inline URL with copy button — matches the Overview page hero's
// base-URL pill so the docs visually agree with the dashboard.
export function CopyableUrl({ url, label }: { url: string; label?: string }) {
  return (
    <div className="flex items-center gap-2 bg-[var(--gray-a2)] border border-[var(--gray-a5)] rounded-md px-3 py-2 my-2">
      {label ? (
        <span className="text-[10px] uppercase tracking-wider text-[var(--gray-11)] shrink-0">
          {label}
        </span>
      ) : null}
      <code className="flex-1 truncate font-mono text-sm text-[var(--gray-12)]">{url}</code>
      <CopyButton value={url} label="" />
    </div>
  );
}
