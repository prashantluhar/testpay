'use client';
import { Heading, Text, Table } from '@radix-ui/themes';

interface ModeRow {
  wire: string;
  http: string | number;
  desc: string;
  useWhen: string;
}

const GROUPS: { title: string; modes: ModeRow[] }[] = [
  {
    title: 'Generic',
    modes: [
      {
        wire: 'success',
        http: 200,
        desc: 'Normal happy-path capture. Webhook fires as configured.',
        useWhen: 'Default baseline for every scenario.',
      },
    ],
  },
  {
    title: 'Bank-side',
    modes: [
      {
        wire: 'bank_decline_hard',
        http: 402,
        desc: 'Issuer declined. Terminal — retry will decline again.',
        useWhen: 'Verify your UX on the classic decline path.',
      },
      {
        wire: 'bank_decline_soft',
        http: 402,
        desc: 'Soft decline / insufficient funds. Same status, softer code.',
        useWhen: 'Prompt the customer to try another card.',
      },
      {
        wire: 'bank_server_down',
        http: 503,
        desc: 'Issuer unreachable. Upstream infra failure.',
        useWhen: 'Verify your retry / fallback-gateway logic.',
      },
      {
        wire: 'bank_timeout',
        http: 503,
        desc: 'Issuer accepted the request but did not respond in time.',
        useWhen: 'Test your reconciliation path for ambiguous-state payments.',
      },
      {
        wire: 'bank_invalid_cvv',
        http: 402,
        desc: 'CVV / security code mismatch.',
        useWhen: 'Prompt the customer to re-enter card details.',
      },
      {
        wire: 'bank_do_not_honour',
        http: 402,
        desc: 'Catch-all issuer refusal (no specific reason given).',
        useWhen: 'Your fallback decline branch.',
      },
    ],
  },
  {
    title: 'PG (gateway) server-side',
    modes: [
      {
        wire: 'pg_server_error',
        http: 500,
        desc: 'Gateway had an internal error. Transient.',
        useWhen: 'Verify retry-on-5xx behavior.',
      },
      {
        wire: 'pg_timeout',
        http: 503,
        desc: 'Gateway did not respond in time.',
        useWhen: 'Same as above — your client-side timeout handling.',
      },
      {
        wire: 'pg_rate_limited',
        http: 429,
        desc: 'Too many requests. Backoff-and-retry.',
        useWhen: 'Exercise your rate-limit handling / 429 backoff.',
      },
      {
        wire: 'pg_maintenance',
        http: 503,
        desc: 'Gateway scheduled maintenance. error_code = "maintenance".',
        useWhen: 'Verify scheduled-maintenance banner / fallback logic.',
      },
      {
        wire: 'network_error',
        http: 503,
        desc: 'Generic upstream network failure.',
        useWhen: 'Same bucket as bank/pg timeout.',
      },
    ],
  },
  {
    title: 'Webhook anomalies',
    modes: [
      {
        wire: 'webhook_missing',
        http: 200,
        desc:
          'HTTP response succeeds; webhook is never dispatched. Tests reconciliation when the webhook just never shows up.',
        useWhen: 'Verify your polling / out-of-band reconciliation path.',
      },
      {
        wire: 'webhook_delayed',
        http: 200,
        desc:
          'Webhook fires, but after the scenario-configured webhook_delay_ms delay.',
        useWhen: 'Verify your idempotency on late webhooks.',
      },
      {
        wire: 'webhook_duplicate',
        http: 200,
        desc:
          'Two webhooks dispatched for the same payment (with a 500 ms offset between them).',
        useWhen: 'Exercise your dedup / idempotency keys.',
      },
      {
        wire: 'webhook_out_of_order',
        http: 200,
        desc: 'Webhook payload ordering scrambled (success after reversal, etc.).',
        useWhen: 'Verify ordering tolerance in your consumer.',
      },
      {
        wire: 'webhook_malformed',
        http: 200,
        desc: 'Webhook body intentionally shaped wrong.',
        useWhen: 'Exercise your schema-validation error branches.',
      },
    ],
  },
  {
    title: 'Redirect / 3DS',
    modes: [
      {
        wire: 'redirect_success',
        http: 200,
        desc: 'Customer completed 3DS / hosted redirect successfully.',
        useWhen: 'Happy path for gateways with hosted payment pages.',
      },
      {
        wire: 'redirect_abandoned',
        http: 402,
        desc: 'Customer closed the 3DS tab without completing.',
        useWhen: 'Test your abandoned-cart follow-up.',
      },
      {
        wire: 'redirect_timeout',
        http: 402,
        desc: 'Hosted payment session timed out.',
        useWhen: 'Verify stale-session recovery.',
      },
      {
        wire: 'redirect_failed',
        http: 402,
        desc: 'Customer failed authentication (wrong OTP, etc.).',
        useWhen: 'Test 3DS failure flow.',
      },
    ],
  },
  {
    title: 'Charge anomalies',
    modes: [
      {
        wire: 'double_charge',
        http: 200,
        desc: 'Successful capture + duplicate webhook. Simulates a double-charge bug.',
        useWhen: 'Exercise your reconciliation / dispute handling.',
      },
      {
        wire: 'amount_mismatch',
        http: 200,
        desc:
          'Returns 200 but the amount on the response differs from what was requested.',
        useWhen: 'Verify you reject / alert on amount mismatches.',
      },
      {
        wire: 'partial_success',
        http: 200,
        desc: 'Split / partial capture scenario (e.g. multi-item order).',
        useWhen: 'Test partial-fulfilment flows.',
      },
    ],
  },
  {
    title: 'Async state transitions',
    modes: [
      {
        wire: 'pending_then_failed',
        http: 200,
        desc:
          'Initial response is "pending" (IsPending=true); subsequent webhook resolves to failed.',
        useWhen: 'Verify pending-UI + failed-final flow.',
      },
      {
        wire: 'pending_then_success',
        http: 200,
        desc: 'Pending response; subsequent webhook resolves to success.',
        useWhen: 'Verify pending-UI + success-final flow.',
      },
      {
        wire: 'failed_then_success',
        http: 200,
        desc: 'Initial response indicates failure; later webhook reverses to success.',
        useWhen: 'Rare but real — reversed-decline scenario.',
      },
      {
        wire: 'success_then_reversed',
        http: 200,
        desc:
          'Success response, then a reversal event (chargeback / cancel). Some adapters (Adyen, MPGS) emit distinct event codes.',
        useWhen: 'Test your dispute / refund downstream handling.',
      },
    ],
  },
];

