'use client';
import { AdapterPage } from '@/components/docs/adapter-page';
import { getApiBaseUrl } from '@/components/docs/base-url';

export default function InstamojoAdapterPage() {
  const base = getApiBaseUrl();
  const adapterBase = `${base}/instamojo`;

  return (
    <AdapterPage
      title="Instamojo"
      slug="instamojo"
      blurb="Indian payment-request API. Response shape: { success, payment_request: {...} } on OK, { success: false, message, errors } on validation failure. Amounts are STRING-encoded (Instamojo quirk)."
      baseUrl={adapterBase}
      operations={[
        {
          label: 'Create payment-request',
          method: 'POST',
          path: '/instamojo/api/1.1/payment-requests/',
        },
      ]}
      requestFields={[
        {
          field: 'amount',
          description:
            'Accepts numeric OR string. Defaults to 5000. Echoed on response as a STRING.',
        },
        {
          field: 'currency',
          description: 'ISO 4217. Defaults to "INR".',
        },
        {
          field: 'purpose',
          description:
            'Free-text purpose string. Echoed as a top-level purpose field on the webhook — commonly used for order correlation.',
        },
      ]}
      successBody={`{
  "success": true,
  "payment_request": {
    "id":          "MOJO_1717000000000000000",
    "phone":       "9999999999",
    "email":       "mock@example.com",
    "buyer_name":  "Mock Buyer",
    "amount":      "5000",
    "purpose":     "Mock charge",
    "status":      "Sent",
    "send_sms":    false,
    "send_email":  false,
    "longurl":     "https://instamojo.com/@mock/MOJO_...",
    "created_at":  "2026-04-17T12:34:56Z",
    "modified_at": "2026-04-17T12:34:56Z",
    "allow_repeated_payments": false
  }
}`}
      successNotes={`status is "Sent" on success, "Pending" on pending modes. amount is a string.`}
      errorExamples={[
        {
          label: 'bank_decline_hard (402)',
          status: 402,
          body: `{
  "success": false,
  "message": "Payment declined by bank",
  "errors":  { "code": ["bank_decline_hard"] }
}`,
        },
        {
          label: 'pg_rate_limited (429)',
          status: 429,
          body: `{
  "success": false,
  "message": "Too many requests",
  "errors":  { "code": ["pg_rate_limited"] }
}`,
        },
      ]}
      webhookBody={`{
  "payment_id":         "<charge_id>",
  "payment_request_id": "MOJOPR_1717000000000000000",
  "status":             "Credit",
  "amount":             "5000",
  "currency":           "INR",
  "buyer":              "mock@example.com",
  "buyer_name":         "Mock Buyer",
  "buyer_phone":        "9999999999",
  "fees":               "0.00",
  "instrument_type":    "Card",
  "created_at":         "2026-04-17T12:34:56Z",
  "purpose":            "Mock charge",
  "request_echo":       { /* full request body */ }
}`}
      webhookNotes={`status on the webhook: "Credit" (success), "Pending", or "Failed". On failure, failure_reason + failure_message are populated.`}
      echoedFields={[
        {
          field: 'purpose',
          description: 'Echoed as a top-level field on the webhook for easy correlation.',
        },
      ]}
      curlExample={`curl -X POST ${adapterBase}/api/1.1/payment-requests/ \\
  -H "Authorization: Bearer <api_key>" \\
  -H "Content-Type: application/json" \\
  -d '{
    "amount":  "5000",
    "purpose": "Order 123",
    "buyer_name": "Alice"
  }'`}
    />
  );
}
