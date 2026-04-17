'use client';
import { AdapterPage } from '@/components/docs/adapter-page';
import { getApiBaseUrl } from '@/components/docs/base-url';

export default function AgnosticAdapterPage() {
  const base = getApiBaseUrl();
  const adapterBase = `${base}/v1`;

  return (
    <AdapterPage
      title="Agnostic (/v1)"
      slug="agnostic"
      blurb={`The generic adapter for callers who don't need gateway-authentic shapes. Responses are a minimal { id, status, error_code } envelope; webhooks carry the full request body under request_echo so you can pluck any field you want.`}
      baseUrl={adapterBase}
      operations={[
        { label: 'Create charge', method: 'POST', path: '/v1/charges' },
        { label: 'Capture', method: 'POST', path: '/v1/charges/:id/capture' },
        { label: 'Refund', method: 'POST', path: '/v1/refunds' },
        { label: 'Retrieve', method: 'GET', path: '/v1/charges/:id' },
      ]}
      opsNote="Any path under /v1/* reaches this adapter. The URL prefix /agnostic/* is registered as an alias and behaves identically."
      requestFields={[
        {
          field: 'amount',
          description: 'Numeric. Defaults to 5000. Echoed on the webhook.',
        },
        {
          field: 'currency',
          description: 'Defaults to "usd". Echoed on the webhook.',
        },
        {
          field: 'order_id / metadata.* / notes.*',
          description:
            'Anything you send is available on the webhook as request_echo — no need to extend this adapter to add custom correlation keys.',
        },
      ]}
      successBody={`{
  "id":         "txn_1717000000000000000",
  "status":     "success",
  "error_code": ""
}`}
      successNotes="status is 'success' | 'pending' | 'failed'. The HTTP status follows the engine result directly (so webhook_missing modes still produce status: 'success' because the HTTP layer is happy — the side effect is just no webhook)."
      errorExamples={[
        {
          label: 'bank_decline_hard (402)',
          status: 402,
          body: `{
  "id":         "txn_1717000000000000000",
  "status":     "failed",
  "error_code": "bank_decline_hard"
}`,
        },
        {
          label: 'pg_rate_limited (429)',
          status: 429,
          body: `{
  "id":         "txn_1717000000000000000",
  "status":     "failed",
  "error_code": "pg_rate_limited"
}`,
        },
        {
          label: 'pg_server_error (500)',
          status: 500,
          body: `{
  "id":         "txn_1717000000000000000",
  "status":     "failed",
  "error_code": ""
}`,
        },
      ]}
      webhookBody={`{
  "event":      "transaction.success",
  "id":         "<charge_id>",
  "amount":     5000,
  "currency":   "usd",
  "timestamp":  1717000000,
  "error_code": "",
  "request_echo": {
    "amount":   5000,
    "currency": "usd",
    "order_id": "ord_123",
    "metadata": { "customer_id": "cust_42" }
  }
}`}
      webhookNotes="event is 'transaction.success' or 'transaction.failed'. The agnostic adapter is the only one that doesn't filter the echo — you get back your entire request body verbatim."
      echoedFields={[
        {
          field: 'Whole request body',
          description:
            'Copied into webhook.request_echo. Use this for any merchant-side correlation.',
        },
        {
          field: 'amount / currency',
          description: 'Also surfaced as top-level fields on the webhook for convenience.',
        },
      ]}
      curlExample={`curl -X POST ${adapterBase}/charges \\
  -H "Authorization: Bearer <api_key>" \\
  -H "Content-Type: application/json" \\
  -d '{
    "amount":   5000,
    "currency": "usd",
    "order_id": "ord_123",
    "metadata": { "customer_id": "cust_42" }
  }'`}
    />
  );
}