export default function FailureModesPage() {
  return (
    <div className="space-y-6">
      <div>
        <Heading size="7" mb="2">
          Failure modes reference
        </Heading>
        <Text size="3" color="gray">
          The engine defines 28 canonical outcomes. Any scenario step&apos;s{' '}
          <code>outcome</code> field must be one of these wire names. The HTTP
          status shown is what the adapter returns; per-gateway error codes are
          documented on each adapter&apos;s page.
        </Text>
      </div>

      {GROUPS.map((g) => (
        <div key={g.title}>
          <Heading size="4" mb="2">
            {g.title}
          </Heading>
          <Table.Root>
            <Table.Header>
              <Table.Row>
                <Table.ColumnHeaderCell>Wire name (outcome)</Table.ColumnHeaderCell>
                <Table.ColumnHeaderCell>HTTP</Table.ColumnHeaderCell>
                <Table.ColumnHeaderCell>What it does</Table.ColumnHeaderCell>
                <Table.ColumnHeaderCell>When to use</Table.ColumnHeaderCell>
              </Table.Row>
            </Table.Header>
            <Table.Body>
              {g.modes.map((m) => (
                <Table.Row key={m.wire}>
                  <Table.Cell>
                    <code className="text-xs">{m.wire}</code>
                  </Table.Cell>
                  <Table.Cell>
                    <code>{m.http}</code>
                  </Table.Cell>
                  <Table.Cell>{m.desc}</Table.Cell>
                  <Table.Cell>{m.useWhen}</Table.Cell>
                </Table.Row>
              ))}
            </Table.Body>
          </Table.Root>
        </div>
      ))}

      <div>
        <Heading size="4" mb="2">
          Notes
        </Heading>
        <ul className="list-disc pl-5 space-y-1 text-sm text-muted-foreground">
          <li>
            For bank-decline modes the <code>error_code</code> defaults to the
            mode name (e.g. <code>bank_decline_hard</code>). Override via the{' '}
            <code>code</code> field on a step to send a gateway-native code.
          </li>
          <li>
            <code>pg_maintenance</code> is the only mode that hard-codes a
            specific <code>error_code</code> — the string <code>maintenance</code>.
          </li>
          <li>
            Webhook-family modes land as HTTP 200 at the API boundary; their
            effect is on the webhook dispatch side (skipped, duplicated,
            delayed, etc.).
          </li>
          <li>
            Async modes (<code>pending_*</code>, <code>failed_then_success</code>,{' '}
            <code>success_then_reversed</code>) set the internal{' '}
            <code>IsPending</code> flag — adapters map that onto their native
            pending / authorized / processing status strings.
          </li>
        </ul>
      </div>
    </div>
  );
}
