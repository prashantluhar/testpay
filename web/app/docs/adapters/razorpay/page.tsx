'use client';
import { AdapterPage } from '@/components/docs/adapter-page';
import { getApiBaseUrl } from '@/components/docs/base-url';

export default function RazorpayAdapterPage() {
  const base = getApiBaseUrl();
  const adapterBase = `${base}/razorpay`;

  return (
    <AdapterPage
      title="Razorpay"
      slug="razorpay"
      blurb="Mirrors Razorpay's payments API: payment entity with id/entity/amount/currency/status, plus the { entity: 'event', event, contains, payload, created_at } webhook envelope."
      baseUrl={adapterBase}
      operations={[
        { label: 'Create payment', method: 'POST', path: '/razorpay/v1/payments' },
        { label: 'Capture', method: 'POST', path: '/razorpay/v1/payments/:id/capture' },
        { label: 'Create refund', method: 'POST', path: '/razorpay/v1/refunds' },
        { label: 'Fetch payment', method: 'GET', path: '/razorpay/v1/payments/:id' },
      ]}
      opsNote="Operation routing is substring-based; any /payment path is treated as charge, /capture as capture, /refund as refund. The adapter emits the same paymentEntity shape on all of them."
      requestFields={[
        {
          field: 'amount',
          description: 'Integer, minor units (INR paise). Defaults to 5000. Echoed.',
        },
        {
          field: 'currency',
          description: 'ISO 4217. Defaults to "INR". Echoed.',
        },
        {
          field: 'notes',
          description:
            'Map — Razorpay\'s merchant-side correlation channel. Echoed into response.notes and the webhook entity.notes.',
        },
        {
          field: 'notes.reference',
          description:
            'Commonly used merchant order ID. Echoed back for correlation (appears unchanged on the webhook).',
        },
      ]}
      successBody={`{
  "id":       "pay_1717000000000000000",
  "entity":   "payment",
  "amount":   10000,
  "currency": "INR",
  "status":   "captured",
  "method":   "card",
  "notes": {
    "reference": "order_123"
  }
}`}
      successNotes="status is 'captured' on success, 'authorized' on pending modes, 'failed' on in-band failures like amount_mismatch."
      errorExamples={[
        {
          label: 'bank_decline_hard (402)',
          status: 402,
          body: `{
  "error": {
    "code":        "GATEWAY_ERROR",
    "description": "Payment was declined by the bank",
    "source":      "bank",
    "step":        "payment_authorization",
    "reason":      "bank_decline_hard"
  }
}`,
        },
        {
          label: 'pg_rate_limited (429)',
          status: 429,
          body: `{
  "error": {
    "code":        "BAD_REQUEST_ERROR",
    "description": "Too many requests — rate limit exceeded",
    "source":      "gateway",
    "step":        "payment_initiation",
    "reason":      "pg_rate_limited"
  }
}`,
        },
        {
          label: 'pg_server_error (500)',
          status: 500,
          body: `{
  "error": {
    "code":        "SERVER_ERROR",
    "description": "Internal server error at gateway",
    "source":      "gateway",
    "step":        "payment_initiation",
    "reason":      "pg_server_error"
  }
}`,
        },
      ]}
      webhookBody={`{
  "entity":     "event",
  "event":      "payment.captured",
  "contains":   ["payment"],
  "created_at": 1717000000,
  "payload": {
    "payment": {
      "entity": {
        "id":       "<charge_id>",
        "entity":   "payment",
        "amount":   10000,
        "currency": "INR",
        "status":   "captured",
        "method":   "card",
        "notes":    { "reference": "order_123" }
      }
    }
  },
  "request_echo": {
    "amount": 10000, "currency": "INR",
    "notes":  { "reference": "order_123" }
  }
}`}
      webhookNotes="event is payment.captured on success, payment.authorized on pending, payment.failed on errors. On failure the entity carries error_code / error_description / error_source / error_step / error_reason fields."
      echoedFields={[
        {
          field: 'notes',
          description: 'Full notes map echoed into response.notes and webhook entity.notes.',
        },
        { field: 'amount / currency', description: 'Echoed on both response and webhook.' },
      ]}
      curlExample={`curl -X POST ${adapterBase}/v1/payments \\
  -H "Authorization: Bearer <api_key>" \\
  -H "Content-Type: application/json" \\
  -d '{
    "amount": 10000,
    "currency": "INR",
    "notes": { "reference": "order_123" }
  }'`}
    />
  );
}
