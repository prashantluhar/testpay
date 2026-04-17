'use client';
import Link from 'next/link';
import { Heading, Text, Card, Box, Table } from '@radix-ui/themes';
import { CodeBlock } from '@/components/docs/code-block';
import { getApiBaseUrl } from '@/components/docs/base-url';

export default function ScenariosDocsPage() {
  const baseUrl = getApiBaseUrl();

  return (
    <div className="space-y-6">
      <div>
        <Heading size="7" mb="2">
          Scenarios
        </Heading>
        <Text size="3" color="gray">
          A scenario is a named, ordered list of steps. Each step specifies one{' '}
          <code>event</code> (charge / refund / capture), one{' '}
          <code>outcome</code> (success, bank_decline_hard, pg_timeout, …), and
          an optional gateway-specific <code>code</code> override. When a scenario
          is active, successive mock requests walk the step list and the matching
          step shapes the response.
        </Text>
      </div>

      <div>
        <Heading size="4" mb="2">
          Shape of a step
        </Heading>
        <CodeBlock language="json">{`{
  "event": "charge",
  "outcome": "bank_decline_hard",
  "code": "insufficient_funds"
}`}</CodeBlock>
        <Text size="2" color="gray" className="block mt-2">
          <code>outcome</code> must be one of the 28 values listed on the{' '}
          <Link href="/docs/failure-modes" className="underline">
            Failure modes
          </Link>{' '}
          page. <code>code</code> is an optional override for the gateway-specific
          error code (e.g. Stripe <code>card_declined</code>, Razorpay{' '}
          <code>BAD_REQUEST_ERROR</code>) — when omitted, the adapter picks a
          sensible default for the chosen outcome.
        </Text>
      </div>

      <div>
        <Heading size="4" mb="2">
          Four ways to pick an outcome
        </Heading>
        <Text size="2" color="gray" className="block mb-3">
          TestPay resolves the response for each mock request in this priority
          order (highest wins):
        </Text>

        <Card className="mb-3">
          <Box p="3">
            <Heading size="3" mb="1">
              1. <code>X-TestPay-Outcome</code> header (inline, no scenario needed)
            </Heading>
            <Text size="2" color="gray" className="block mb-2">
              Fastest single-shot override. Pass the wire name of any failure
              mode directly — TestPay synthesizes a one-step scenario on the
              fly for this request. Ideal for CI smoke tests where creating a
              named scenario is overkill.
            </Text>
            <CodeBlock language="bash">{`curl -X POST ${baseUrl}/stripe/v1/charges \\
  -H "Authorization: Bearer <api_key>" \\
  -H "X-TestPay-Outcome: bank_do_not_honour" \\
  -H "Content-Type: application/json" \\
  -d '{"amount":5000,"currency":"usd"}'
# → HTTP 402 with the Stripe-shaped error envelope`}</CodeBlock>
            <Text size="1" color="gray" mt="2" as="p">
              Unknown mode values fall through to the next source and are logged
              server-side as a warning — typos don&apos;t silently return success.
              See <Link href="/docs/failure-modes" className="underline">failure modes reference</Link> for the full list of wire names.
            </Text>
          </Box>
        </Card>

        <Card className="mb-3">
          <Box p="3">
            <Heading size="3" mb="1">
              2. <code>X-TestPay-Scenario-ID</code> header
            </Heading>
            <Text size="2" color="gray" className="block mb-2">
              Per-request override that picks an existing named scenario.
              Useful when you&apos;ve built a multi-step scenario and want to
              fire its first step.
            </Text>
            <CodeBlock language="bash">{`curl -X POST ${baseUrl}/stripe/v1/charges \\
  -H "Authorization: Bearer <api_key>" \\
  -H "X-TestPay-Scenario-ID: <scenario_uuid>" \\
  -H "Content-Type: application/json" \\
  -d '{"amount":5000,"currency":"usd"}'`}</CodeBlock>
          </Box>
        </Card>

        <Card className="mb-3">
          <Box p="3">
            <Heading size="3" mb="1">
              3. Active session
            </Heading>
            <Text size="2" color="gray" className="block mb-2">
              Create a session via <code>POST /api/sessions</code>. While the
              session is alive for your workspace, every mock request walks the
              linked scenario&apos;s step list — first call uses step 0, next
              uses step 1, and so on.
            </Text>
            <CodeBlock language="bash">{`curl -X POST ${baseUrl}/api/sessions \\
  -H "Cookie: testpay_session=..." \\
  -H "Content-Type: application/json" \\
  -d '{"scenario_id": "<scenario_uuid>", "ttl_seconds": 3600}'`}</CodeBlock>
          </Box>
        </Card>

        <Card>
          <Box p="3">
            <Heading size="3" mb="1">
              4. Workspace default
            </Heading>
            <Text size="2" color="gray">
              Mark one scenario <code>is_default: true</code> — it&apos;s used
              for every request that doesn&apos;t specify a header or have an
              active session. Set via the{' '}
              <Link href="/scenarios" className="underline">
                Scenarios
              </Link>{' '}
              tab. If none matches, TestPay falls back to a built-in
              always-succeed scenario.
            </Text>
          </Box>
        </Card>
      </div>

      <div>
        <Heading size="4" mb="2">
          How steps advance
        </Heading>
        <Text size="2" color="gray" className="block mb-2">
          Step advancement only happens under path 2 (active session). Each mock
          request atomically increments the session&apos;s <code>call_index</code>{' '}
          counter, and the step used is{' '}
          <code>pre_bump_index % len(steps)</code>.
        </Text>
        <Text size="2" color="gray" className="block mb-2">
          Paths 1 (header) and 3 (workspace default) always use step <code>0</code>.
        </Text>

        <Heading size="3" mt="4" mb="2">
          Worked example
        </Heading>
        <Text size="2" color="gray" className="block mb-2">
          Session points at a scenario with two steps:
        </Text>
        <CodeBlock language="json">{`[
  { "event": "charge", "outcome": "bank_decline_hard" },
  { "event": "charge", "outcome": "success" }
]`}</CodeBlock>
        <Table.Root className="mt-3">
          <Table.Header>
            <Table.Row>
              <Table.ColumnHeaderCell>Call</Table.ColumnHeaderCell>
              <Table.ColumnHeaderCell>Pre-bump call_index</Table.ColumnHeaderCell>
              <Table.ColumnHeaderCell>Step</Table.ColumnHeaderCell>
              <Table.ColumnHeaderCell>Result</Table.ColumnHeaderCell>
            </Table.Row>
          </Table.Header>
          <Table.Body>
            <Table.Row>
              <Table.Cell>1</Table.Cell>
              <Table.Cell>0</Table.Cell>
              <Table.Cell>0</Table.Cell>
              <Table.Cell>HTTP 402 — bank_decline_hard</Table.Cell>
            </Table.Row>
            <Table.Row>
              <Table.Cell>2</Table.Cell>
              <Table.Cell>1</Table.Cell>
              <Table.Cell>1</Table.Cell>
              <Table.Cell>HTTP 200 — success</Table.Cell>
            </Table.Row>
            <Table.Row>
              <Table.Cell>3</Table.Cell>
              <Table.Cell>2</Table.Cell>
              <Table.Cell>0 (wraps)</Table.Cell>
              <Table.Cell>HTTP 402 — bank_decline_hard</Table.Cell>
            </Table.Row>
          </Table.Body>
        </Table.Root>
      </div>

      <div>
        <Heading size="4" mb="2">
          Running a scenario manually
        </Heading>
        <Text size="2" color="gray" className="block mb-2">
          <code>POST /api/scenarios/&#123;id&#125;/run</code> iterates every step
          in the scenario at once and records a <code>scenario_run</code> entry.
          Useful for CI replay / &quot;run all steps once&quot; smoke tests. This
          path does <strong>not</strong> touch the per-session <code>call_index</code>;
          use an active session when you want progressive stepping driven by
          your real SDK calls.
        </Text>
        <CodeBlock language="bash">{`curl -X POST ${baseUrl}/api/scenarios/<id>/run \\
  -H "Cookie: testpay_session=..."`}</CodeBlock>
      </div>

      <div>
        <Heading size="4" mb="2">
          Creating a scenario
        </Heading>
        <Text size="2" color="gray" className="block mb-2">
          Two options:
        </Text>
        <ul className="list-disc pl-5 space-y-1 text-sm text-muted-foreground">
          <li>
            UI —{' '}
            <Link href="/scenarios/new" className="underline">
              New scenario
            </Link>{' '}
            form with a step builder.
          </li>
          <li>
            API —{' '}
            <code>POST /api/scenarios</code> (see{' '}
            <Link href="/docs/api" className="underline">
              API reference
            </Link>
            ).
          </li>
        </ul>
      </div>
    </div>
  );
}
