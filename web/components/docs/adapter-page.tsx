'use client';
import { Heading, Text, Tabs, Table, Card, Box } from '@radix-ui/themes';
import { CodeBlock } from '@/components/docs/code-block';
import { CopyableUrl } from '@/components/docs/copyable-url';

export interface Operation {
  label: string;
  method: string;
  path: string;
  note?: string;
}

export interface EchoedField {
  field: string;
  description: string;
}

export interface ErrorExample {
  label: string; // e.g. "bank_decline_hard"
  status: number;
  body: string; // JSON string
}

export interface AdapterPageProps {
  title: string;
  slug: string; // "stripe" — used for the base URL
  blurb: string;
  baseUrl: string; // full URL, e.g. https://.../stripe
  operations: Operation[];
  opsNote?: string;
  requestFields: EchoedField[];
  successBody: string; // JSON
  successNotes?: string;
  errorExamples: ErrorExample[];
  webhookBody: string; // JSON
  webhookNotes?: string;
  echoedFields: EchoedField[];
  curlExample: string;
}

export function AdapterPage(props: AdapterPageProps) {
  return (
    <div className="space-y-6">
      <div>
        <Heading size="7" mb="2">
          {props.title}
        </Heading>
        <Text size="3" color="gray">
          {props.blurb}
        </Text>
      </div>

      <div>
        <Heading size="4" mb="2">
          Base URL
        </Heading>
        <CopyableUrl url={props.baseUrl} />
      </div>

      <Tabs.Root defaultValue="ops">
        <Tabs.List>
          <Tabs.Trigger value="ops">Operations</Tabs.Trigger>
          <Tabs.Trigger value="request">Request</Tabs.Trigger>
          <Tabs.Trigger value="response">Response</Tabs.Trigger>
          <Tabs.Trigger value="webhook">Webhook</Tabs.Trigger>
          <Tabs.Trigger value="example">Example curl</Tabs.Trigger>
        </Tabs.List>

        <Tabs.Content value="ops">
          <div className="space-y-3 mt-4">
            <Heading size="3">Supported operations</Heading>
            <Table.Root>
              <Table.Header>
                <Table.Row>
                  <Table.ColumnHeaderCell>Operation</Table.ColumnHeaderCell>
                  <Table.ColumnHeaderCell>Method + path</Table.ColumnHeaderCell>
                </Table.Row>
              </Table.Header>
              <Table.Body>
                {props.operations.map((op) => (
                  <Table.Row key={op.path + op.method}>
                    <Table.Cell>
                      <div>{op.label}</div>
                      {op.note ? (
                        <div className="text-xs text-muted-foreground mt-0.5">{op.note}</div>
                      ) : null}
                    </Table.Cell>
                    <Table.Cell>
                      <code className="text-xs">
                        {op.method} {op.path}
                      </code>
                    </Table.Cell>
                  </Table.Row>
                ))}
              </Table.Body>
            </Table.Root>
            {props.opsNote ? (
              <Text size="2" color="gray" className="block mt-2">
                {props.opsNote}
              </Text>
            ) : null}
          </div>
        </Tabs.Content>

        <Tabs.Content value="request">
          <div className="space-y-3 mt-4">
            <Heading size="3">Request fields we read</Heading>
            <Text size="2" color="gray">
              The handler parses the request body once and passes it to the
              adapter. These are the fields the code actually reads:
            </Text>
            <Table.Root>
              <Table.Header>
                <Table.Row>
                  <Table.ColumnHeaderCell>Field</Table.ColumnHeaderCell>
                  <Table.ColumnHeaderCell>What it does</Table.ColumnHeaderCell>
                </Table.Row>
              </Table.Header>
              <Table.Body>
                {props.requestFields.map((f) => (
                  <Table.Row key={f.field}>
                    <Table.Cell>
                      <code className="text-xs">{f.field}</code>
                    </Table.Cell>
                    <Table.Cell>{f.description}</Table.Cell>
                  </Table.Row>
                ))}
                <Table.Row>
                  <Table.Cell>
                    <code className="text-xs">X-Webhook-URL</code>
                  </Table.Cell>
                  <Table.Cell>
                    Header. Overrides the workspace&apos;s configured webhook target for this one
                    request.
                  </Table.Cell>
                </Table.Row>
                <Table.Row>
                  <Table.Cell>
                    <code className="text-xs">X-TestPay-Scenario-ID</code>
                  </Table.Cell>
                  <Table.Cell>
                    Header. Picks a specific scenario for this request (skips active session /
                    workspace default resolution).
                  </Table.Cell>
                </Table.Row>
                <Table.Row>
                  <Table.Cell>
                    <code className="text-xs">Authorization</code>
                  </Table.Cell>
                  <Table.Cell>
                    <code>Bearer &lt;api_key&gt;</code>. Required in hosted mode; ignored
                    in local.
                  </Table.Cell>
                </Table.Row>
              </Table.Body>
            </Table.Root>
          </div>
        </Tabs.Content>

        <Tabs.Content value="response">
          <div className="space-y-4 mt-4">
            <div>
              <Heading size="3" mb="2">
                Success response
              </Heading>
              {props.successNotes ? (
                <Text size="2" color="gray" className="block mb-2">
                  {props.successNotes}
                </Text>
              ) : null}
              <CodeBlock language="json">{props.successBody}</CodeBlock>
            </div>

            <div>
              <Heading size="3" mb="2">
                Failure responses
              </Heading>
              <Text size="2" color="gray" className="block mb-2">
                One example per error class. Outcome names refer to the{' '}
                <code>outcome</code> field on a scenario step — see the Failure
                modes reference.
              </Text>
              {props.errorExamples.map((ex) => (
                <div key={ex.label} className="mb-3">
                  <div className="flex items-center gap-2 mb-1">
                    <span className="text-[11px] font-mono px-1.5 py-0.5 rounded bg-[var(--red-a3)] text-[var(--red-11)]">
                      {ex.status}
                    </span>
                    <code className="text-xs">{ex.label}</code>
                  </div>
                  <CodeBlock language="json">{ex.body}</CodeBlock>
                </div>
              ))}
            </div>
          </div>
        </Tabs.Content>

        <Tabs.Content value="webhook">
          <div className="space-y-3 mt-4">
            <Heading size="3">Webhook payload</Heading>
            {props.webhookNotes ? (
              <Text size="2" color="gray" className="block">
                {props.webhookNotes}
              </Text>
            ) : null}
            <CodeBlock language="json">{props.webhookBody}</CodeBlock>

            <Heading size="3" mt="4">
              What gets echoed
            </Heading>
            <Text size="2" color="gray" className="block">
              TestPay preserves these fields from the request into the
              response and/or the webhook so your code can correlate:
            </Text>
            <Card>
              <Box p="3">
                <ul className="list-disc pl-5 space-y-1.5 text-sm">
                  {props.echoedFields.map((f) => (
                    <li key={f.field}>
                      <code className="text-xs">{f.field}</code> — {f.description}
                    </li>
                  ))}
                  <li>
                    <code className="text-xs">request_echo</code> — the full parsed
                    request body is attached to every webhook payload (at the
                    top level) so you can read any custom field without needing
                    to extend this adapter.
                  </li>
                </ul>
              </Box>
            </Card>
          </div>
        </Tabs.Content>

        <Tabs.Content value="example">
          <div className="space-y-3 mt-4">
            <Heading size="3">Example curl</Heading>
            <CodeBlock language="bash">{props.curlExample}</CodeBlock>
          </div>
        </Tabs.Content>
      </Tabs.Root>
    </div>
  );
}
