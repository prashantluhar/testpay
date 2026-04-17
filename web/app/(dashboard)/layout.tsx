'use client';
import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useMe } from '@/lib/hooks';
import { Sidebar } from '@/components/shell/sidebar';
import { Topbar } from '@/components/shell/topbar';
import { ApiError } from '@/lib/api';

export default function DashboardLayout({ children }: { children: React.ReactNode }) {
  const router = useRouter();
  const { data, error, isLoading } = useMe();

  useEffect(() => {
    if (error instanceof ApiError && error.status === 401) {
      router.push('/login');
    }
  }, [error, router]);

  if (isLoading) {
    return (
      <div className="min-h-screen grid place-items-center text-muted-foreground">
        <div className="flex flex-col items-center gap-3">
          <div className="h-6 w-6 border-2 border-muted-foreground/40 border-t-emerald-500 rounded-full animate-spin" />
          <div className="text-sm">Loading…</div>
        </div>
      </div>
    );
  }

  // No data (either 401 or other error) → block render; effect above redirects.
  if (!data || error) return null;
  // Extra guard: the Me endpoint should not return user=null for authenticated
  // callers. If it does, treat as unauthenticated and redirect.
  if (!data.user) {
    router.push('/login');
    return null;
  }

  return (
    <div className="h-screen flex bg-background overflow-hidden">
      <Sidebar user={data.user} workspace={data.workspace} />
      <div className="flex-1 flex flex-col min-w-0">
        <Topbar user={data.user} />
        <main className="flex-1 overflow-y-auto">
          <div className="max-w-6xl mx-auto px-6 py-6 w-full">{children}</div>
        </main>
      </div>
    </div>
  );
}
