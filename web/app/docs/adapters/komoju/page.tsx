'use client';
import { AdapterPage } from '@/components/docs/adapter-page';
import { getApiBaseUrl } from '@/components/docs/base-url';

export default function KomojuAdapterPage() {
  const base = getApiBaseUrl();
  const adapterBase = `${base}/komoju`;

  return (
    <AdapterPage
      title="Komoju (Japan)"
      slug="komoju"
      blurb={`Resource-oriented JSON — { id, resource: "payment", status, amount, ... } on success, { error: { code, message, ... } } on failure. Webhook wraps the payment in a { id, type: "payment.<state>", resource: "event", data } envelope. JPY has no minor units, so amount is a plain integer.`}
      baseUrl={adapterBase}
      operations={[
        { label: 'Create payment', method: 'POST', path: '/komoju/api/v1/payments' },
        { label: 'Capture', method: 'POST', path: '/komoju/api/v1/payments/:id/capture' },
        { label: 'Refund', method: 'POST', path: '/komoju/api/v1/payments/:id/refund' },
      ]}
      requestFields={[
        {
          field: 'amount',
          description: 'Integer, JPY whole units. Defaults to 5000.',
        },
        {
          field: 'currency',
          description: 'ISO 4217. Defaults to "JPY".',
        },
        {
          field: 'metadata',
          description:
            'Arbitrary key/value map. Echoed verbatim on the webhook data.metadata.',
        },
      ]}
      successBody={`{
  "id":                 "komoju_1717000000000000000",
  "resource":           "payment",
  "status":             "captured",
  "amount":             5000,
  "tax":                0,
  "total":              5000,
  "currency":           "JPY",
  "description":        "Mock Komoju charge",
  "payment_method_fee": 0,
  "payment_details": {
    "type":         "credit_card",
    "email":        "mock@example.com",
    "redirect_url": "https://komoju.com/mock/komoju_..."
  },
  "captured_at":      "2026-04-17T12:34:56Z",
  "created_at":       "2026-04-17T12:34:56Z",
  "amount_refunded":  0,
  "locale":           "ja",
  "metadata":         {},
  "refunds":          []
}`}
      successNotes={`status follows Komoju's lifecycle: captured (success) | authorized (pending) | failed | cancelled (for success_then_reversed). captured_at is null unless status == "captured".`}
      errorExamples={[
        {
          label: 'bank_decline_hard (402)',
          status: 402,
          body: `{
  "error": {
    "code":    "card_declined",
    "message": "The card was declined.",
    "param":   "bank_decline_hard"
  }
}`,
        },
        {
          label: 'pg_rate_limited (429)',
          status: 429,
          body: `{
  "error": {
    "code":    "rate_limit_exceeded",
    "message": "Too many requests.",
    "param":   "pg_rate_limited"
  }
}`,
        },
        {
          label: 'pg_server_error (500)',
          status: 500,
          body: `{
  "error": {
    "code":    "internal_server_error",
    "message": "An internal error occurred.",
    "param":   "pg_server_error"
  }
}`,
        },
      ]}
      webhookBody={`{
  "id":        "event_1717000000000000000",
  "type":      "payment.captured",
  "resource":  "event",
  "created_at":"2026-04-17T12:34:56Z",
  "data": {
    "id":                 "<charge_id>",
    "resource":           "payment",
    "status":             "captured",
    "amount":             5000,
    "total":              5000,
    "currency":           "JPY",
    "payment_method_fee": 0,
    "payment_details":    { "type": "credit_card", "email": "mock@example.com" },
    "captured_at":        "2026-04-17T12:34:56Z",
    "created_at":         "2026-04-17T12:34:56Z",
    "metadata": {
      "order_id": "ord_123"
    },
    "refunds": []
  },
  "request_echo": { /* full request body */ }
}`}
      webhookNotes={`type is "payment.captured" on success, "payment.authorized" on pending, "payment.failed" on error, "payment.refunded" on success_then_reversed. On failure, an extra details.error_code field is populated at the event level.`}
      echoedFields={[
        {
          field: 'metadata',
          description: 'Copied from the request into data.metadata on the webhook.',
        },
      ]}
      curlExample={`curl -X POST ${adapterBase}/api/v1/payments \\
  -H "Authorization: Bearer <api_key>" \\
  -H "Content-Type: application/json" \\
  -d '{
    "amount":   5000,
    "currency": "JPY",
    "metadata": { "order_id": "ord_123" }
  }'`}
    />
  );
}
