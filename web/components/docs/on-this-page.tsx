'use client';
import { useEffect, useState } from 'react';

type TocItem = { id: string; text: string; depth: number };

// Scans the docs page main content for h1/h2 headings after mount, slugifies
// their text to build anchor IDs, and renders a scroll-spy nav. Each time the
// pathname changes we re-scan so navigating between docs rebuilds the list.
//
// Works with Radix Themes <Heading> even though it defaults to <h1> because
// we walk tag names, not component identity. We skip the page's FIRST h1
// (the page title) to avoid a redundant top entry.
export function OnThisPage({ containerSelector = 'main' }: { containerSelector?: string }) {
  const [items, setItems] = useState<TocItem[]>([]);
  const [activeId, setActiveId] = useState<string | null>(null);

  useEffect(() => {
    const root = document.querySelector(containerSelector);
    if (!root) return;
    const nodes = Array.from(root.querySelectorAll('h1, h2, h3')) as HTMLElement[];
    if (nodes.length === 0) return;

    // Drop the first heading — it's the page title and would be a
    // no-op scroll target from its own TOC.
    const body = nodes.slice(1);

    const toc: TocItem[] = body.map((el) => {
      if (!el.id) el.id = slugify(el.textContent || '');
      return { id: el.id, text: el.textContent || '', depth: parseInt(el.tagName[1], 10) };
    });
    setItems(toc);

    // Scroll-spy via IntersectionObserver.
    const obs = new IntersectionObserver(
      (entries) => {
        const visible = entries.filter((e) => e.isIntersecting);
        if (visible.length > 0) {
          visible.sort((a, b) => (a.target as HTMLElement).offsetTop - (b.target as HTMLElement).offsetTop);
          setActiveId(visible[0].target.id);
        }
      },
      { rootMargin: '0px 0px -70% 0px', threshold: 0.1 },
    );
    body.forEach((el) => obs.observe(el));
    return () => obs.disconnect();
  }, [containerSelector]);

  if (items.length === 0) return null;

  return (
    <nav aria-labelledby="on-this-page-heading" className="text-sm">
      <div
        id="on-this-page-heading"
        className="text-[10px] uppercase tracking-wider text-[var(--gray-11)] mb-2 px-1"
      >
        On this page
      </div>
      <ul className="space-y-1">
        {items.map((it) => (
          <li key={it.id} style={{ paddingLeft: (it.depth - 1) * 8 }}>
            <a
              href={`#${it.id}`}
              className={`block py-1 px-2 rounded transition-colors border-l-2 ${
                activeId === it.id
                  ? 'border-[var(--accent-9)] text-[var(--accent-11)] bg-[var(--accent-a3)]'
                  : 'border-transparent text-[var(--gray-11)] hover:text-[var(--gray-12)] hover:border-[var(--gray-a6)]'
              }`}
            >
              {it.text}
            </a>
          </li>
        ))}
      </ul>
    </nav>
  );
}

function slugify(s: string) {
  return s
    .toLowerCase()
    .trim()
    .replace(/[^\w\s-]/g, '')
    .replace(/\s+/g, '-')
    .slice(0, 64);
}
