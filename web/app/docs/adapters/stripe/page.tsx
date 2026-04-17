'use client';
import { AdapterPage } from '@/components/docs/adapter-page';
import { getApiBaseUrl } from '@/components/docs/base-url';

export default function StripeAdapterPage() {
  const base = getApiBaseUrl();
  const adapterBase = `${base}/stripe`;

  return (
    <AdapterPage
      title="Stripe"
      slug="stripe"
      blurb="Drop-in for Stripe's /v1 API. Point your Stripe SDK at this base URL and every charge / payment_intent / webhook fires through TestPay. Responses match Stripe's payment_intent object shape."
      baseUrl={adapterBase}
      operations={[
        { label: 'Create charge / payment intent', method: 'POST', path: '/stripe/v1/charges' },
        {
          label: 'Create payment intent (alt path)',
          method: 'POST',
          path: '/stripe/v1/payment_intents',
        },
        { label: 'Capture', method: 'POST', path: '/stripe/v1/charges/:id/capture' },
        { label: 'Create refund', method: 'POST', path: '/stripe/v1/refunds' },
        { label: 'Retrieve', method: 'GET', path: '/stripe/v1/charges/:id' },
      ]}
      opsNote="Any path under /stripe/* reaches the adapter. The handler's extractOperation() categorises as charge / refund / capture based on substring match — the response shape is the same payment_intent object regardless of path."
      requestFields={[
        {
          field: 'amount',
          description:
            'Integer, minor units. Defaults to 5000 if absent. Echoed into the response + webhook.',
        },
        {
          field: 'currency',
          description: 'ISO 4217, lowercase. Defaults to "usd". Echoed.',
        },
        {
          field: 'metadata.order_id',
          description:
            'Surfaced on the webhook as data.object.metadata.order_id. Commonly used for correlation.',
        },
        {
          field: 'metadata.*',
          description: 'Any metadata keys are passed through verbatim into the webhook.',
        },
      ]}
      successBody={`{
  "id":       "pi_1717000000000000000",
  "object":   "payment_intent",
  "status":   "succeeded",
  "amount":   5000,
  "currency": "usd"
}`}
      successNotes="When the scenario step is a pending mode (pending_then_success etc.) the status field is 'processing' instead of 'succeeded'. Other fields unchanged."
      errorExamples={[
        {
          label: 'bank_decline_hard (402)',
          status: 402,
          body: `{
  "error": {
    "type":    "card_error",
    "code":    "bank_decline_hard",
    "message": "Your card was declined."
  }
}`,
        },
        {
          label: 'pg_rate_limited (429)',
          status: 429,
          body: `{
  "error": {
    "type":    "card_error",
    "code":    "pg_rate_limited",
    "message": "Too many requests."
  }
}`,
        },
        {
          label: 'pg_server_error (500)',
          status: 500,
          body: `{
  "error": {
    "type":    "card_error",
    "code":    "pg_server_error",
    "message": "An unexpected error occurred."
  }
}`,
        },
      ]}
      webhookBody={`{
  "id":      "evt_1717000000000000000",
  "type":    "payment_intent.succeeded",
  "created": 1717000000,
  "data": {
    "object": {
      "id":       "<charge_id>",
      "object":   "payment_intent",
      "amount":   5000,
      "currency": "usd",
      "status":   "succeeded",
      "metadata": {
        "order_id": "ord_123"
      },
      "request_echo": {
        "amount":   5000,
        "currency": "usd",
        "metadata": { "order_id": "ord_123" }
      }
    }
  }
}`}
      webhookNotes={
        'type switches to "payment_intent.payment_failed" on 4xx/5xx outcomes, or "payment_intent.processing" on pending modes. status inside data.object follows Stripe\'s vocabulary: succeeded | processing | canceled | requires_payment_method.'
      }
      echoedFields={[
        {
          field: 'metadata',
          description: 'Complete metadata object passed through to the webhook verbatim.',
        },
        { field: 'amount', description: 'Echoed on both response and webhook.' },
        { field: 'currency', description: 'Echoed on both response and webhook.' },
      ]}
      curlExample={`curl -X POST ${adapterBase}/v1/charges \\
  -H "Authorization: Bearer <api_key>" \\
  -H "Content-Type: application/json" \\
  -d '{
    "amount": 5000,
    "currency": "usd",
    "metadata": { "order_id": "ord_123" }
  }'`}
    />
  );
}
