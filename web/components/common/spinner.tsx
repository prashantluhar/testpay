'use client';
// Thin re-export of Radix Themes' <Spinner> with a friendlier size API.
// Keeps one import site so we can swap implementations later if needed.
import { Spinner as RadixSpinner } from '@radix-ui/themes';
import type { ComponentProps } from 'react';

type RadixSize = ComponentProps<typeof RadixSpinner>['size'];
type Size = 'small' | 'medium' | 'large';

const SIZE_MAP: Record<Size, RadixSize> = {
  small: '1',
  medium: '2',
  large: '3',
};

export interface SpinnerProps
  extends Omit<ComponentProps<typeof RadixSpinner>, 'size'> {
  size?: Size;
}

export function Spinner({ size = 'medium', ...rest }: SpinnerProps) {
  return <RadixSpinner size={SIZE_MAP[size]} {...rest} />;
}
