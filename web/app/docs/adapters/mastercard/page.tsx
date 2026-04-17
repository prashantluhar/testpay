'use client';
import { AdapterPage } from '@/components/docs/adapter-page';
import { getApiBaseUrl } from '@/components/docs/base-url';

export default function MastercardAdapterPage() {
  const base = getApiBaseUrl();
  const adapterBase = `${base}/mastercard`;

  return (
    <AdapterPage
      title="Mastercard (MPGS)"
      slug="mastercard"
      blurb="Mirrors Mastercard Payment Gateway Services (MPGS) — result (SUCCESS | FAILURE | PENDING | ERROR) + response.gatewayCode taxonomy + order.status + transaction.id. Amounts are nested under `order` (float, not minor units)."
      baseUrl={adapterBase}
      operations={[
        {
          label: 'Create / authorise payment',
          method: 'POST',
          path: '/mastercard/api/rest/version/79/merchant/:id/order/:orderId',
        },
        { label: 'Capture', method: 'POST', path: '/mastercard/.../order/:id/transaction/:txId (capture body)' },
        { label: 'Refund', method: 'POST', path: '/mastercard/.../order/:id/transaction/:txId (refund body)' },
      ]}
      opsNote="MPGS uses a single REST endpoint with the operation encoded in the body (apiOperation: PAY / CAPTURE / REFUND / VOID). The adapter emits the same full paymentResponse shape regardless."
      requestFields={[
        {
          field: 'order.amount',
          description: 'Numeric amount (minor units). Nested — MPGS convention. Defaults to 5000.',
        },
        {
          field: 'order.currency',
          description: 'ISO 4217 uppercase. Defaults to "USD".',
        },
      ]}
      successBody={`{
  "result":          "SUCCESS",
  "merchant":        "TESTMERCHANT",
  "version":         "79",
  "timeOfRecord":    "2026-04-17T12:34:56Z",
  "timeOfLastUpdate":"2026-04-17T12:34:56Z",
  "gatewayEntryPoint":"https://test-gateway.mastercard.com/api/rest/version/79",
  "order": {
    "id":                    "ORD1717000000000000000",
    "amount":                5000,
    "currency":              "USD",
    "status":                "CAPTURED",
    "reference":             "ORD1717000000000000000",
    "totalAuthorizedAmount": 5000,
    "totalCapturedAmount":   5000,
    "authenticationStatus":  "AUTHENTICATION_NOT_IN_SCOPE"
  },
  "response": {
    "gatewayCode":           "APPROVED",
    "gatewayRecommendation": "PROCEED",
    "acquirerCode":          "00",
    "acquirerMessage":       "Approved",
    "cardSecurityCode":      { "gatewayCode": "MATCH", "acquirerCode": "M" }
  },
  "transaction": {
    "id":                "TXN1717000000000000000",
    "type":              "PAYMENT",
    "amount":            5000,
    "currency":          "USD",
    "authorizationCode": "831000",
    "source":            "INTERNET",
    "acquirer":          { "id": "TESTACQ", "merchantId": "TESTMERCHANT" }
  },
  "authorizationResponse": {
    "stan":           "123456",
    "responseCode":   "00",
    "processingCode": "000000"
  }
}`}
      successNotes="result is SUCCESS / FAILURE / PENDING / ERROR. gatewayCode follows MPGS's taxonomy (APPROVED / DECLINED / INVALID_CSC / TIMED_OUT / BLOCKED / EXPIRED_CARD / ...). gatewayRecommendation tells your caller what to do next: PROCEED / DO_NOT_PROCEED / RESUBMIT."
      errorExamples={[
        {
          label: 'bank_decline_hard (402) — same body on HTTP 200 Refused path',
          status: 402,
          body: `{
  "result": "FAILURE",
  ... /* full order + response block */
  "response": {
    "gatewayCode":           "DECLINED",
    "gatewayRecommendation": "DO_NOT_PROCEED",
    "acquirerCode":          "05",
    "acquirerMessage":       "Declined"
  },
  "error": {
    "cause":          "DECLINED",
    "explanation":    "Declined"
  }
}`,
        },
        {
          label: 'pg_server_error (500)',
          status: 500,
          body: `{
  "result": "ERROR",
  "response": {
    "gatewayCode":           "SYSTEM_ERROR",
    "gatewayRecommendation": "RESUBMIT",
    "acquirerCode":          "96",
    "acquirerMessage":       "System Error"
  },
  "error": {
    "cause":       "SERVER_FAILED",
    "explanation": "System Error"
  }
}`,
        },
        {
          label: 'pg_rate_limited (429)',
          status: 429,
          body: `{
  "result": "FAILURE",
  "response": {
    "gatewayCode":    "BLOCKED",
    "acquirerCode":   "62",
    "acquirerMessage":"Blocked by Gateway"
  },
  "error": { "cause": "REQUEST_REJECTED", "explanation": "Blocked by Gateway" }
}`,
        },
      ]}
      webhookBody={`{
  "notificationId":     "NOTIF1717000000000000000",
  "notificationType":   "ORDER",
  "timeOfNotification": "2026-04-17T12:34:56Z",
  "result":             "SUCCESS",
  "merchant":           "TESTMERCHANT",
  "version":            "79",
  "order": {
    "id":                    "<charge_id>",
    "amount":                5000,
    "currency":              "USD",
    "status":                "CAPTURED",
    "reference":             "<charge_id>",
    "totalAuthorizedAmount": 5000,
    "totalCapturedAmount":   5000
  },
  "response": {
    "gatewayCode":    "APPROVED",
    "gatewayRecommendation": "PROCEED",
    "acquirerCode":   "00",
    "acquirerMessage":"Approved"
  },
  "transaction": {
    "id":                "TXN...",
    "type":              "PAYMENT",
    "amount":            5000,
    "currency":          "USD",
    "authorizationCode": "831000",
    "source":            "INTERNET",
    "acquirer":          { "merchantId": "TESTMERCHANT" }
  },
  "request_echo": { /* full request body */ }
}`}
      webhookNotes="notificationType is ORDER for most events, TRANSACTION for reversal / refund. transaction.type is PAYMENT normally, REFUND for success_then_reversed."
      echoedFields={[
        {
          field: 'order.reference',
          description: 'Populated from the synthetic order ID — echoed on response + webhook.',
        },
        {
          field: 'Whole request body',
          description: 'Attached to the webhook as request_echo.',
        },
      ]}
      curlExample={`curl -X POST ${adapterBase}/api/rest/version/79/merchant/TESTMERCHANT/order/ORD123 \\
  -H "Authorization: Bearer <api_key>" \\
  -H "Content-Type: application/json" \\
  -d '{
    "apiOperation": "PAY",
    "order":        { "amount": 5000, "currency": "USD" }
  }'`}
    />
  );
}
