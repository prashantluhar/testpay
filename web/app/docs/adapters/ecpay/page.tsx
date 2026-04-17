'use client';
import { AdapterPage } from '@/components/docs/adapter-page';
import { getApiBaseUrl } from '@/components/docs/base-url';

export default function EcpayAdapterPage() {
  const base = getApiBaseUrl();
  const adapterBase = `${base}/ecpay`;

  return (
    <AdapterPage
      title="ECPay (綠界)"
      slug="ecpay"
      blurb="Taiwan's dominant acquirer. Real wire shape: AioCheckOut v5 for order creation + ReturnURL form-POST webhook. TWD is whole-dollar (no minor units). The mock emits JSON on both sides (real ECPay uses x-www-form-urlencoded on webhook) — field names preserved exactly."
      baseUrl={adapterBase}
      operations={[
        {
          label: 'Create order (AioCheckOut v5)',
          method: 'POST',
          path: '/ecpay/Cashier/AioCheckOut/V5',
        },
        { label: 'Query order status', method: 'POST', path: '/ecpay/Cashier/QueryTradeInfo/V5' },
      ]}
      requestFields={[
        {
          field: 'MerchantTradeNo',
          description:
            "Merchant's correlation ID. Echoed on response + webhook. Also accepted as lowercase merchantTradeNo.",
        },
        {
          field: 'TradeAmt',
          description:
            'Integer TWD (no decimals). Defaults to 100. Also accepted as lowercase tradeAmt or "amount".',
        },
        {
          field: 'PaymentMethod',
          description:
            'Tender channel. Values: ECPAY_CREDIT_CARD | ECPAY_ATM_CARD | ECPAY_CVS | ECPAY_BARCODE. Mapped to PaymentType on the response/webhook.',
        },
        {
          field: 'CustomField1..CustomField4',
          description:
            'ECPay\'s free-text correlation channel. Each is echoed verbatim on the webhook.',
        },
      ]}
      successBody={`{
  "MerchantID":           "3002607",
  "MerchantTradeNo":      "ORDER_123",
  "TradeNo":              "EC1717000000000000000",
  "TradeAmt":             100,
  "RtnCode":              1,
  "RtnMsg":               "交易成功",
  "PaymentType":          "Credit_CreditCard",
  "PaymentTypeChargeFee": "0",
  "TradeDate":            "2026/04/17 12:34:56",
  "CheckMacValue":        "MOCK_CHECK_MAC_VALUE"
}`}
      successNotes="RtnCode == 1 is the universal success marker. 10100073 = pending (async ATM/CVS), 10100248 = generic decline. CheckMacValue is a placeholder — production code signing with MD5 will reject it, which is the desired opt-in signal."
      errorExamples={[
        {
          label: 'bank_decline_hard (402)',
          status: 402,
          body: `{
  "MerchantID":      "3002607",
  "MerchantTradeNo": "ORDER_123",
  "RtnCode":         10100248,
  "RtnMsg":          "Issuer declined transaction",
  "ErrorType":       "validation",
  "CheckMacValue":   "MOCK_CHECK_MAC_VALUE"
}`,
        },
        {
          label: 'pg_rate_limited (429)',
          status: 429,
          body: `{
  "MerchantID":      "3002607",
  "MerchantTradeNo": "ORDER_123",
  "RtnCode":         10100248,
  "RtnMsg":          "Too many requests",
  "ErrorType":       "validation",
  "CheckMacValue":   "MOCK_CHECK_MAC_VALUE"
}`,
        },
      ]}
      webhookBody={`{
  "MerchantID":           "3002607",
  "MerchantTradeNo":      "ORDER_123",
  "StoreID":              "TESTSTORE",
  "RtnCode":              1,
  "RtnMsg":               "交易成功",
  "TradeNo":              "<charge_id>",
  "TradeAmt":             100,
  "PaymentDate":          "2026/04/17 12:34:56",
  "PaymentType":          "Credit_CreditCard",
  "PaymentTypeChargeFee": "0",
  "TradeDate":            "2026/04/17 12:34:56",
  "SimulatePaid":         "1",
  "CustomField1":         "anything-you-sent",
  "CheckMacValue":        "MOCK_CHECK_MAC_VALUE",
  "request_echo":         { /* full request body */ }
}`}
      webhookNotes={
        'SimulatePaid is always "1" in the mock — your production webhook consumer MUST refuse to fulfil orders where SimulatePaid=1 to prevent a test request from committing real inventory.'
      }
      echoedFields={[
        {
          field: 'MerchantTradeNo',
          description:
            "Passed through on the webhook. If omitted, a synthetic MTN_… ID is generated.",
        },
        {
          field: 'CustomField1..4',
          description: 'Each echoed verbatim on the webhook.',
        },
      ]}
      curlExample={`curl -X POST ${adapterBase}/Cashier/AioCheckOut/V5 \\
  -H "Authorization: Bearer <api_key>" \\
  -H "Content-Type: application/json" \\
  -d '{
    "MerchantTradeNo": "ORDER_123",
    "TradeAmt":        100,
    "PaymentMethod":   "ECPAY_CREDIT_CARD",
    "CustomField1":    "user-42"
  }'`}
    />
  );
}
