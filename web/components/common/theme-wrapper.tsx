'use client';
import { useEffect, useState } from 'react';
import { Theme } from '@radix-ui/themes';
import { useTheme } from './theme-provider';

// Bridges the existing `ThemeProvider` (which toggles `.dark` on <html>)
// into Radix Themes by passing the resolved appearance to <Theme>.
export function ThemeWrapper({ children }: { children: React.ReactNode }) {
  const { theme } = useTheme();
  const [resolved, setResolved] = useState<'light' | 'dark'>('dark');

  useEffect(() => {
    if (theme === 'system') {
      const mq = window.matchMedia('(prefers-color-scheme: dark)');
      const apply = () => setResolved(mq.matches ? 'dark' : 'light');
      apply();
      mq.addEventListener('change', apply);
      return () => mq.removeEventListener('change', apply);
    }
    setResolved(theme);
  }, [theme]);

  return (
    <Theme
      appearance={resolved}
      accentColor="indigo"
      grayColor="slate"
      radius="medium"
      panelBackground="solid"
    >
      {children}
    </Theme>
  );
}
