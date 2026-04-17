'use client';
import { useState } from 'react';
import { EyeOpenIcon, EyeClosedIcon } from '@radix-ui/react-icons';
import { Button } from '@radix-ui/themes';
import { CopyButton } from './copy-button';

export function ApiKeyReveal({ value }: { value: string }) {
  const [shown, setShown] = useState(false);
  const masked = value.slice(0, 4) + '•'.repeat(16) + value.slice(-4);
  return (
    <div className="flex items-center gap-2">
      <code className="flex-1 font-mono text-sm bg-muted px-3 py-2 rounded-md">
        {shown ? value : masked}
      </code>
      <Button size="1" variant="ghost" color="gray" onClick={() => setShown((v) => !v)}>
        {shown ? <EyeClosedIcon /> : <EyeOpenIcon />}
      </Button>
      <CopyButton value={value} label="" />
    </div>
  );
}
