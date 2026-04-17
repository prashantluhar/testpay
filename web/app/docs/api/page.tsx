'use client';
import { Heading, Text, Table, Card, Box } from '@radix-ui/themes';
import { CodeBlock } from '@/components/docs/code-block';

interface Endpoint {
  method: 'GET' | 'POST' | 'PUT' | 'DELETE';
  path: string;
  auth: 'session' | 'public' | 'either';
  summary: string;
  request?: string;
  response?: string;
  notes?: string;
}

const AUTH_ENDPOINTS: Endpoint[] = [
  {
    method: 'POST',
    path: '/api/auth/signup',
    auth: 'public',
    summary:
      'Creates a workspace + owner user. Issues a session cookie. Rate limited 10/min per IP.',
    request: `{
  "email": "you@example.com",
  "password": "at-least-8-chars"
}`,
    response: `{
  "user":      { "id": "...", "email": "...", "role": "owner", ... },
  "workspace": { "id": "...", "slug": "you-xxxx", "api_key": "...", ... }
}`,
    notes: 'Returns 409 if email already exists, 400 on invalid body.',
  },
  {
    method: 'POST',
    path: '/api/auth/login',
    auth: 'public',
    summary: 'Validates credentials, issues a session cookie. Rate limited 10/min per IP.',
    request: `{ "email": "...", "password": "..." }`,
    response: `{ "user": {...}, "workspace": {...} }`,
    notes: '401 on bad credentials.',
  },
  {
    method: 'POST',
    path: '/api/auth/logout',
    auth: 'either',
    summary: 'Clears the session cookie.',
    response: '204 No Content',
  },
  {
    method: 'GET',
    path: '/api/auth/me',
    auth: 'either',
    summary: 'Returns the current session user + workspace.',
    response: `{ "user": {...} | null, "workspace": {...} }`,
    notes:
      'Local mode returns user=null and the "local" workspace. Hosted mode returns 401 if no valid session.',
  },
];

const CONTROL_ENDPOINTS: Endpoint[] = [
  {
    method: 'GET',
    path: '/api/gateways',
    auth: 'session',
    summary: 'List the canonical gateway names registered on this server.',
    response: `["adyen", "agnostic", "ecpay", "epay", "espay", "instamojo", "komoju", "mastercard", "omise", "payletter", "paynamics", "razorpay", "stripe", "tappay", "tillpay"]`,
  },
  {
    method: 'GET',
    path: '/api/workspace',
    auth: 'session',
    summary: 'Fetch the workspace for the current session.',
    response: `{
  "id": "...",
  "slug": "...",
  "api_key": "...",
  "webhook_urls": { "_default": "https://...", "stripe": "https://..." },
  "created_at": "..."
}`,
  },
  {
    method: 'PUT',
    path: '/api/workspace',
    auth: 'session',
    summary: 'Update workspace fields. Currently only webhook_urls is editable.',
    request: `{
  "webhook_urls": {
    "_default": "https://my-app.example.com/hooks/testpay",
    "stripe":   "https://my-app.example.com/hooks/stripe"
  }
}`,
    response: `Updated workspace object.`,
    notes:
      'Empty-string values are stripped server-side. Missing keys in the payload are left unchanged.',
  },
];

const SCENARIO_ENDPOINTS: Endpoint[] = [
  {
    method: 'GET',
    path: '/api/scenarios',
    auth: 'session',
    summary: 'List all scenarios in the workspace.',
    response: `[
  {
    "id": "...",
    "name": "Decline then succeed",
    "description": "",
    "gateway": "stripe",
    "steps": [
      { "event": "charge", "outcome": "bank_decline_hard" },
      { "event": "charge", "outcome": "success" }
    ],
    "webhook_delay_ms": 0,
    "is_default": false,
    "created_at": "..."
  }
]`,
  },
  {
    method: 'POST',
    path: '/api/scenarios',
    auth: 'session',
    summary: 'Create a scenario.',
    request: `{
  "name": "My scenario",
  "description": "optional",
  "gateway": "stripe",
  "steps": [
    { "event": "charge", "outcome": "success" }
  ],
  "webhook_delay_ms": 0,
  "is_default": false
}`,
    response: 'The created scenario (201).',
  },
  {
    method: 'GET',
    path: '/api/scenarios/{id}',
    auth: 'session',
    summary: 'Fetch one scenario by id.',
    response: 'Scenario object. 404 if not found.',
  },
  {
    method: 'PUT',
    path: '/api/scenarios/{id}',
    auth: 'session',
    summary: 'Replace a scenario.',
    request: 'Full Scenario object (id in URL takes precedence).',
    response: 'Updated scenario.',
  },
  {
    method: 'DELETE',
    path: '/api/scenarios/{id}',
    auth: 'session',
    summary: 'Delete a scenario. 204 on success.',
  },
  {
    method: 'POST',
    path: '/api/scenarios/{id}/run',
    auth: 'session',
    summary:
      'Execute every step in the scenario once (engine-side simulation). Records a scenario_run entry. Does NOT advance any active session call_index.',
    response: `{
  "id": "...", "scenario_id": "...",
  "status": "completed",
  "started_at": "...", "completed_at": "..."
}`,
  },
];

