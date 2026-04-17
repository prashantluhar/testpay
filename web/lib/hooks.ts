import useSWR, { SWRConfiguration } from 'swr';
import { useEffect, useRef, useState } from 'react';
import { swrFetcher } from './api';
import type { AuthResponse, RequestLog, Scenario, Workspace, WebhookLog } from './types';

// Eases a displayed number from 0 → `target` over `duration` ms on first mount.
// After first mount, subsequent target changes snap directly to the new value
// (so polling updates don't re-animate). Callers using non-numeric values pass
// through unchanged via the consumer's logic.
export function useCountUp(target: number, duration = 600) {
  const [value, setValue] = useState(0);
  const mountedRef = useRef(false);
  const rafRef = useRef<number | null>(null);

  useEffect(() => {
    if (mountedRef.current) {
      setValue(target);
      return;
    }
    mountedRef.current = true;
    const start = performance.now();
    const from = 0;
    const delta = target - from;
    const tick = (now: number) => {
      const t = Math.min(1, (now - start) / duration);
      // easeOutCubic
      const eased = 1 - Math.pow(1 - t, 3);
      setValue(Math.round(from + delta * eased));
      if (t < 1) {
        rafRef.current = requestAnimationFrame(tick);
      }
    };
    rafRef.current = requestAnimationFrame(tick);
    return () => {
      if (rafRef.current != null) cancelAnimationFrame(rafRef.current);
    };
    // Intentionally run once on mount only.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  return value;
}

export function useMe() {
  return useSWR<AuthResponse>('/api/auth/me', swrFetcher, {
    revalidateOnFocus: false,
    shouldRetryOnError: false,
  });
}

export function useWorkspace() {
  return useSWR<Workspace>('/api/workspace', swrFetcher);
}

export function useScenarios() {
  return useSWR<Scenario[]>('/api/scenarios', swrFetcher);
}

export function useScenario(id: string | null) {
  return useSWR<Scenario>(id ? `/api/scenarios/${id}` : null, swrFetcher);
}

interface LogFilters {
  limit?: number;
  offset?: number;
  pollInterval?: number;
}
export function useLogs(filters: LogFilters = {}) {
  const params = new URLSearchParams();
  if (filters.limit) params.set('limit', String(filters.limit));
  if (filters.offset) params.set('offset', String(filters.offset));
  const qs = params.toString();
  const opts: SWRConfiguration = filters.pollInterval
    ? { refreshInterval: filters.pollInterval }
    : {};
  return useSWR<RequestLog[]>(`/api/logs${qs ? '?' + qs : ''}`, swrFetcher, opts);
}

export function useLog(id: string | null) {
  return useSWR<{ request: RequestLog; webhook: WebhookLog | null }>(
    id ? `/api/logs/${id}` : null,
    swrFetcher,
  );
}

export function useWebhooks(filters: { limit?: number; offset?: number; pollInterval?: number } = {}) {
  const params = new URLSearchParams();
  if (filters.limit) params.set('limit', String(filters.limit));
  if (filters.offset) params.set('offset', String(filters.offset));
  const qs = params.toString();
  const opts: SWRConfiguration = filters.pollInterval
    ? { refreshInterval: filters.pollInterval }
    : {};
  return useSWR<WebhookLog[]>(`/api/webhooks${qs ? '?' + qs : ''}`, swrFetcher, opts);
}

export function useWebhook(id: string | null) {
  return useSWR<WebhookLog>(id ? `/api/webhooks/${id}` : null, swrFetcher);
}

export function useGateways() {
  return useSWR<string[]>('/api/gateways', swrFetcher, { revalidateOnFocus: false });
}

export interface WorkspaceUsage {
  used_today: number;
  cap: number | null;
}

// Polls the workspace's rolling-24h request count every 30s so the quota
// pill in the topbar stays fresh without a full page reload.
export function useWorkspaceUsage() {
  return useSWR<WorkspaceUsage>('/api/workspace/usage', swrFetcher, {
    refreshInterval: 30_000,
    revalidateOnFocus: true,
  });
}
