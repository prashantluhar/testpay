// Package ecpay mocks ECPay (綠界 / Green World) — Taiwan's dominant
// acquiring gateway. Real-world wire shape: AioCheckOut v5 for order
// creation, ReturnURL form-POST for async settlement notifications.
// See dto.go for the sourcing notes on where each field originates.
package ecpay

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/prashantluhar/testpay/internal/engine"
)

const (
	// merchantID is the test-mode stage MerchantID ECPay ships with its
	// developer stage SDK — safe to hard-code for a mock.
	merchantID = "3002607"
	// storeID is our synthetic store identifier; merchants partition by it
	// in production but a mock only needs one.
	storeID = "TESTSTORE"
	// checkMacPlaceholder replaces the real MD5 signature ECPay emits.
	// A production consumer that *verifies* CheckMacValue will reject this,
	// which is correct behaviour for a mock — it forces the consumer to
	// opt-in to test mode rather than silently accepting fake signatures.
	checkMacPlaceholder = "MOCK_CHECK_MAC_VALUE"
)

type Adapter struct{}

func New() *Adapter             { return &Adapter{} }
func (a *Adapter) Name() string { return "ecpay" }

func (a *Adapter) BuildResponse(result *engine.Result, body []byte) (int, []byte, map[string]string) {
	headers := map[string]string{"Content-Type": "application/json"}
	tradeNo := fmt.Sprintf("EC%d", time.Now().UnixNano())
	merchantTradeNo := extractMerchantTradeNo(body, fmt.Sprintf("MTN%d", time.Now().UnixNano()))

	if result.HTTPStatus >= 400 {
		errResp := errorResponse{
			MerchantID:      merchantID,
			MerchantTradeNo: merchantTradeNo,
			RtnCode:         ecpayRtnCode(result.Mode),
			RtnMsg:          ecpayMessage(result.Mode),
			ErrorType:       "validation",
			CheckMacValue:   checkMacPlaceholder,
		}
		out, _ := json.Marshal(errResp)
		return result.HTTPStatus, out, headers
	}

	amount, _ := extractAmount(body, 100)
	resp := aioResponse{
		MerchantID:           merchantID,
		MerchantTradeNo:      merchantTradeNo,
		TradeNo:              tradeNo,
		TradeAmt:             amount,
		RtnCode:              ecpayRtnCode(result.Mode),
		RtnMsg:               ecpayMessage(result.Mode),
		PaymentType:          ecpayPaymentType(body),
		PaymentTypeChargeFee: "0",
		TradeDate:            time.Now().UTC().Format("2006/01/02 15:04:05"),
		CheckMacValue:        checkMacPlaceholder,
	}
	out, _ := json.Marshal(resp)
	return 200, out, headers
}

func (a *Adapter) BuildWebhookPayload(result *engine.Result, chargeID string, amount int64, currency string, requestBody map[string]any) map[string]any {
	// ECPay doesn't carry a `currency` field — TWD is implicit. We accept
	// it in the adapter signature but don't emit it.
	_ = currency

	rtnCode := ecpayRtnCode(result.Mode)
	if result.HTTPStatus >= 400 {
		// Override: an HTTP-failure result is always a non-success webhook
		// regardless of the mode's intrinsic classification. Matters for
		// modes like ModePGServerError where the mode's "intent" is ambiguous
		// but the HTTP layer already decided the fate.
		rtnCode = rtnCodeFailed
	}

	now := time.Now().UTC()
	notif := webhookCallback{
		MerchantID:           merchantID,
		MerchantTradeNo:      extractMerchantTradeNoFromMap(requestBody, fmt.Sprintf("MTN_%d", time.Now().UnixNano())),
		StoreID:              storeID,
		RtnCode:              rtnCode,
		RtnMsg:               ecpayMessage(result.Mode),
		TradeNo:              chargeID,
		TradeAmt:             amount,
		PaymentDate:          now.Format("2006/01/02 15:04:05"),
		PaymentType:          ecpayPaymentTypeFromMap(requestBody),
		PaymentTypeChargeFee: "0",
		TradeDate:            now.Format("2006/01/02 15:04:05"),
		// SimulatePaid is "1" in test mode so the merchant SDK knows not to
		// fulfil the order against real inventory.
		SimulatePaid:  "1",
		CheckMacValue: checkMacPlaceholder,
	}
	// Echo merchant custom fields verbatim — these are ECPay's correlation
	// channel for merchant-side state (order IDs, user IDs, etc.).
	if requestBody != nil {
		if v, ok := requestBody["CustomField1"].(string); ok {
			notif.CustomField1 = v
		}
		if v, ok := requestBody["CustomField2"].(string); ok {
			notif.CustomField2 = v
		}
		if v, ok := requestBody["CustomField3"].(string); ok {
			notif.CustomField3 = v
		}
		if v, ok := requestBody["CustomField4"].(string); ok {
			notif.CustomField4 = v
		}
	}

	raw, _ := json.Marshal(notif)
	var out map[string]any
	_ = json.Unmarshal(raw, &out)
	if requestBody != nil {
		out["request_echo"] = requestBody
	}
	return out
}

