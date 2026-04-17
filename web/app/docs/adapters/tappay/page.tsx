'use client';
import { AdapterPage } from '@/components/docs/adapter-page';
import { getApiBaseUrl } from '@/components/docs/base-url';

export default function TappayAdapterPage() {
  const base = getApiBaseUrl();
  const adapterBase = `${base}/tappay`;

  return (
    <AdapterPage
      title="TapPay (Appotapay)"
      slug="tappay"
      blurb={`"Zero is success" — errorCode 0 means success, 1 means processing, anything else is a failure. HMAC-SHA256-signed in production; the mock emits a fixed signature placeholder. Wire format mirrors Appotapay's DTO package.`}
      baseUrl={adapterBase}
      operations={[
        { label: 'Create payment', method: 'POST', path: '/tappay/api/v1/payments' },
      ]}
      requestFields={[
        {
          field: 'amount',
          description: 'Integer, minor units. Defaults to 5000. Echoed as int on webhook.',
        },
        {
          field: 'currency',
          description: 'ISO 4217. Defaults to "TWD".',
        },
      ]}
      successBody={`{
  "errorCode":  0,
  "message":    "Success",
  "orderId":    "TAP_1717000000000000000",
  "amount":     5000,
  "paymentUrl": "https://mock.tappay.io/pay/TAP_...",
  "signature":  "mock-hmac-signature"
}`}
      successNotes={`Pending modes return HTTP 200 with errorCode 1 and message "Processing". paymentUrl drives the hosted payment flow — check errorCode == 0 before trusting it.`}
      errorExamples={[
        {
          label: 'bank_decline_hard (402)',
          status: 402,
          body: `{
  "errorCode": 2,
  "message":   "Transaction declined",
  "signature": "mock-hmac-signature",
  "orderId":   "TAP_1717000000000000000"
}`,
        },
        {
          label: 'pg_rate_limited (429)',
          status: 429,
          body: `{
  "errorCode": 6,
  "message":   "Too many requests",
  "signature": "mock-hmac-signature",
  "orderId":   "TAP_..."
}`,
        },
        {
          label: 'pg_server_error (500)',
          status: 500,
          body: `{
  "errorCode": 99,
  "message":   "Internal server error",
  "signature": "mock-hmac-signature",
  "orderId":   "TAP_..."
}`,
        },
      ]}
      webhookBody={`{
  "errorCode":       0,
  "message":         "Success",
  "partnerCode":     "MOCK_PARTNER",
  "apiKey":          "MOCK_API_KEY",
  "amount":          5000,
  "currency":        "TWD",
  "orderId":         "<charge_id>",
  "bankCode":        "MOCK_BANK",
  "paymentMethod":   "ATM",
  "paymentType":     "BANK_TRANSFER",
  "appotapayTransId":"APT_1717000000000000000",
  "transactionTs":   1717000000,
  "signature":       "mock-hmac-signature",
  "request_echo":    { /* full request body */ }
}`}
      webhookNotes={`partnerCode and apiKey are deterministic mock values — real Appotapay uses per-merchant values. The webhook carries a Unix-timestamp transactionTs and a paymentMethod/paymentType pair.`}
      echoedFields={[
        {
          field: 'amount / currency',
          description: 'Copied onto the webhook as int and string respectively.',
        },
      ]}
      curlExample={`curl -X POST ${adapterBase}/api/v1/payments \\
  -H "Authorization: Bearer <api_key>" \\
  -H "Content-Type: application/json" \\
  -d '{
    "amount":   5000,
    "currency": "TWD"
  }'`}
    />
  );
}
