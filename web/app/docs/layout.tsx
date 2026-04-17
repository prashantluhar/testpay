import Link from 'next/link';
import { Flex, Heading } from '@radix-ui/themes';
import { LightningBoltIcon, PersonIcon } from '@radix-ui/react-icons';
import { DocsSidebar } from '@/components/docs/docs-sidebar';

// Docs live at a PUBLIC route (not nested under (dashboard)) so any visitor
// — including someone on /signup or /login — can link to the same pages
// without needing an account.
//
// Two-column shell: left sticky TOC, right the active page. A minimal top
// bar brands the section and gives an easy way back to the app.
export default function DocsLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="min-h-screen bg-[var(--color-background)]">
      <header className="sticky top-0 z-20 h-14 border-b border-[var(--gray-a5)] bg-[var(--color-panel-solid)]/90 backdrop-blur px-6 flex items-center justify-between">
        <Link href="/" className="flex items-center gap-2 group">
          <span className="flex h-7 w-7 items-center justify-center rounded-md bg-[var(--accent-a5)] text-[var(--accent-11)] transition-colors group-hover:bg-[var(--accent-a6)]">
            <LightningBoltIcon width="16" height="16" />
          </span>
          <Heading size="3" weight="bold" className="tracking-tight">
            TestPay <span className="font-normal text-[var(--gray-11)]">docs</span>
          </Heading>
        </Link>
        <Flex gap="4" align="center" className="text-sm">
          <Link
            href="/login"
            className="text-[var(--gray-11)] hover:text-[var(--gray-12)] transition-colors"
          >
            Sign in
          </Link>
          <Link
            href="/signup"
            className="inline-flex items-center gap-1.5 rounded-md bg-[var(--accent-9)] px-3 py-1.5 text-[var(--accent-contrast)] hover:bg-[var(--accent-10)] transition-colors"
          >
            <PersonIcon width="14" height="14" />
            Get started
          </Link>
        </Flex>
      </header>
      <div className="flex gap-6 px-6 py-6">
        <DocsSidebar />
        <div className="flex-1 min-w-0 max-w-3xl pl-2 py-2">{children}</div>
      </div>
    </div>
  );
}
