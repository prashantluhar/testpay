'use client';
import { useRef, type ReactNode } from 'react';
import { usePathname } from 'next/navigation';

// Ordered tabs so we can compute slide direction:
//   /login  → index 0  (left)
//   /signup → index 1  (right)
// Navigating from login → signup slides IN from the right (new tab is to the
// right). Going back slides IN from the left.
const TAB_ORDER: Record<string, number> = {
  '/login': 0,
  '/signup': 1,
};

export function AuthFormTransition({ children }: { children: ReactNode }) {
  const pathname = usePathname();
  const prevRef = useRef(pathname);

  // Compute direction using the PREVIOUS pathname (ref value is still stale
  // during this render — the effect below updates it after paint).
  const prevIdx = TAB_ORDER[prevRef.current] ?? 0;
  const nextIdx = TAB_ORDER[pathname] ?? 0;
  const slideFrom = nextIdx >= prevIdx ? 'slide-in-from-right-4' : 'slide-in-from-left-4';

  // Schedule ref update for after render — using a microtask keeps it cheap
  // and avoids an extra useEffect call.
  if (prevRef.current !== pathname) {
    queueMicrotask(() => {
      prevRef.current = pathname;
    });
  }

  return (
    <div key={pathname} className={`w-full max-w-sm animate-in fade-in duration-300 ease-out ${slideFrom}`}>
      {children}
    </div>
  );
}
