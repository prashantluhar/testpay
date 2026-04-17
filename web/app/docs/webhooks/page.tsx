'use client';
import Link from 'next/link';
import { Heading, Text, Table, Card, Box } from '@radix-ui/themes';
import { CodeBlock } from '@/components/docs/code-block';

export default function WebhookSpecPage() {
  return (
    <div className="space-y-6">
      <div>
        <Heading size="7" mb="2">
          Webhook spec
        </Heading>
        <Text size="3" color="gray">
          TestPay dispatches a webhook after every successful (non-skipped) mock
          request, provided a target URL is configured. The payload shape is
          gateway-specific — see each adapter&apos;s page for concrete examples.
          This page covers the transport behavior common to all adapters.
        </Text>
      </div>

      <div>
        <Heading size="4" mb="2">
          Target URL resolution
        </Heading>
        <Text size="2" color="gray" className="block mb-2">
          For each outgoing webhook the dispatcher picks the first non-empty URL
          from this priority order:
        </Text>
        <Table.Root>
          <Table.Header>
            <Table.Row>
              <Table.ColumnHeaderCell>Priority</Table.ColumnHeaderCell>
              <Table.ColumnHeaderCell>Source</Table.ColumnHeaderCell>
              <Table.ColumnHeaderCell>Configured where</Table.ColumnHeaderCell>
            </Table.Row>
          </Table.Header>
          <Table.Body>
            <Table.Row>
              <Table.Cell>1</Table.Cell>
              <Table.Cell>
                <code>X-Webhook-URL</code> request header
              </Table.Cell>
              <Table.Cell>Per-request, on the mock call itself.</Table.Cell>
            </Table.Row>
            <Table.Row>
              <Table.Cell>2</Table.Cell>
              <Table.Cell>
                <code>workspace.webhook_urls[gateway]</code>
              </Table.Cell>
              <Table.Cell>
                <Link href="/settings" className="underline">
                  Settings
                </Link>{' '}
                → per-gateway override.
              </Table.Cell>
            </Table.Row>
            <Table.Row>
              <Table.Cell>3</Table.Cell>
              <Table.Cell>
                <code>workspace.webhook_urls._default</code>
              </Table.Cell>
              <Table.Cell>Settings → &quot;Default webhook URL&quot;.</Table.Cell>
            </Table.Row>
            <Table.Row>
              <Table.Cell>—</Table.Cell>
              <Table.Cell>(none)</Table.Cell>
              <Table.Cell>
                Webhook skipped; log shows{' '}
                <code>webhook_scheduled=false, reason=&quot;no target URL&quot;</code>.
              </Table.Cell>
            </Table.Row>
          </Table.Body>
        </Table.Root>
      </div>

      <div>
        <Heading size="4" mb="2">
          Retry policy
        </Heading>
        <Text size="2" color="gray" className="block mb-2">
          Default configuration (from <code>deploy/config</code>):
        </Text>
        <Card>
          <Box p="3">
            <ul className="list-disc pl-5 space-y-1 text-sm text-muted-foreground">
              <li>
                <code>max_attempts = 3</code> (configurable)
              </li>
              <li>
                <code>base_delay_ms = 500</code>
              </li>
              <li>
                Backoff: exponential — attempt N waits{' '}
                <code>base_delay_ms * 2^(N-1)</code> before firing. For defaults
                that means 0 ms → 500 ms → 1000 ms.
              </li>
              <li>
                Each attempt has a <code>10 s</code> HTTP client timeout.
              </li>
              <li>
                Any HTTP 2xx status stops the retry loop. Anything else — 3xx,
                4xx, 5xx, transport error — counts as a failed attempt.
              </li>
            </ul>
          </Box>
        </Card>
      </div>

      <div>
        <Heading size="4" mb="2">
          Request shape
        </Heading>
        <Text size="2" color="gray" className="block mb-2">
          Outbound webhooks are <code>POST</code>s with:
        </Text>
        <CodeBlock language="http">{`POST <target_url> HTTP/1.1
Content-Type: application/json
User-Agent: TestPay-Webhook/1.0

<gateway-specific payload>`}</CodeBlock>
        <Text size="2" color="gray" className="block mt-2">
          No signature header yet — the mock emits a placeholder signature field
          inside gateway payloads where the real gateway would sign
          (ESPay, Paynamics, TillPayment, ECPay, etc.). Consumers that verify
          signatures should detect the <code>mock-</code> prefix and skip
          verification when running against TestPay.
        </Text>
      </div>

      <div>
        <Heading size="4" mb="2">
          Attempt log (per delivery)
        </Heading>
        <Text size="2" color="gray" className="block mb-2">
          Every attempt — success or failure — is captured in{' '}
          <code>webhook_logs.attempt_logs</code>:
        </Text>
        <CodeBlock language="json">{`{
  "status":        200,
  "duration_ms":   34,
  "attempted_at":  "2026-04-17T12:34:56.789Z",
  "response":      "<response body, capped at 8 KB>",
  "error":         ""
}`}</CodeBlock>
        <ul className="list-disc pl-5 space-y-1 text-sm text-muted-foreground mt-2">
          <li>
            <code>status</code> — HTTP status from your endpoint. 0 on transport error.
          </li>
          <li>
            <code>duration_ms</code> — wall time for this single attempt.
          </li>
          <li>
            <code>response</code> — body returned by your endpoint, truncated at
            8 KB so a misconfigured endpoint can&apos;t blow the JSONB column.
          </li>
          <li>
            <code>error</code> — transport-level error string (timeout, DNS,
            TCP). Empty when any HTTP response was received.
          </li>
        </ul>
        <Text size="2" color="gray" className="block mt-2">
          Inspect attempts in the{' '}
          <Link href="/webhooks" className="underline">
            Webhooks
          </Link>{' '}
          tab — each row has a detail drawer with the full attempt history.
        </Text>
      </div>

      <div>
        <Heading size="4" mb="2">
          Interaction with webhook-family failure modes
        </Heading>
        <Table.Root>
          <Table.Header>
            <Table.Row>
              <Table.ColumnHeaderCell>Mode</Table.ColumnHeaderCell>
              <Table.ColumnHeaderCell>Dispatch behavior</Table.ColumnHeaderCell>
            </Table.Row>
          </Table.Header>
          <Table.Body>
            <Table.Row>
              <Table.Cell>
                <code>webhook_missing</code>
              </Table.Cell>
              <Table.Cell>
                Dispatch is skipped entirely. No webhook_log row is created.
              </Table.Cell>
            </Table.Row>
            <Table.Row>
              <Table.Cell>
                <code>webhook_delayed</code>
              </Table.Cell>
              <Table.Cell>
                Dispatch waits <code>scenario.webhook_delay_ms</code> before the
                first attempt.
              </Table.Cell>
            </Table.Row>
            <Table.Row>
              <Table.Cell>
                <code>webhook_duplicate</code>
              </Table.Cell>
              <Table.Cell>
                Two webhook_log rows created; the second fires 500 ms after the
                first. The second&apos;s <code>delivery_status</code> is{' '}
                <code>duplicate</code>.
              </Table.Cell>
            </Table.Row>
            <Table.Row>
              <Table.Cell>
                <code>double_charge</code>
              </Table.Cell>
              <Table.Cell>
                Same as <code>webhook_duplicate</code> — same underlying flag.
              </Table.Cell>
            </Table.Row>
          </Table.Body>
        </Table.Root>
      </div>

      <div>
        <Heading size="4" mb="2">
          Delivery statuses
        </Heading>
        <Table.Root>
          <Table.Header>
            <Table.Row>
              <Table.ColumnHeaderCell>Status</Table.ColumnHeaderCell>
              <Table.ColumnHeaderCell>Meaning</Table.ColumnHeaderCell>
            </Table.Row>
          </Table.Header>
          <Table.Body>
            <Table.Row>
              <Table.Cell>
                <code>pending</code>
              </Table.Cell>
              <Table.Cell>Created, not yet dispatched (or mid-retry).</Table.Cell>
            </Table.Row>
            <Table.Row>
              <Table.Cell>
                <code>delivered</code>
              </Table.Cell>
              <Table.Cell>
                One attempt returned an HTTP 2xx status. <code>delivered_at</code>{' '}
                is set.
              </Table.Cell>
            </Table.Row>
            <Table.Row>
              <Table.Cell>
                <code>failed</code>
              </Table.Cell>
              <Table.Cell>
                All <code>max_attempts</code> retries exhausted without a 2xx.
              </Table.Cell>
            </Table.Row>
            <Table.Row>
              <Table.Cell>
                <code>duplicate</code>
              </Table.Cell>
              <Table.Cell>
                The second dispatch emitted by <code>webhook_duplicate</code> /{' '}
                <code>double_charge</code>.
              </Table.Cell>
            </Table.Row>
          </Table.Body>
        </Table.Root>
      </div>
    </div>
  );
}
