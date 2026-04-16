'use client';
import { useState } from 'react';
import { Eye, EyeOff } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { CopyButton } from './copy-button';

export function ApiKeyReveal({ value }: { value: string }) {
  const [shown, setShown] = useState(false);
  const masked = value.slice(0, 4) + '•'.repeat(16) + value.slice(-4);
  return (
    <div className="flex items-center gap-2">
      <code className="flex-1 font-mono text-sm bg-muted px-3 py-2 rounded-md">
        {shown ? value : masked}
      </code>
      <Button size="sm" variant="ghost" onClick={() => setShown((v) => !v)}>
        {shown ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
      </Button>
      <CopyButton value={value} label="" />
    </div>
  );
}