const SESSION_ENDPOINTS: Endpoint[] = [
  {
    method: 'POST',
    path: '/api/sessions',
    auth: 'session',
    summary:
      'Create an active session binding a scenario to this workspace. Subsequent mock requests walk the scenario step list using the session call_index.',
    request: `{
  "scenario_id": "...",
  "ttl_seconds": 3600
}`,
    response: `{
  "id": "...", "workspace_id": "...",
  "scenario_id": "...",
  "ttl_seconds": 3600,
  "expires_at": "..."
}`,
    notes: 'ttl_seconds defaults to 3600 (1 hour). Only one session is active per workspace at a time.',
  },
  {
    method: 'DELETE',
    path: '/api/sessions/{id}',
    auth: 'session',
    summary: 'Delete an active session. 204 regardless (idempotent).',
  },
];

const LOG_ENDPOINTS: Endpoint[] = [
  {
    method: 'GET',
    path: '/api/logs',
    auth: 'session',
    summary:
      'List request_logs for the workspace, newest first. Pagination via ?limit= (default 50) and ?offset=.',
    response: 'Array of RequestLog objects — see web/lib/types.ts.',
  },
  {
    method: 'GET',
    path: '/api/logs/{id}',
    auth: 'session',
    summary: 'Fetch one request_log + its linked webhook_log (if any).',
    response: `{
  "request": { ...RequestLog },
  "webhook": { ...WebhookLog } | null
}`,
  },
  {
    method: 'POST',
    path: '/api/logs/{id}/replay',
    auth: 'session',
    summary:
      'Re-run the adapter against the logged request path. Uses an always-succeed scenario — this is for smoke-testing response shape, not state replay.',
    notes:
      'Does NOT create a new request_log. Response is written directly back to the caller.',
  },
];

const WEBHOOK_ENDPOINTS: Endpoint[] = [
  {
    method: 'GET',
    path: '/api/webhooks',
    auth: 'session',
    summary:
      'List webhook_logs for the workspace, newest first. Pagination via ?limit= / ?offset=.',
    response: 'Array of WebhookLog — includes attempt_logs (per-retry capture).',
  },
  {
    method: 'GET',
    path: '/api/webhooks/{id}',
    auth: 'session',
    summary: 'Fetch one webhook_log by its id.',
  },
  {
    method: 'GET',
    path: '/api/webhooks/{id}/status',
    auth: 'session',
    summary:
      'Status-only lookup keyed by the request_log_id (not webhook_log id). Used to poll delivery status for a given request.',
  },
  {
    method: 'POST',
    path: '/api/webhooks/test',
    auth: 'session',
    summary:
      'Send an arbitrary payload to a target URL using the production dispatcher (exercises retries + attempt logging). Useful for verifying your webhook endpoint.',
    request: `{
  "target_url": "https://...",
  "payload": { ... }
}`,
    response: `{
  "status_code": 200,
  "attempts": 1,
  "attempt_logs": [...]
}`,
    notes: 'Returns 502 with the attempt_logs if delivery fails after all retries.',
  },
];

