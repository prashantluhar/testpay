'use client';
// Right-anchored drawer ("Sheet") built on the Radix Dialog primitive.
// Radix Themes ships a centered Dialog but not a side-anchored one, so this
// fills that gap with the same a11y guarantees (focus trap, escape-to-close,
// scroll lock, aria-labelling) + a slide-in-from-right animation.
//
// Note: the Portal renders content outside the root <Theme>, so Radix Themes
// components inside (Tabs, Badge, Button, etc.) would otherwise lose their
// CSS tokens and render unstyled. We wrap the portalled content in
// ThemeWrapper so the active appearance + preset propagate correctly.
import * as DialogPrimitive from '@radix-ui/react-dialog';
import { Cross2Icon } from '@radix-ui/react-icons';
import type { ComponentPropsWithoutRef, ReactNode } from 'react';
import { ThemeWrapper } from './theme-wrapper';

export const Sheet = DialogPrimitive.Root;
export const SheetTrigger = DialogPrimitive.Trigger;
export const SheetClose = DialogPrimitive.Close;
export const SheetTitle = DialogPrimitive.Title;
export const SheetDescription = DialogPrimitive.Description;

type SheetContentProps = {
  children: ReactNode;
  width?: number;
  className?: string;
} & Omit<ComponentPropsWithoutRef<typeof DialogPrimitive.Content>, 'children'>;

export function SheetContent({
  children,
  width = 640,
  className = '',
  style,
  ...props
}: SheetContentProps) {
  return (
    <DialogPrimitive.Portal>
      <ThemeWrapper>
        <DialogPrimitive.Overlay className="fixed inset-0 z-40 bg-black/60 backdrop-blur-sm data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=open]:fade-in-0 data-[state=closed]:fade-out-0 duration-200" />
        <DialogPrimitive.Content
          {...props}
          style={{ maxWidth: width, ...style }}
          className={`fixed right-0 top-0 z-50 h-full w-full overflow-y-auto border-l border-[var(--gray-a5)] bg-[var(--color-panel-solid)] p-6 shadow-2xl outline-none data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=open]:slide-in-from-right data-[state=closed]:slide-out-to-right duration-300 ease-out ${className}`}
        >
          <DialogPrimitive.Close
            aria-label="Close"
            className="absolute right-4 top-4 rounded-md p-1.5 text-[var(--gray-11)] transition-colors hover:bg-[var(--gray-a4)] hover:text-[var(--gray-12)] focus:outline-none focus-visible:ring-2 focus-visible:ring-[var(--accent-8)]"
          >
            <Cross2Icon width="18" height="18" />
          </DialogPrimitive.Close>
          {children}
        </DialogPrimitive.Content>
      </ThemeWrapper>
    </DialogPrimitive.Portal>
  );
}
