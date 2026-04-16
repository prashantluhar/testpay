import { describe, it, expect, vi, beforeEach } from 'vitest';
import { api, ApiError } from '@/lib/api';

describe('api()', () => {
  beforeEach(() => {
    global.fetch = vi.fn();
  });

  it('returns parsed JSON on 2xx', async () => {
    (global.fetch as unknown as ReturnType<typeof vi.fn>).mockResolvedValue({
      ok: true,
      status: 200,
      text: async () => JSON.stringify({ hello: 'world' }),
    });
    const res = await api<{ hello: string }>('/api/test');
    expect(res.hello).toBe('world');
  });

  it('throws ApiError on non-2xx', async () => {
    (global.fetch as unknown as ReturnType<typeof vi.fn>).mockResolvedValue({
      ok: false,
      status: 401,
      statusText: 'Unauthorized',
      text: async () => JSON.stringify({ error: 'unauthorized' }),
    });
    await expect(api('/api/test')).rejects.toBeInstanceOf(ApiError);
  });

  it('includes credentials by default', async () => {
    (global.fetch as unknown as ReturnType<typeof vi.fn>).mockResolvedValue({
      ok: true,
      status: 204,
      text: async () => '',
    });
    await api('/api/test');
    expect(global.fetch).toHaveBeenCalledWith(
      expect.any(String),
      expect.objectContaining({ credentials: 'include' }),
    );
  });
});
