import { Badge } from '@/components/ui/badge';
export function GatewayBadge({ gateway }: { gateway: string }) {
  return (
    <Badge variant="outline" className="font-mono text-xs">
      {gateway}
    </Badge>
  );
}
