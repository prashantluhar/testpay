'use client';
import { Text } from '@radix-ui/themes';
import { useWorkspaceUsage } from '@/lib/hooks';

// Small "used_today / cap" pill that surfaces the daily request quota.
// Only visible when the workspace actually has a cap set — signed-up users
// without a per-row cap see nothing so the topbar stays uncluttered.
//
// Color scales with how close to the cap the caller is:
//   < 70%     — muted
//   70%-90%   — amber
//   >= 90%    — red
// so pilots notice before they hit 429s.
export function QuotaPill() {
  const { data } = useWorkspaceUsage();
  if (!data || data.cap === null || data.cap === 0) return null;

  const { used_today, cap } = data;
  const pct = Math.min(100, Math.round((used_today / cap) * 100));
  const tone = pct >= 90 ? 'red' : pct >= 70 ? 'amber' : 'muted';
  const bg =
    tone === 'red'
      ? 'bg-[var(--red-a3)] text-[var(--red-11)] border-[var(--red-a6)]'
      : tone === 'amber'
        ? 'bg-[var(--amber-a3)] text-[var(--amber-11)] border-[var(--amber-a6)]'
        : 'bg-[var(--gray-a3)] text-[var(--gray-11)] border-[var(--gray-a5)]';

  return (
    <div
      className={`hidden md:flex items-center gap-1.5 border rounded-full px-2.5 py-1 text-xs font-mono tabular-nums ${bg}`}
      title={`${used_today} of ${cap} mock requests used in the last 24h. Resets rolling.`}
    >
      <span className="inline-block h-1.5 w-1.5 rounded-full bg-current opacity-80" />
      <Text size="1" weight="medium" className="tabular-nums">
        {used_today} / {cap}
      </Text>
    </div>
  );
}
