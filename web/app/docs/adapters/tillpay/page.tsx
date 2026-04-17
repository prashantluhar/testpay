'use client';
import { AdapterPage } from '@/components/docs/adapter-page';
import { getApiBaseUrl } from '@/components/docs/base-url';

export default function TillpayAdapterPage() {
  const base = getApiBaseUrl();
  const adapterBase = `${base}/tillpay`;

  return (
    <AdapterPage
      title="TillPayment"
      slug="tillpay"
      blurb="Card-acquiring style — `{ success, uuid, purchaseId, returnType, paymentMethod }` response shape, webhook centered on `{ result, transactionStatus, uuid }`. Note: webhook amount is a STRING (TillPayment quirk) with 2-decimal precision."
      baseUrl={adapterBase}
      operations={[
        { label: 'Debit (charge)', method: 'POST', path: '/tillpay/v1/transaction/debit' },
        { label: 'Capture', method: 'POST', path: '/tillpay/v1/transaction/capture' },
        { label: 'Refund', method: 'POST', path: '/tillpay/v1/transaction/refund' },
      ]}
      opsNote="Any POST under /tillpay/* is treated as a transaction. The X-Signature response header carries a static placeholder; real TillPayment signs with HMAC-SHA512."
      requestFields={[
        {
          field: 'amount',
          description: 'Integer minor units on request. Defaults to 5000. Echoed as STRING on webhook.',
        },
        {
          field: 'currency',
          description: 'ISO 4217 uppercase. Defaults to "USD".',
        },
        {
          field: 'merchantTransactionId',
          description:
            "Merchant's correlation ID. Echoed on webhook merchantTransactionId — critical for reconciliation.",
        },
        {
          field: 'customer',
          description:
            'Nested object — identification, firstName, lastName, billingCountry, email, emailVerified, ipAddress. Each is read and echoed on the webhook customer block.',
        },
      ]}
      successBody={`{
  "success":       true,
  "uuid":          "TP-1717000000000000000",
  "purchaseId":    "PUR_1717000000000000000",
  "returnType":    "FINISHED",
  "paymentMethod": "CREDITCARD",
  "message":       "OK"
}`}
      successNotes="On redirect_success or any pending mode, returnType flips to 'REDIRECT' and redirectUrl is populated. Response carries X-Signature header."
      errorExamples={[
        {
          label: 'bank_decline_hard (402)',
          status: 402,
          body: `{
  "success":       false,
  "errorMessage":  "Transaction declined by issuer",
  "errorCode":     2001,
  "message":       "Transaction declined by issuer",
  "uuid":          "TP-1717000000000000000",
  "purchaseId":    "PUR_1717000000000000000",
  "returnType":    "ERROR",
  "paymentMethod": "CREDITCARD",
  "errors": [
    {
      "errorMessage":   "Transaction declined by issuer",
      "errorCode":      2001,
      "message":        "Transaction declined by issuer",
      "code":           "2001",
      "adapterMessage": "Transaction declined by issuer",
      "adapterCode":    "PAYMENT_SYSTEM_ERROR"
    }
  ]
}`,
        },
        {
          label: 'pg_rate_limited (429)',
          status: 429,
          body: `{
  "success":    false,
  "errorMessage":"Rate limit exceeded",
  "errorCode":  3003,
  "returnType": "ERROR",
  "errors": [ { "errorCode": 3003, "adapterCode": "PAYMENT_SYSTEM_ERROR" } ]
}`,
        },
      ]}
      webhookBody={`{
  "result":                "OK",
  "success":               true,
  "transactionStatus":     "SUCCESS",
  "uuid":                  "<charge_id>",
  "merchantTransactionId": "MTX_1717000000000000000",
  "purchaseId":            "PUR_<charge_id>",
  "transactionType":       "DEBIT",
  "paymentMethod":         "CREDITCARD",
  "amount":                "50.00",
  "currency":              "USD",
  "customer": {
    "identification": "test-customer",
    "firstName":      "Test",
    "lastName":       "User",
    "billingCountry": "US",
    "email":          "test@example.com",
    "emailVerified":  true,
    "ipAddress":      "127.0.0.1"
  },
  "returnData": {
    "_TYPE":          "CREDITCARD",
    "cardHolder":     "Test User",
    "expiryMonth":    "12",
    "expiryYear":     "2030",
    "firstSixDigits": "411111",
    "lastFourDigits": "1111",
    "fingerprint":    "fp_mock_tillpay",
    "binBrand":       "VISA",
    "binCountry":     "US",
    "threeDSecure":   "NOT_ENROLLED"
  },
  "request_echo": { /* full request body */ }
}`}
      webhookNotes={
        'transactionStatus drives the state machine: DEBIT / SUCCESS / FAILED / CHARGEBACK / CHARGEBACK-REVERSED. amount is a STRING with 2-decimal precision (TillPayment quirk). On failure, errors[] is populated and top-level message/code mirror the first error.'
      }
      echoedFields={[
        {
          field: 'merchantTransactionId',
          description:
            'Passed through from the request. If you omit it a synthetic MTX_… ID is generated.',
        },
        {
          field: 'customer.*',
          description:
            'identification, firstName, lastName, billingCountry, email, emailVerified, ipAddress — each read from the request body and echoed verbatim.',
        },
      ]}
      curlExample={`curl -X POST ${adapterBase}/v1/transaction/debit \\
  -H "Authorization: Bearer <api_key>" \\
  -H "Content-Type: application/json" \\
  -d '{
    "amount":                5000,
    "currency":              "USD",
    "merchantTransactionId": "MTX_123",
    "customer": {
      "identification": "cust_456",
      "email":          "buyer@example.com"
    }
  }'`}
    />
  );
}
