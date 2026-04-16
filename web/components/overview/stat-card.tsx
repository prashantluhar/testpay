import { Card, CardContent } from '@/components/ui/card';

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
      <CardContent className="pt-6">
        <div className="text-xs uppercase tracking-wider text-muted-foreground">{label}</div>
        <div className={`text-3xl font-semibold mt-2 ${color}`}>{value}</div>
      </CardContent>
    </Card>
  );
}
