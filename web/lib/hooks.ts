import useSWR, { SWRConfiguration } from 'swr';
import { swrFetcher } from './api';
import type { AuthResponse, RequestLog, Scenario, Workspace, WebhookLog } from './types';

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
