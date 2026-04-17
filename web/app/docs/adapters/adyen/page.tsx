'use client';
import { AdapterPage } from '@/components/docs/adapter-page';
import { getApiBaseUrl } from '@/components/docs/base-url';

export default function AdyenAdapterPage() {
  const base = getApiBaseUrl();
  const adapterBase = `${base}/adyen`;

  return (
    <AdapterPage
      title="Adyen"
      slug="adyen"
      blurb="Mirrors Adyen's /payments API: pspReference + resultCode + additionalData map on responses, and a notificationItems batch envelope for webhooks. refusalReason / refusalReasonCode populate on Refused results."
      baseUrl={adapterBase}
      operations={[
        { label: 'Create payment', method: 'POST', path: '/adyen/v68/payments' },
        { label: 'Capture', method: 'POST', path: '/adyen/v68/payments/:id/captures' },
        { label: 'Refund', method: 'POST', path: '/adyen/v68/payments/:id/refunds' },
      ]}
      opsNote="The handler's extractOperation() maps anything containing /payment to 'charge'. Adyen's real modifications API (captures, refunds) shares the same adapter logic — resultCode semantics don't change."
      requestFields={[
        {
          field: 'amount.value',
          description: 'Integer, minor units. Nested under amount (Adyen convention). Defaults to 5000.',
        },
        {
          field: 'amount.currency',
          description: 'ISO 4217 uppercase. Defaults to "USD".',
        },
        {
          field: 'additionalData',
          description:
            'Map of string→string echoed into the webhook\'s additionalData (merged with authCode placeholder).',
        },
        {
          field: 'merchantAccount / reference',
          description:
            'Adyen\'s usual routing fields. Read by the underlying extractMerchantOrderID helper but not required by this mock.',
        },
      ]}
      successBody={`{
  "pspReference":      "PSP1717000000000000000",
  "resultCode":        "Authorised",
  "merchantReference": "ref_1717000000000000000",
  "amount":            { "value": 5000, "currency": "USD" },
  "paymentMethod":     { "type": "scheme", "brand": "visa" },
  "additionalData":    { "acquirerCode": "TestAcquirer" }
}`}
      successNotes="resultCode is 'Authorised' on success, 'Received' on pending, 'Refused' on declines that still emit HTTP 200. Refused results include refusalReason + refusalReasonCode."
      errorExamples={[
        {
          label: 'bank_decline_hard (402)',
          status: 402,
          body: `{
  "status":       402,
  "errorCode":    "2",
  "message":      "Refused",
  "errorType":    "validation",
  "pspReference": "PSP1717000000000000000"
}`,
        },
        {
          label: 'pg_rate_limited (429)',
          status: 429,
          body: `{
  "status":       429,
  "errorCode":    "29",
  "message":      "Too many requests",
  "errorType":    "validation",
  "pspReference": "PSP1717000000000000000"
}`,
        },
        {
          label: 'pg_server_error (500)',
          status: 500,
          body: `{
  "status":       500,
  "errorCode":    "10",
  "message":      "Internal Error",
  "errorType":    "validation",
  "pspReference": "PSP1717000000000000000"
}`,
        },
      ]}
      webhookBody={`{
  "live": "false",
  "notificationItems": [
    {
      "NotificationRequestItem": {
        "additionalData":      { "authCode": "123456" },
        "amount":              { "value": 5000, "currency": "USD" },
        "eventCode":           "AUTHORISATION",
        "eventDate":           "2026-04-17T12:34:56.789Z",
        "merchantAccountCode": "TestMerchantAccount",
        "merchantReference":   "ref_1717000000000000000",
        "paymentMethod":       "scheme",
        "pspReference":        "<charge_id>",
        "success":             "true",
        "operations":          ["CANCEL", "CAPTURE", "REFUND"]
      }
    }
  ],
  "request_echo": { /* full request body */ }
}`}
      webhookNotes="eventCode is AUTHORISATION by default, CANCELLATION on success_then_reversed. success is the string 'true' / 'false' (Adyen quirk — not a bool). On failures the item carries a reason field."
      echoedFields={[
        {
          field: 'additionalData',
          description: 'Every string key/value is copied into the webhook additionalData map.',
        },
        {
          field: 'amount.value / amount.currency',
          description: 'Echoed on both response and webhook.',
        },
      ]}
      curlExample={`curl -X POST ${adapterBase}/v68/payments \\
  -H "Authorization: Bearer <api_key>" \\
  -H "Content-Type: application/json" \\
  -d '{
    "merchantAccount": "TEST",
    "amount":          { "value": 5000, "currency": "EUR" },
    "paymentMethod":   { "type": "scheme" },
    "reference":       "ref_123",
    "additionalData":  { "customData": "anything" }
  }'`}
    />
  );
}
