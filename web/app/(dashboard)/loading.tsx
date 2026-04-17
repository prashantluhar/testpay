import { PageLoader } from '@/components/common/page-loader';

// Route-level Suspense fallback rendered by Next.js while a dashboard
// segment is loading.
export default function DashboardLoading() {
  return <PageLoader />;
}
