'use client';
import { CopyIcon, CheckIcon } from '@radix-ui/react-icons';
import { Button } from '@radix-ui/themes';
import { useState } from 'react';

export function CopyButton({ value, label = 'Copy' }: { value: string; label?: string }) {
  const [done, setDone] = useState(false);
  return (
    <Button
      variant="ghost"
      size="1"
      color="gray"
      onClick={() => {
        navigator.clipboard.writeText(value);
        setDone(true);
        setTimeout(() => setDone(false), 1500);
      }}
    >
      {done ? <CheckIcon /> : <CopyIcon />}
      {label ? <span>{done ? 'Copied' : label}</span> : null}
    </Button>
  );
}