function EndpointTable({ endpoints }: { endpoints: Endpoint[] }) {
  return (
    <div className="space-y-4">
      {endpoints.map((ep) => (
        <Card key={`${ep.method}-${ep.path}`}>
          <Box p="3">
            <div className="flex items-center gap-2 mb-2 flex-wrap">
              <span
                className={`text-[11px] font-mono px-1.5 py-0.5 rounded ${
                  ep.method === 'GET'
                    ? 'bg-[var(--blue-a3)] text-[var(--blue-11)]'
                    : ep.method === 'POST'
                      ? 'bg-[var(--green-a3)] text-[var(--green-11)]'
                      : ep.method === 'PUT'
                        ? 'bg-[var(--amber-a3)] text-[var(--amber-11)]'
                        : 'bg-[var(--red-a3)] text-[var(--red-11)]'
                }`}
              >
                {ep.method}
              </span>
              <code className="text-sm font-mono">{ep.path}</code>
              <span className="text-[10px] uppercase tracking-wider text-muted-foreground ml-auto">
                auth: {ep.auth}
              </span>
            </div>
            <Text size="2" color="gray" className="block mb-2">
              {ep.summary}
            </Text>
            {ep.request ? (
              <div className="mt-2">
                <div className="text-[10px] uppercase tracking-wider text-muted-foreground mb-1">
                  Request body
                </div>
                <CodeBlock language="json">{ep.request}</CodeBlock>
              </div>
            ) : null}
            {ep.response ? (
              <div className="mt-2">
                <div className="text-[10px] uppercase tracking-wider text-muted-foreground mb-1">
                  Response
                </div>
                <CodeBlock language="json">{ep.response}</CodeBlock>
              </div>
            ) : null}
            {ep.notes ? (
              <Text size="1" color="gray" className="block mt-2 italic">
                {ep.notes}
              </Text>
            ) : null}
          </Box>
        </Card>
      ))}
    </div>
  );
}

export default function ApiReferencePage() {
  return (
    <div className="space-y-6">
      <div>
        <Heading size="7" mb="2">
          API reference
        </Heading>
        <Text size="3" color="gray">
          The Control API — everything under <code>/api/*</code>. These
          endpoints drive the dashboard and can be hit programmatically.
        </Text>
      </div>

      <Card>
        <Box p="3">
          <Heading size="3" mb="2">
            Auth model
          </Heading>
          <Table.Root>
            <Table.Header>
              <Table.Row>
                <Table.ColumnHeaderCell>Where you&apos;re calling from</Table.ColumnHeaderCell>
                <Table.ColumnHeaderCell>What to send</Table.ColumnHeaderCell>
              </Table.Row>
            </Table.Header>
            <Table.Body>
              <Table.Row>
                <Table.Cell>Browser (dashboard, same origin)</Table.Cell>
                <Table.Cell>Session cookie — set by signup/login, sent automatically.</Table.Cell>
              </Table.Row>
              <Table.Row>
                <Table.Cell>Server-to-server, user-scoped</Table.Cell>
                <Table.Cell>
                  Set the <code>testpay_session</code> cookie manually, or login first.
                </Table.Cell>
              </Table.Row>
              <Table.Row>
                <Table.Cell>Mock endpoints only (/stripe/*, /v1/*, etc.)</Table.Cell>
                <Table.Cell>
                  <code>Authorization: Bearer &lt;workspace_api_key&gt;</code>. API key auth is
                  NOT accepted on <code>/api/*</code>.
                </Table.Cell>
              </Table.Row>
            </Table.Body>
          </Table.Root>
        </Box>
      </Card>

      <div>
        <Heading size="4" mb="3">
          Auth
        </Heading>
        <EndpointTable endpoints={AUTH_ENDPOINTS} />
      </div>

      <div>
        <Heading size="4" mb="3">
          Workspace + gateways
        </Heading>
        <EndpointTable endpoints={CONTROL_ENDPOINTS} />
      </div>

      <div>
        <Heading size="4" mb="3">
          Scenarios
        </Heading>
        <EndpointTable endpoints={SCENARIO_ENDPOINTS} />
      </div>

      <div>
        <Heading size="4" mb="3">
          Sessions
        </Heading>
        <EndpointTable endpoints={SESSION_ENDPOINTS} />
      </div>

      <div>
        <Heading size="4" mb="3">
          Logs
        </Heading>
        <EndpointTable endpoints={LOG_ENDPOINTS} />
      </div>

      <div>
        <Heading size="4" mb="3">
          Webhooks
        </Heading>
        <EndpointTable endpoints={WEBHOOK_ENDPOINTS} />
      </div>
    </div>
  );
}
