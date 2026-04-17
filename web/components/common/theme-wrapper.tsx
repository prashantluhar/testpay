'use client';
import { useEffect, useState } from 'react';
import { Theme } from '@radix-ui/themes';
import { useTheme } from './theme-provider';
import { useThemePreset } from './theme-preset-provider';

// Bridges the existing `ThemeProvider` (which toggles `.dark` on <html>)
// and the `ThemePresetProvider` (which holds the selected color preset)
// into Radix Themes by passing the resolved props to <Theme>.
export function ThemeWrapper({ children }: { children: React.ReactNode }) {
  const { theme } = useTheme();
  const { preset } = useThemePreset();
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
      accentColor={preset.accentColor}
      grayColor={preset.grayColor}
      radius={preset.radius}
      panelBackground={preset.panelBackground}
    >
      {children}
    </Theme>
  );
}
