'use client';
import { Table } from '@radix-ui/themes';

// Shimmer placeholder rows used while Table data is still loading.
// Matches the column count of the surrounding table via `columns`, and renders
// `rows` pulsing cells. Width variance in each column keeps it feeling organic.
export function TableSkeleton({
  rows = 6,
  columns = 5,
}: {
  rows?: number;
  columns?: number;
}) {
  // Fixed-per-column widths so rows line up into a column-like shape.
  const widths = ['w-24', 'w-16', 'w-14', 'w-64', 'w-20', 'w-16', 'w-12', 'w-32'];
  return (
    <Table.Body>
      {Array.from({ length: rows }).map((_, r) => (
        <Table.Row key={r} className="animate-in fade-in duration-300">
          {Array.from({ length: columns }).map((_, c) => (
            <Table.Cell key={c}>
              <div
                className={`h-3 rounded bg-muted animate-pulse ${widths[c % widths.length]} max-w-full`}
                style={{ animationDelay: `${(r * columns + c) * 40}ms` }}
              />
            </Table.Cell>
          ))}
        </Table.Row>
      ))}
    </Table.Body>
  );
}
