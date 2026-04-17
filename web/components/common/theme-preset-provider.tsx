'use client';
import { createContext, useContext, useEffect, useState } from 'react';
import {
  DEFAULT_PRESET_ID,
  THEME_PRESETS,
  getPresetById,
  type ThemePreset,
} from '@/lib/themes';

const STORAGE_KEY = 'testpay-theme-preset';

interface ThemePresetCtxValue {
  preset: ThemePreset;
  presetId: string;
  setPresetId: (id: string) => void;
}

const ThemePresetCtx = createContext<ThemePresetCtxValue>({
  preset: THEME_PRESETS[0],
  presetId: DEFAULT_PRESET_ID,
  setPresetId: () => {},
});

export function ThemePresetProvider({ children }: { children: React.ReactNode }) {
  // First render always uses the default to avoid SSR/client hydration mismatch.
  // Actual persisted value is loaded on mount via useEffect below.
  const [presetId, setPresetIdState] = useState<string>(DEFAULT_PRESET_ID);

  useEffect(() => {
    try {
      const saved = localStorage.getItem(STORAGE_KEY);
      if (saved && THEME_PRESETS.some((p) => p.id === saved)) {
        setPresetIdState(saved);
      }
    } catch {
      // localStorage unavailable; keep default.
    }
  }, []);

  function setPresetId(id: string) {
    setPresetIdState(id);
    try {
      localStorage.setItem(STORAGE_KEY, id);
    } catch {
      // ignore persistence failure
    }
  }

  const preset = getPresetById(presetId);

  return (
    <ThemePresetCtx.Provider value={{ preset, presetId, setPresetId }}>
      {children}
    </ThemePresetCtx.Provider>
  );
}

export const useThemePreset = () => useContext(ThemePresetCtx);
