import { Box, Flex, Heading, Text, Badge } from '@radix-ui/themes';
import { LightningBoltIcon } from '@radix-ui/react-icons';
import { AuthFormTransition } from '@/components/auth/auth-form-transition';

// Split-screen auth layout.
// - Below md: form takes the full screen.
// - md+: form left, product showcase right.
export default function AuthLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="min-h-screen w-full grid md:grid-cols-2">
      {/* Left: form column — child slides in/out directionally on navigation. */}
      <div className="flex items-center justify-center p-6 md:p-10">
        <AuthFormTransition>{children}</AuthFormTransition>
      </div>

      {/* Right: product showcase (hidden below md) */}
      <aside className="hidden md:flex relative flex-col justify-between overflow-hidden p-10 bg-gradient-to-br from-indigo-950 via-indigo-900 to-slate-900 text-slate-100">
        {/* decorative glow */}
        <div
          aria-hidden
          className="pointer-events-none absolute -top-24 -right-24 h-72 w-72 rounded-full bg-indigo-500/30 blur-3xl"
        />
        <div
          aria-hidden
          className="pointer-events-none absolute bottom-0 left-0 h-72 w-72 rounded-full bg-violet-500/20 blur-3xl"
        />

        {/* Top: brand */}
        <Flex
          align="center"
          gap="2"
          className="relative animate-in fade-in slide-in-from-left-4 duration-500"
        >
          <span className="flex h-9 w-9 items-center justify-center rounded-lg bg-indigo-500/20 text-indigo-300">
            <LightningBoltIcon width="20" height="20" />
          </span>
          <Heading size="6" weight="bold" style={{ color: 'inherit' }}>
            TestPay
          </Heading>
        </Flex>

        {/* Middle: tagline + numbers + pills */}
        <div className="relative flex flex-col gap-7 max-w-md">
          <div>
            <Heading
              size="8"
              weight="bold"
              className="animate-in fade-in slide-in-from-left-6 duration-700"
              style={{ color: 'inherit' }}
            >
              Postman for Payments
            </Heading>
            <Text
              size="3"
              className="mt-3 block text-slate-300 animate-in fade-in slide-in-from-left-6 duration-700"
              style={{ animationDelay: '120ms', animationFillMode: 'backwards' }}
            >
              A mock payment gateway that lets you test every real-world failure mode — locally and
              in CI — without touching production.
            </Text>
          </div>

          {/* By-the-numbers: staggered slide-in */}
          <div className="grid grid-cols-3 gap-3">
            {[
              { n: '28', l: 'failure modes' },
              { n: '13', l: 'gateways' },
              { n: '1', l: 'binary' },
            ].map((s, i) => (
              <div
                key={s.l}
                className="rounded-lg border border-white/10 bg-white/5 px-3 py-3 backdrop-blur-sm animate-in fade-in slide-in-from-bottom-3 duration-500"
                style={{ animationDelay: `${240 + i * 120}ms`, animationFillMode: 'backwards' }}
              >
                <div className="text-2xl font-semibold tabular-nums">{s.n}</div>
                <div className="text-[11px] uppercase tracking-wider text-slate-400">{s.l}</div>
              </div>
            ))}
          </div>

          {/* Feature pills */}
          <Flex wrap="wrap" gap="2">
            {[
              'Stripe & Razorpay & Adyen drop-in',
              'Webhook retry & replay',
              'Session-scoped scenarios',
              'Log every request + response',
              'Deployable as single binary',
              'Hosted demo on Render',
            ].map((label, i) => (
              <span
                key={label}
                className="animate-in fade-in duration-500"
                style={{ animationDelay: `${640 + i * 90}ms`, animationFillMode: 'backwards' }}
              >
                <Badge
                  color="indigo"
                  variant="soft"
                  radius="full"
                  style={{ background: 'rgba(255,255,255,0.06)', color: '#c7d2fe' }}
                >
                  {label}
                </Badge>
              </span>
            ))}
          </Flex>
        </div>

        {/* Bottom: flow diagram */}
        <Box
          className="relative animate-in fade-in duration-700"
          style={{ animationDelay: '1200ms', animationFillMode: 'backwards' }}
        >
          <FlowDiagram />
          <Text size="1" className="mt-3 block text-slate-500">
            Drop TestPay in front of your integration — keep webhooks, logs, and the dashboard for
            free.
          </Text>
        </Box>
      </aside>
    </div>
  );
}

function FlowDiagram() {
  return (
    <svg
      viewBox="0 0 420 90"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      className="w-full max-w-md text-slate-300"
      aria-hidden
    >
      {/* SDK */}
      <rect x="4" y="32" width="90" height="28" rx="6" stroke="currentColor" strokeWidth="1" />
      <text x="49" y="50" textAnchor="middle" fontSize="11" fill="currentColor" fontFamily="monospace">
        Your SDK
      </text>

      {/* arrow */}
      <path d="M98 46 L132 46" stroke="currentColor" strokeWidth="1" />
      <path d="M128 42 L132 46 L128 50" stroke="currentColor" strokeWidth="1" fill="none" />

      {/* TestPay */}
      <rect
        x="134"
        y="26"
        width="100"
        height="40"
        rx="6"
        stroke="currentColor"
        strokeWidth="1.5"
        fill="rgba(99,102,241,0.18)"
      />
      <text x="184" y="50" textAnchor="middle" fontSize="12" fill="#c7d2fe" fontFamily="monospace">
        TestPay
      </text>

      {/* branching arrows */}
      <path d="M238 38 L272 20" stroke="currentColor" strokeWidth="1" />
      <path d="M238 46 L272 46" stroke="currentColor" strokeWidth="1" />
      <path d="M238 54 L272 72" stroke="currentColor" strokeWidth="1" />

      {/* outputs */}
      <rect x="272" y="8" width="140" height="22" rx="4" stroke="currentColor" strokeWidth="1" />
      <text x="342" y="23" textAnchor="middle" fontSize="10" fill="currentColor" fontFamily="monospace">
        webhook
      </text>

      <rect x="272" y="34" width="140" height="22" rx="4" stroke="currentColor" strokeWidth="1" />
      <text x="342" y="49" textAnchor="middle" fontSize="10" fill="currentColor" fontFamily="monospace">
        logs
      </text>

      <rect x="272" y="60" width="140" height="22" rx="4" stroke="currentColor" strokeWidth="1" />
      <text x="342" y="75" textAnchor="middle" fontSize="10" fill="currentColor" fontFamily="monospace">
        dashboard
      </text>
    </svg>
  );
}
