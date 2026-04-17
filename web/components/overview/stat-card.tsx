import { Card } from '@radix-ui/themes';

export function StatCard({
  label,
  value,
  accent,
}: {
  label: string;
  value: string | number;
  accent?: 'good' | 'bad';
}) {
  const color = accent === 'good' ? 'text-emerald-500' : accent === 'bad' ? 'text-red-500' : '';
  return (
    <Card>
      <div className="pt-2">
        <div className="text-xs uppercase tracking-wider text-muted-foreground">{label}</div>
        <div className={`text-3xl font-semibold mt-2 ${color}`}>{value}</div>
      </div>
    </Card>
  );
}
