'use client';
import { AdapterPage } from '@/components/docs/adapter-page';
import { getApiBaseUrl } from '@/components/docs/base-url';

export default function PaynamicsAdapterPage() {
  const base = getApiBaseUrl();
  const adapterBase = `${base}/paynamics`;

  return (
    <AdapterPage
      title="Paynamics (PH / SEA)"
      slug="paynamics"
      blurb={`MD5-signed JSON with a response_code / response_advise / response_message triple. "GR001" = success; "GR051" and GR0xx codes cover failures. signature is a fixed mock placeholder. total_amount is a decimal STRING ("50.00" not 5000).`}
      baseUrl={adapterBase}
      operations={[
        { label: 'Initialize transaction', method: 'POST', path: '/paynamics/pnxpay/initialize' },
        { label: 'Query status', method: 'POST', path: '/paynamics/pnxpay/status' },
      ]}
      requestFields={[
        {
          field: 'amount',
          description:
            'Numeric at the top level OR nested under transaction.amount (Paynamics supports both). String amounts under transaction.amount are parsed as decimal (e.g. "50.00" → 5000 minor units). Defaults to 5000.',
        },
        {
          field: 'currency',
          description:
            'Top-level or transaction.currency. Defaults to "PHP".',
        },
      ]}
      successBody={`{
  "response_code":    "GR001",
  "response_advise":  "SUCCESS",
  "response_message": "Transaction successful",
  "signature":        "mock-md5-signature",
  "response_id":      "PNMC_1717000000000000000",
  "merchant_id":      "MOCK_MERCHANT",
  "request_id":       "REQ_1717000000000000000",
  "redirect_url":     "https://mock.paynamics.net/pay/PNMC_...",
  "timestamp":        "2026-04-17T12:34:56Z",
  "currency":         "PHP",
  "total_amount":     "50.00"
}`}
      successNotes={`Pending modes return HTTP 200 with response_code "GR002", response_advise "PENDING". total_amount is a string with 2 decimals.`}
      errorExamples={[
        {
          label: 'bank_decline_hard (402)',
          status: 402,
          body: `{
  "response_code":    "GR051",
  "response_advise":  "DECLINED",
  "response_message": "Transaction declined by issuer",
  "signature":        "mock-md5-signature",
  "response_id":      "PNMC_...",
  "merchant_id":      "MOCK_MERCHANT",
  "request_id":       "REQ_...",
  "timestamp":        "2026-04-17T12:34:56Z"
}`,
        },
        {
          label: 'pg_rate_limited (429)',
          status: 429,
          body: `{
  "response_code":    "GR055",
  "response_advise":  "RATE_LIMITED",
  "response_message": "Too many requests",
  "signature":        "mock-md5-signature"
}`,
        },
        {
          label: 'pg_server_error (500)',
          status: 500,
          body: `{
  "response_code":    "GR099",
  "response_advise":  "SYSTEM_ERROR",
  "response_message": "Internal server error",
  "signature":        "mock-md5-signature"
}`,
        },
      ]}
      webhookBody={`{
  "response_code":    "GR001",
  "response_advise":  "SUCCESS",
  "response_message": "Transaction successful",
  "signature":        "mock-md5-signature",
  "response_id":      "<charge_id>",
  "merchant_id":      "MOCK_MERCHANT",
  "request_id":       "REQ_1717000000000000000",
  "timestamp":        "2026-04-17T12:34:56Z",
  "total_amount":     "50.00",
  "currency":         "PHP",
  "request_echo":     { /* full request body */ }
}`}
      webhookNotes={`response_code mirrors the initial response for success/pending. On failure (HTTP 4xx) the webhook still fires with response_code/response_advise/response_message reflecting the failure.`}
      echoedFields={[
        {
          field: 'currency',
          description: 'Passed from the request into the response + webhook.',
        },
        {
          field: 'Whole request body',
          description: 'Attached to the webhook as request_echo.',
        },
      ]}
      curlExample={`curl -X POST ${adapterBase}/pnxpay/initialize \\
  -H "Authorization: Bearer <api_key>" \\
  -H "Content-Type: application/json" \\
  -d '{
    "transaction": { "amount": "50.00", "currency": "PHP" }
  }'`}
    />
  );
}
