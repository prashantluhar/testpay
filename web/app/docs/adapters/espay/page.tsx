'use client';
import { AdapterPage } from '@/components/docs/adapter-page';
import { getApiBaseUrl } from '@/components/docs/base-url';

export default function EspayAdapterPage() {
  const base = getApiBaseUrl();
  const adapterBase = `${base}/espay`;

  return (
    <AdapterPage
      title="ESPay (Indonesia)"
      slug="espay"
      blurb={`RSA-signed JSON payloads with an error_code / error_message envelope. "0000" is success; anything else is a failure (ESPxx-style codes). signature field is a fixed mock placeholder — consumers validating signatures should detect it and skip verification.`}
      baseUrl={adapterBase}
      operations={[
        {
          label: 'Inquire payment',
          method: 'POST',
          path: '/espay/rest/inquiry',
        },
        {
          label: 'Payment status',
          method: 'POST',
          path: '/espay/rest/paymentstatus',
        },
      ]}
      requestFields={[
        { field: 'amount', description: 'Numeric. Defaults to 5000. Echoed as STRING in response.' },
        {
          field: 'ccy',
          description: 'Currency. Accepts fallback key "currency". Defaults to "IDR".',
        },
      ]}
      successBody={`{
  "rq_uuid":       "rq_1717000000000000000",
  "rs_datetime":   "20260417123456",
  "error_code":    "0000",
  "error_message": "Success",
  "signature":     "mock-rsa-signature",
  "order_id":      "ESP_1717000000000000000",
  "amount":        "5000",
  "ccy":           "IDR",
  "description":   "Mock ESPay charge",
  "trx_date":      "20260417123456",
  "customer_details": {
    "firstname":    "Mock",
    "lastname":     "Customer",
    "phone_number": "081200000000",
    "email":        "mock@example.com"
  }
}`}
      successNotes={`Pending modes return HTTP 200 with error_code "0001" / error_message "Processing" — still a non-success response from the caller's perspective.`}
      errorExamples={[
        {
          label: 'bank_decline_hard (402)',
          status: 402,
          body: `{
  "rq_uuid":       "rq_1717000000000000000",
  "rs_datetime":   "20260417123456",
  "error_code":    "ESP02",
  "error_message": "Transaction declined",
  "signature":     "mock-rsa-signature",
  "order_id":      "ESP_1717000000000000000"
}`,
        },
        {
          label: 'pg_server_error (500)',
          status: 500,
          body: `{
  "rq_uuid":       "rq_1717000000000000000",
  "rs_datetime":   "20260417123456",
  "error_code":    "ESP99",
  "error_message": "Internal server error",
  "signature":     "mock-rsa-signature",
  "order_id":      "ESP_1717000000000000000"
}`,
        },
        {
          label: 'pg_rate_limited (429)',
          status: 429,
          body: `{
  "rq_uuid":       "rq_1717000000000000000",
  "rs_datetime":   "20260417123456",
  "error_code":    "ESP05",
  "error_message": "Too many requests",
  "signature":     "mock-rsa-signature",
  "order_id":      "ESP_1717000000000000000"
}`,
        },
      ]}
      webhookBody={`{
  "rq_uuid":             "rq_1717000000000000001",
  "rs_datetime":         "20260417123500",
  "error_code":          "0000",
  "error_message":       "Success",
  "signature":           "mock-rsa-signature",
  "order_id":            "<charge_id>",
  "reconcile_id":        "RCN_1717000000000000000",
  "reconcile_datetime":  "20260417123500",
  "amount":              5000,
  "currency":            "IDR",
  "request_echo":        { /* full request body */ }
}`}
      webhookNotes={`amount + currency are added as top-level fields on the webhook for convenience — ESPay's native envelope doesn't carry them. On failures, error_code / error_message carry the failure details.`}
      echoedFields={[
        {
          field: 'amount / currency',
          description: 'Surfaced on the webhook (the only adapter that flattens these outside the primary envelope).',
        },
        {
          field: 'Whole request body',
          description: 'Attached to the webhook as request_echo.',
        },
      ]}
      curlExample={`curl -X POST ${adapterBase}/rest/inquiry \\
  -H "Authorization: Bearer <api_key>" \\
  -H "Content-Type: application/json" \\
  -d '{
    "amount": 5000,
    "ccy":    "IDR"
  }'`}
    />
  );
}
