'use client';
import Link from 'next/link';
import { Heading, Text, Card, Box } from '@radix-ui/themes';
import { CodeBlock } from '@/components/docs/code-block';
import { CopyableUrl } from '@/components/docs/copyable-url';
import { getApiBaseUrl } from '@/components/docs/base-url';

const GATEWAYS = [
  { slug: 'stripe', label: 'Stripe' },
  { slug: 'razorpay', label: 'Razorpay' },
  { slug: 'adyen', label: 'Adyen' },
  { slug: 'mastercard', label: 'Mastercard (MPGS)' },
  { slug: 'tillpay', label: 'TillPayment' },
  { slug: 'ecpay', label: 'ECPay' },
  { slug: 'espay', label: 'ESPay' },
  { slug: 'instamojo', label: 'Instamojo' },
  { slug: 'komoju', label: 'Komoju' },
  { slug: 'paynamics', label: 'Paynamics' },
  { slug: 'tappay', label: 'TapPay' },
  { slug: 'agnostic', label: 'Agnostic' },
];

export default function GettingStartedPage() {
  const baseUrl = getApiBaseUrl();

  return (
    <div className="space-y-6">
      <div>
        <Heading size="7" mb="2">
          Getting started
        </Heading>
        <Text size="3" color="gray">
          TestPay is a mock payment gateway that impersonates 10+ real PSPs on the
          wire. Point any Stripe/Razorpay/Adyen/etc. SDK at the base URL below,
          and TestPay returns responses and webhooks that match the real
          gateway&apos;s shape — except you control the outcome via scenarios.
        </Text>
      </div>

      <div>
        <Heading size="4" mb="2">
          1. Grab the base URL
        </Heading>
        <Text size="2" color="gray">
          Every mock lives under this URL. Append the gateway name (e.g.{' '}
          <code>/stripe</code>) to hit the matching adapter.
        </Text>
        <CopyableUrl url={baseUrl} label="base" />
      </div>

      <div>
        <Heading size="4" mb="2">
          2. Prove it works
        </Heading>
        <Text size="2" color="gray" className="block mb-2">
          The agnostic endpoint at <code>/v1/*</code> is the fastest smoke test —
          no per-gateway quirks, just a JSON echo.
        </Text>
        <CodeBlock language="bash">{`curl -X POST ${baseUrl}/v1/charges \\
  -H "Content-Type: application/json" \\
  -d '{"amount": 5000, "currency": "usd", "order_id": "ord_123"}'`}</CodeBlock>
        <Text size="2" color="gray" className="block mt-2">
          You should see a <code>200</code> with{' '}
          <code>{`{ "id": "txn_...", "status": "success", "error_code": "" }`}</code>
          . The request will appear in the{' '}
          <Link href="/logs" className="underline">
            Logs
          </Link>{' '}
          tab within a second.
        </Text>
      </div>

      <div>
        <Heading size="4" mb="2">
          3. Pick your gateway
        </Heading>
        <Text size="2" color="gray" className="block mb-3">
          Each gateway has its own adapter page with the exact wire shapes
          (response bodies, webhook payloads, accepted request fields).
        </Text>
        <div className="grid grid-cols-2 md:grid-cols-3 gap-2">
          {GATEWAYS.map((g) => (
            <Link
              key={g.slug}
              href={`/docs/adapters/${g.slug}`}
              className="border rounded-md px-3 py-2 text-sm hover:bg-accent/50 transition-colors"
            >
              <div className="font-medium">{g.label}</div>
              <div className="text-xs text-muted-foreground font-mono mt-0.5">
                /{g.slug === 'agnostic' ? 'v1' : g.slug}
              </div>
            </Link>
          ))}
        </div>
      </div>

      <div>
        <Heading size="4" mb="2">
          4. Shape the outcomes
        </Heading>
        <Text size="2" color="gray" className="block mb-2">
          By default every call returns success. To simulate bank declines,
          timeouts, rate limits, or any of the{' '}
          <Link href="/docs/failure-modes" className="underline">
            28 failure modes
          </Link>
          , create a scenario and activate it. Three ways to attach one:
        </Text>
        <Card>
          <Box p="3">
            <ul className="list-disc pl-5 space-y-1.5 text-sm text-muted-foreground">
              <li>
                <code>X-TestPay-Scenario-ID</code> header on the request
                (per-request)
              </li>
              <li>
                Active session — see{' '}
                <Link href="/docs/scenarios" className="underline">
                  Scenarios
                </Link>
              </li>
              <li>
                Workspace default — set from the{' '}
                <Link href="/scenarios" className="underline">
                  Scenarios
                </Link>{' '}
                tab
              </li>
            </ul>
          </Box>
        </Card>
      </div>

      <div>
        <Heading size="4" mb="2">
          5. Hosted mode: get an API key
        </Heading>
        <Text size="2" color="gray" className="block mb-2">
          If you&apos;re pointing at the hosted demo, every mock request must
          carry an <code>Authorization: Bearer &lt;api_key&gt;</code> header. Your
          key lives in the{' '}
          <Link href="/settings" className="underline">
            Settings
          </Link>{' '}
          tab — click &quot;Reveal&quot; to copy it.
        </Text>
        <Text size="2" color="gray">
          Local mode skips this — any call is attributed to the <code>local</code>{' '}
          workspace.
        </Text>
      </div>
    </div>
  );
}
