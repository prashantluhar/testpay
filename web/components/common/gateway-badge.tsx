import { Badge } from '@radix-ui/themes';

export function GatewayBadge({ gateway }: { gateway: string }) {
  return (
    <Badge variant="outline" color="gray" className="font-mono">
      {gateway}
    </Badge>
  );
}