// ecpayRtnCode maps engine FailureMode onto ECPay's numeric RtnCode
// vocabulary: 1 is success, everything else is a failure. Pending (ATM/CVS
// await payment) uses 10100073; other declines use 10100248 as the generic
// acquirer-decline code.
func ecpayRtnCode(mode engine.FailureMode) int {
	switch mode {
	case engine.ModeSuccess, engine.ModeWebhookMissing, engine.ModeWebhookDelayed,
		engine.ModeWebhookDuplicate, engine.ModeWebhookOutOfOrder,
		engine.ModeWebhookMalformed, engine.ModeAmountMismatch,
		engine.ModePartialSuccess, engine.ModeDoubleCharge,
		engine.ModeRedirectSuccess:
		return rtnCodeSuccess
	case engine.ModePendingThenFailed, engine.ModePendingThenSuccess,
		engine.ModeFailedThenSuccess, engine.ModeSuccessThenReversed:
		return rtnCodePending
	default:
		return rtnCodeFailed
	}
}

// ecpayMessage returns merchant-facing RtnMsg text matching ECPay's docs.
// Success is always "交易成功" in production; we emit English for test
// readability but keep the canonical success message so string-matching
// consumers still work.
func ecpayMessage(mode engine.FailureMode) string {
	switch mode {
	case engine.ModeSuccess, engine.ModeRedirectSuccess,
		engine.ModeWebhookMissing, engine.ModeWebhookDelayed,
		engine.ModeWebhookDuplicate, engine.ModeWebhookOutOfOrder,
		engine.ModeWebhookMalformed, engine.ModeAmountMismatch,
		engine.ModePartialSuccess, engine.ModeDoubleCharge:
		return "交易成功"
	case engine.ModePendingThenFailed, engine.ModePendingThenSuccess,
		engine.ModeFailedThenSuccess:
		return "Get CVS Code Succeeded"
	case engine.ModeSuccessThenReversed:
		return "Refund Succeeded"
	}
	m := map[engine.FailureMode]string{
		engine.ModeBankDeclineHard:   "Issuer declined transaction",
		engine.ModeBankDeclineSoft:   "Insufficient funds",
		engine.ModeBankInvalidCVV:    "Invalid security code",
		engine.ModeBankDoNotHonour:   "Do not honour",
		engine.ModeBankTimeout:       "Issuer timeout",
		engine.ModeBankServerDown:    "Issuer unavailable",
		engine.ModePGTimeout:         "Gateway timeout",
		engine.ModePGServerError:     "Gateway error",
		engine.ModePGRateLimited:     "Too many requests",
		engine.ModePGMaintenance:     "Service under maintenance",
		engine.ModeNetworkError:      "Network error",
		engine.ModeRedirectAbandoned: "User cancelled payment",
		engine.ModeRedirectTimeout:   "Payment session expired",
		engine.ModeRedirectFailed:    "3DS authentication failed",
	}
	if v, ok := m[mode]; ok {
		return v
	}
	return "Transaction failed"
}

// ecpayPaymentType picks the PaymentType echo string from the request body.
// Maps from our internal ECPAY_* vocabulary (matching the production DTO's
// PaymentMethodToECPay map) to ECPay's wire format (Credit_CreditCard,
// ATM_TAISHIN, CVS_CVS, BARCODE_BARCODE).
func ecpayPaymentType(body []byte) string {
	if len(body) == 0 {
		return paymentCredit
	}
	var m map[string]any
	if json.Unmarshal(body, &m) != nil {
		return paymentCredit
	}
	return ecpayPaymentTypeFromMap(m)
}

func ecpayPaymentTypeFromMap(m map[string]any) string {
	if m == nil {
		return paymentCredit
	}
	pm, _ := m["PaymentMethod"].(string)
	if pm == "" {
		pm, _ = m["paymentMethod"].(string)
	}
	switch pm {
	case methodATM:
		return paymentATM
	case methodCVS:
		return paymentCVS
	case methodBarcode:
		return paymentBarcode
	case methodCredit, "":
		return paymentCredit
	}
	return paymentCredit
}

// extractAmount reads TradeAmt from the request body. ECPay uses integer
// TWD (no decimals); default stage-tier test amount is 100.
func extractAmount(body []byte, def int64) (int64, bool) {
	if len(body) == 0 {
		return def, false
	}
	var m map[string]any
	if json.Unmarshal(body, &m) != nil {
		return def, false
	}
	if v, ok := m["TradeAmt"].(float64); ok {
		return int64(v), true
	}
	if v, ok := m["tradeAmt"].(float64); ok {
		return int64(v), true
	}
	if v, ok := m["amount"].(float64); ok {
		return int64(v), true
	}
	return def, false
}

func extractMerchantTradeNo(body []byte, fallback string) string {
	if len(body) == 0 {
		return fallback
	}
	var m map[string]any
	if json.Unmarshal(body, &m) != nil {
		return fallback
	}
	return extractMerchantTradeNoFromMap(m, fallback)
}

func extractMerchantTradeNoFromMap(m map[string]any, fallback string) string {
	if m == nil {
		return fallback
	}
	if v, ok := m["MerchantTradeNo"].(string); ok && v != "" {
		return v
	}
	if v, ok := m["merchantTradeNo"].(string); ok && v != "" {
		return v
	}
	return fallback
}
