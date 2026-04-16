'use client';
import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useMe } from '@/lib/hooks';
import { MODE } from '@/lib/types';
import { Sidebar } from '@/components/shell/sidebar';
import { Topbar } from '@/components/shell/topbar';
import { ApiError } from '@/lib/api';

export default function DashboardLayout({ children }: { children: React.ReactNode }) {
  const router = useRouter();
  const { data, error, isLoading } = useMe();

  useEffect(() => {
    if (MODE === 'hosted' && error instanceof ApiError && error.status === 401) {
      router.push('/login');
    }
  }, [error, router]);

  if (isLoading) {
    return (
      <div className="min-h-screen grid place-items-center text-muted-foreground">Loading…</div>
    );
  }

  if (MODE === 'hosted' && (error || !data)) return null;
  if (!data) return null;

  return (
    <div className="min-h-screen flex">
      <Sidebar user={data.user} workspace={data.workspace} />
      <div className="flex-1 flex flex-col">
        <Topbar workspace={data.workspace} />
        <main className="flex-1 p-6 overflow-auto">{children}</main>
      </div>
    </div>
  );
}
