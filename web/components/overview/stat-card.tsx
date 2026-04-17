'use client';
import { Card } from '@radix-ui/themes';
import { useCountUp } from '@/lib/hooks';

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
  const isNumber = typeof value === 'number';
  return (
    <Card className="animate-in fade-in slide-in-from-bottom-2 duration-500">
      <div className="pt-2">
        <div className="text-xs uppercase tracking-wider text-muted-foreground">{label}</div>
        <div className={`text-3xl font-semibold mt-2 tabular-nums ${color}`}>
          {isNumber ? <CountUpValue target={value} /> : value}
        </div>
      </div>
    </Card>
  );
}

// Separate component so the count-up hook only runs for numeric values.
function CountUpValue({ target }: { target: number }) {
  const v = useCountUp(target);
  return <>{v}</>;
}
