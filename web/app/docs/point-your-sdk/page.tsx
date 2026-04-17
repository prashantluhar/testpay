'use client';
import { Heading, Text, Tabs, Card, Box } from '@radix-ui/themes';
import { CodeBlock } from '@/components/docs/code-block';
import { getApiBaseUrl } from '@/components/docs/base-url';

export default function PointYourSdkPage() {
  const baseUrl = getApiBaseUrl();

  return (
    <div className="space-y-6">
      <div>
        <Heading size="7" mb="2">
          Point your SDK at the mock
        </Heading>
        <Text size="3" color="gray">
          TestPay speaks each gateway&apos;s native HTTP shape — you don&apos;t
          need a special SDK. Set the base URL on your existing SDK and every
          call goes through the mock.
        </Text>
      </div>

      <Card>
        <Box p="3">
          <Heading size="3" mb="2">
            The one rule
          </Heading>
          <Text size="2" color="gray">
            For hosted deployments, pass your workspace API key via{' '}
            <code>Authorization: Bearer &lt;api_key&gt;</code>. Find it in{' '}
            <code>Settings → API key</code>. Local mode ignores auth — every call
            is attributed to the <code>local</code> workspace.
          </Text>
        </Box>
      </Card>

      <Tabs.Root defaultValue="stripe">
        <Tabs.List>
          <Tabs.Trigger value="stripe">Stripe</Tabs.Trigger>
          <Tabs.Trigger value="razorpay">Razorpay</Tabs.Trigger>
          <Tabs.Trigger value="adyen">Adyen</Tabs.Trigger>
          <Tabs.Trigger value="generic">Generic HTTP</Tabs.Trigger>
        </Tabs.List>

        <Tabs.Content value="stripe">
          <div className="space-y-3 mt-4">
            <Heading size="3">Stripe SDK</Heading>
            <Text size="2" color="gray">
              Stripe&apos;s Node SDK accepts an <code>apiBase</code> / host
              override on the constructor.
            </Text>
            <CodeBlock language="javascript">{`import Stripe from 'stripe';

const stripe = new Stripe('sk_test_unused', {
  host: '${baseUrl.replace(/^https?:\/\//, '')}',
  protocol: '${baseUrl.startsWith('https') ? 'https' : 'http'}',
  port: ${baseUrl.startsWith('https') ? '443' : '7700'},
});

// Requests now hit ${baseUrl}/stripe/v1/...
const pi = await stripe.paymentIntents.create({
  amount: 5000,
  currency: 'usd',
  metadata: { order_id: 'ord_123' },
});`}</CodeBlock>

            <Text size="2" color="gray" className="block mt-4">
              Python — set <code>stripe.api_base</code>:
            </Text>
            <CodeBlock language="python">{`import stripe

stripe.api_base = "${baseUrl}/stripe"
stripe.api_key = "sk_test_unused"

charge = stripe.Charge.create(
    amount=5000,
    currency="usd",
    metadata={"order_id": "ord_123"},
)`}</CodeBlock>
          </div>
        </Tabs.Content>

        <Tabs.Content value="razorpay">
          <div className="space-y-3 mt-4">
            <Heading size="3">Razorpay SDK</Heading>
            <Text size="2" color="gray">
              Razorpay&apos;s SDKs don&apos;t expose a clean base-URL override.
              The pragmatic approach is an HTTP client pointed at the mock
              directly — the wire shape matches, so your existing webhook handler
              still works.
            </Text>
            <CodeBlock language="bash">{`export RAZORPAY_BASE=${baseUrl}/razorpay

curl -X POST "$RAZORPAY_BASE/v1/payments" \\
  -H "Authorization: Bearer <workspace_api_key>" \\
  -H "Content-Type: application/json" \\
  -d '{
    "amount": 10000,
    "currency": "INR",
    "notes": { "reference": "order_123" }
  }'`}</CodeBlock>

            <Text size="2" color="gray" className="block mt-4">
              For a full-SDK swap, fork the Razorpay client, override{' '}
              <code>BASE_URL</code>, and reinstall locally.
            </Text>
          </div>
        </Tabs.Content>

        <Tabs.Content value="adyen">
          <div className="space-y-3 mt-4">
            <Heading size="3">Adyen SDK</Heading>
            <Text size="2" color="gray">
              Adyen&apos;s Node SDK exposes an <code>environment</code> override
              plus a custom <code>liveEndpointUrlPrefix</code>.
            </Text>
            <CodeBlock language="javascript">{`import { Client, CheckoutAPI } from '@adyen/api-library';

const client = new Client({
  apiKey: 'api_key_unused',
  environment: 'LIVE',
  liveEndpointUrlPrefix: '${baseUrl.replace(/^https?:\/\//, '')}',
});

// Override the generated service endpoints:
client.config.endpoint = '${baseUrl}';
client.config.checkoutEndpoint = '${baseUrl}';

const checkout = new CheckoutAPI(client);
const res = await checkout.PaymentsApi.payments({
  merchantAccount: 'TEST',
  amount: { value: 5000, currency: 'EUR' },
  paymentMethod: { type: 'scheme' },
});`}</CodeBlock>

            <Text size="2" color="gray" className="block mt-4">
              If your SDK wraps the HTTP client too tightly to override, fall
              back to raw <code>fetch</code> against{' '}
              <code>{baseUrl}/adyen/v68/payments</code>.
            </Text>
          </div>
        </Tabs.Content>

        <Tabs.Content value="generic">
          <div className="space-y-3 mt-4">
            <Heading size="3">Any HTTP client</Heading>
            <Text size="2" color="gray">
              All adapters are vanilla JSON-over-HTTPS. Swap the base URL in
              whatever client you use (axios, got, requests, httpx, curl, etc.).
            </Text>
            <CodeBlock language="bash">{`curl -X POST ${baseUrl}/v1/charges \\
  -H "Authorization: Bearer <workspace_api_key>" \\
  -H "Content-Type: application/json" \\
  -d '{
    "amount": 5000,
    "currency": "usd",
    "order_id": "ord_123"
  }'`}</CodeBlock>

            <Text size="2" color="gray" className="block mt-4">
              Two TestPay-specific headers (both optional):
            </Text>
            <ul className="list-disc pl-5 space-y-1 text-sm text-muted-foreground">
              <li>
                <code>X-Webhook-URL</code> — override the workspace webhook
                target for this one call.
              </li>
              <li>
                <code>X-TestPay-Scenario-ID</code> — run the request against a
                specific scenario without needing a session.
              </li>
            </ul>
          </div>
        </Tabs.Content>
      </Tabs.Root>
    </div>
  );
}
