export class ApiError extends Error {
  constructor(
    public status: number,
    message: string,
    public body?: unknown,
  ) {
    super(message);
    this.name = 'ApiError';
  }
}

const DEFAULT_API_BASE =
  typeof window !== 'undefined' && window.location.origin === 'http://localhost:7701'
    ? 'http://localhost:7700'
    : '';

const API_BASE = process.env.NEXT_PUBLIC_API_BASE || DEFAULT_API_BASE;

export async function api<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    credentials: 'include',
    headers: { 'Content-Type': 'application/json', ...(init?.headers || {}) },
    ...init,
  });
  const text = await res.text();
  let body: unknown = undefined;
  if (text) {
    try {
      body = JSON.parse(text);
    } catch {
      body = text;
    }
  }
  if (!res.ok) {
    const msg = (body as { error?: string })?.error || res.statusText || 'request failed';
    throw new ApiError(res.status, msg, body);
  }
  return body as T;
}

export const swrFetcher = <T>(path: string) => api<T>(path);
