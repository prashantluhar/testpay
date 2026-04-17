// Theme presets for the Radix <Theme> component.
// Each preset bundles a named combination of accentColor / grayColor / radius /
// panelBackground so users can pick a look from Settings → Appearance.
// The `indigo-slate` preset stays the default to preserve the existing palette.

export type AccentColor =
  | 'gray'
  | 'gold'
  | 'bronze'
  | 'brown'
  | 'yellow'
  | 'amber'
  | 'orange'
  | 'tomato'
  | 'red'
  | 'ruby'
  | 'crimson'
  | 'pink'
  | 'plum'
  | 'purple'
  | 'violet'
  | 'iris'
  | 'indigo'
  | 'blue'
  | 'cyan'
  | 'teal'
  | 'jade'
  | 'green'
  | 'grass'
  | 'lime'
  | 'mint'
  | 'sky';

export type GrayColor = 'slate' | 'mauve' | 'sage' | 'olive' | 'sand';
export type RadiusOption = 'none' | 'small' | 'medium' | 'large' | 'full';
export type PanelBackground = 'solid' | 'translucent';

export interface ThemePreset {
  id: string;
  name: string;
  description: string;
  accentColor: AccentColor;
  grayColor: GrayColor;
  radius: RadiusOption;
  panelBackground: PanelBackground;
}

export const DEFAULT_PRESET_ID = 'indigo-slate';

export const THEME_PRESETS: ThemePreset[] = [
  {
    id: 'indigo-slate',
    name: 'Indigo',
    description: 'Calm, balanced default',
    accentColor: 'indigo',
    grayColor: 'slate',
    radius: 'medium',
    panelBackground: 'solid',
  },
  {
    id: 'violet-mauve',
    name: 'Violet',
    description: 'Soft purple with warm gray',
    accentColor: 'violet',
    grayColor: 'mauve',
    radius: 'medium',
    panelBackground: 'solid',
  },
  {
    id: 'crimson-sand',
    name: 'Crimson',
    description: 'Bold red on warm sand',
    accentColor: 'crimson',
    grayColor: 'sand',
    radius: 'small',
    panelBackground: 'solid',
  },
  {
    id: 'grass-olive',
    name: 'Forest',
    description: 'Green on earthy olive',
    accentColor: 'grass',
    grayColor: 'olive',
    radius: 'medium',
    panelBackground: 'solid',
  },
  {
    id: 'cyan-sage',
    name: 'Ocean',
    description: 'Airy cyan with translucent panels',
    accentColor: 'cyan',
    grayColor: 'sage',
    radius: 'large',
    panelBackground: 'translucent',
  },
];

export function getPresetById(id: string | null | undefined): ThemePreset {
  if (!id) return THEME_PRESETS[0];
  return THEME_PRESETS.find((p) => p.id === id) ?? THEME_PRESETS[0];
}
