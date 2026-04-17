import type { Metadata } from 'next';
import Script from 'next/script';
import { Inter, JetBrains_Mono } from 'next/font/google';
import '@radix-ui/themes/styles.css';
import './globals.css';
import { ThemeProvider } from '@/components/common/theme-provider';
import { ThemePresetProvider } from '@/components/common/theme-preset-provider';
import { ThemeWrapper } from '@/components/common/theme-wrapper';
import { Toaster } from 'sonner';

const inter = Inter({ subsets: ['latin'], variable: '--font-sans' });
const mono = JetBrains_Mono({ subsets: ['latin'], variable: '--font-mono' });

export const metadata: Metadata = {
  title: 'TestPay',
  description: 'Mock payment gateway for local development and CI',
};

// Analytics config — env-driven so production can pick exactly one
// (or neither). Setting NEXT_PUBLIC_PLAUSIBLE_DOMAIN enables Plausible;
// NEXT_PUBLIC_GOATCOUNTER_URL enables GoatCounter. When both are set,
// both load — you probably don't want that, but it's cheap enough to
// support if you're migrating between them.
const plausibleDomain = process.env.NEXT_PUBLIC_PLAUSIBLE_DOMAIN;
const goatCounterUrl = process.env.NEXT_PUBLIC_GOATCOUNTER_URL;

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en" className={`${inter.variable} ${mono.variable}`} suppressHydrationWarning>
      <body className="bg-background text-foreground antialiased">
        <ThemeProvider>
          <ThemePresetProvider>
            <ThemeWrapper>{children}</ThemeWrapper>
          </ThemePresetProvider>
        </ThemeProvider>
        <Toaster richColors position="top-right" />
        {plausibleDomain && (
          <Script
            strategy="afterInteractive"
            data-domain={plausibleDomain}
            src="https://plausible.io/js/script.js"
          />
        )}
        {goatCounterUrl && (
          <Script
            strategy="afterInteractive"
            data-goatcounter={goatCounterUrl}
            src="//gc.zgo.at/count.js"
          />
        )}
      </body>
    </html>
  );
}
