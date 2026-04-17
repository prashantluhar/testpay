import { MODE } from '@/lib/types';

// Resolve the user-facing API base URL. Mirrors the logic the Overview hero
// uses so the docs show whatever the dashboard would point callers at.
export function getApiBaseUrl(): string {
  if (MODE === 'local') return 'http://localhost:7700';
  return process.env.NEXT_PUBLIC_API_BASE || 'https://testpay-przk.onrender.com';
}
