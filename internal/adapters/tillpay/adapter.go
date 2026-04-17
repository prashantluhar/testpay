// Package tillpay mocks the TillPayment acquirer API.
//
// Real-world shape (from TillPayment's merchant integration docs and the
// production reference DTO): HMAC-SHA512 signed requests, responses carry
// `{ success, uuid, purchaseId, returnType, paymentMethod }`, and webhooks
// use a separate envelope centred on `{ result, transactionStatus, uuid }`.
// The `amount` field on the webhook is a *string*, not a number — a
// TillPayment quirk we preserve so integration consumers see realistic data.
package tillpay

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/prashantluhar/testpay/internal/engine"
)

// signatureHeader matches the header TillPayment returns on every response.
// Our value is a static placeholder; the real gateway emits a base64-encoded
// HMAC-SHA512 over the response body.
const signatureHeader = "mock-sha512-signature"

type Adapter struct{}

func New() *Adapter             { return &Adapter{} }
func (a *Adapter) Name() string { return "tillpay" }

func (a *Adapter) BuildResponse(result *engine.Result, body []byte) (int, []byte, map[string]string) {
	headers := map[string]string{
		"Content-Type": "application/json",
		"X-Signature":  signatureHeader,
	}
	uuid := fmt.Sprintf("TP-%d", time.Now().UnixNano())
	purchaseID := fmt.Sprintf("PUR_%d", time.Now().UnixNano())

	if result.HTTPStatus >= 400 {
		errCode, errMsg := tillErrorCode(result.Mode), tillMessage(result.Mode)
		resp := transactionResponse{
			Success:       false,
			ErrorMessage:  errMsg,
			ErrorCode:     errCode,
			Message:       errMsg,
			UUID:          uuid,
			PurchaseID:    purchaseID,
			ReturnType:    "ERROR",
			PaymentMethod: paymentMethodCreditCard,
			Errors: []paymentError{{
				ErrorMessage:   errMsg,
				ErrorCode:      errCode,
				Message:        errMsg,
				Code:           strconv.Itoa(errCode),
				AdapterMessage: errMsg,
				AdapterCode:    paymentSystemError,
			}},
		}
		out, _ := json.Marshal(resp)
		return result.HTTPStatus, out, headers
	}

	resp := transactionResponse{
		Success:       true,
		UUID:          uuid,
		PurchaseID:    purchaseID,
		ReturnType:    returnTypeFinished,
		PaymentMethod: paymentMethodCreditCard,
		Message:       "OK",
	}
	// Redirect-family successful modes still need a redirect URL so the
	// merchant sees a 3DS / hosted-page flow kick off. On pending modes,
	// flip returnType to REDIRECT to cue the merchant to wait for webhook.
	if result.Mode == engine.ModeRedirectSuccess || result.IsPending {
		resp.ReturnType = "REDIRECT"
		resp.RedirectURL = fmt.Sprintf("https://test.tillpayments.com/redirect/%s", uuid)
	}
	out, _ := json.Marshal(resp)
	return 200, out, headers
}

func (a *Adapter) BuildWebhookPayload(result *engine.Result, chargeID string, amount int64, currency string, requestBody map[string]any) map[string]any {
	success := result.HTTPStatus < 400 && !result.IsPending
	txnStatus := tillTransactionStatus(result.Mode, result.HTTPStatus)

	// TillPayment sends amount as a *string* with 2-decimal precision on
	// webhooks. Our engine works in minor-units int64, so we convert.
	amtStr := fmt.Sprintf("%.2f", float64(amount)/100.0)

	notif := webhookNotification{
		Result:                tillResult(success),
		Success:               success,
		TransactionStatus:     txnStatus,
		UUID:                  chargeID,
		MerchantTransactionID: extractMerchantTxnID(requestBody, fmt.Sprintf("MTX_%d", time.Now().UnixNano())),
		PurchaseID:            fmt.Sprintf("PUR_%s", chargeID),
		TransactionType:       "DEBIT",
		PaymentMethod:         paymentMethodCreditCard,
		Amount:                amtStr,
		Currency:              currency,
		Customer:              extractCustomer(requestBody),
		ReturnData: returnData{
			Type:           "CREDITCARD",
			CardHolder:     "Test User",
			ExpiryMonth:    "12",
			ExpiryYear:     "2030",
			FirstSixDigits: "411111",
			LastFourDigits: "1111",
			Fingerprint:    "fp_mock_tillpay",
			BinBrand:       "VISA",
			BinCountry:     "US",
			ThreeDSecure:   "NOT_ENROLLED",
		},
	}
	if !success {
		msg := tillMessage(result.Mode)
		code := tillErrorCode(result.Mode)
		notif.Message = msg
		notif.Code = code
		notif.AdapterMessage = msg
		notif.AdapterCode = paymentSystemError
		notif.Errors = []paymentError{{
			ErrorMessage:   msg,
			ErrorCode:      code,
			Message:        msg,
			Code:           strconv.Itoa(code),
			AdapterMessage: msg,
			AdapterCode:    paymentSystemError,
		}}
	}

	// Round-trip to map[string]any so the dispatcher sees a normalised
	// payload, while keeping the typed DTO at the adapter boundary.
	raw, _ := json.Marshal(notif)
	var out map[string]any
	_ = json.Unmarshal(raw, &out)
	if requestBody != nil {
		out["request_echo"] = requestBody
	}
	return out
}

// tillResult maps the internal success flag onto TillPayment's short
// "OK"/"NOK" result vocabulary used at the top of each webhook payload.
func tillResult(success bool) string {
	if success {
		return "OK"
	}
	return "NOK"
}

// tillTransactionStatus maps engine FailureMode onto TillPayment's
// transactionStatus values: DEBIT/SUCCESS/FAILED/CHARGEBACK/REFUND.
// The reference DTO defines these five plus CHARGEBACK-REVERSED.
func tillTransactionStatus(mode engine.FailureMode, httpStatus int) string {
	if httpStatus >= 400 {
		return statusFailed
	}
	switch mode {
	case engine.ModeSuccessThenReversed:
		return statusChargebackReversed
	case engine.ModeDoubleCharge:
		// A second capture on the same uuid — TillPayment surfaces as DEBIT.
		return statusDebit
	case engine.ModePendingThenFailed, engine.ModePendingThenSuccess,
		engine.ModeFailedThenSuccess:
		// Pending states don't have a first-class TillPayment status; the
		// merchant integration treats DEBIT as the success-terminal marker.
		return statusDebit
	case engine.ModeSuccess, engine.ModeWebhookMissing, engine.ModeWebhookDelayed,
		engine.ModeWebhookDuplicate, engine.ModeWebhookOutOfOrder,
		engine.ModeWebhookMalformed, engine.ModeAmountMismatch,
		engine.ModePartialSuccess, engine.ModeRedirectSuccess:
		return statusSuccess
	default:
		return statusFailed
	}
}

// tillMessage maps engine modes to merchant-facing decline text. Mirrors the
// phrasing the real TillPayment gateway uses in its reference integration.
func tillMessage(mode engine.FailureMode) string {
	m := map[engine.FailureMode]string{
		engine.ModeBankDeclineHard:   "Transaction declined by issuer",
		engine.ModeBankDeclineSoft:   "Insufficient funds",
		engine.ModeBankInvalidCVV:    "Invalid CVV",
		engine.ModeBankDoNotHonour:   "Do not honour",
		engine.ModeBankTimeout:       "Issuer timeout",
		engine.ModeBankServerDown:    "Issuer unavailable",
		engine.ModePGTimeout:         "Gateway timeout",
		engine.ModePGServerError:     "Gateway internal error",
		engine.ModePGRateLimited:     "Rate limit exceeded",
		engine.ModePGMaintenance:     "Gateway under maintenance",
		engine.ModeNetworkError:      "Network error",
		engine.ModeRedirectAbandoned: "Customer abandoned redirect",
		engine.ModeRedirectTimeout:   "Redirect session expired",
		engine.ModeRedirectFailed:    "Redirect authentication failed",
	}
	if v, ok := m[mode]; ok {
		return v
	}
	return defaultUserMessage
}

// tillErrorCode maps engine modes to TillPayment's numeric error-code space.
// The reference DTO uses 5000 as the catch-all; we keep that and add mode-
// specific codes close to what the production gateway emits for each class.
func tillErrorCode(mode engine.FailureMode) int {
	m := map[engine.FailureMode]int{
		engine.ModeBankDeclineHard: 2001,
		engine.ModeBankDeclineSoft: 2002,
		engine.ModeBankInvalidCVV:  2003,
		engine.ModeBankDoNotHonour: 2004,
		engine.ModeBankTimeout:     2005,
		engine.ModeBankServerDown:  2006,
		engine.ModePGTimeout:       3001,
		engine.ModePGServerError:   3002,
		engine.ModePGRateLimited:   3003,
		engine.ModePGMaintenance:   3004,
		engine.ModeNetworkError:    3005,
	}
	if v, ok := m[mode]; ok {
		return v
	}
	// Matches DefaultErrorCode in the reference DTO — catch-all for unmapped
	// or composite failure modes.
	n, _ := strconv.Atoi(defaultErrorCode)
	return n
}

// extractCustomer pulls customer fields from the decoded request body. Falls
// back to a synthetic test customer so webhook consumers always see a
// well-formed payload, matching production behaviour.
func extractCustomer(req map[string]any) customer {
	c := customer{
		Identification: "test-customer",
		FirstName:      "Test",
		LastName:       "User",
		BillingCountry: "US",
		Email:          "test@example.com",
		EmailVerified:  true,
		IPAddress:      "127.0.0.1",
	}
	if req == nil {
		return c
	}
	raw, ok := req["customer"].(map[string]any)
	if !ok {
		return c
	}
	if v, ok := raw["identification"].(string); ok && v != "" {
		c.Identification = v
	}
	if v, ok := raw["firstName"].(string); ok && v != "" {
		c.FirstName = v
	}
	if v, ok := raw["lastName"].(string); ok && v != "" {
		c.LastName = v
	}
	if v, ok := raw["billingCountry"].(string); ok && v != "" {
		c.BillingCountry = v
	}
	if v, ok := raw["email"].(string); ok && v != "" {
		c.Email = v
	}
	if v, ok := raw["ipAddress"].(string); ok && v != "" {
		c.IPAddress = v
	}
	if v, ok := raw["emailVerified"].(bool); ok {
		c.EmailVerified = v
	}
	return c
}

// extractMerchantTxnID pulls the merchant-supplied correlation ID so the
// webhook echoes back the same value the merchant sent. Critical for
// reconciliation — without it the merchant can't pair webhook → order.
func extractMerchantTxnID(req map[string]any, fallback string) string {
	if req == nil {
		return fallback
	}
	if v, ok := req["merchantTransactionId"].(string); ok && v != "" {
		return v
	}
	return fallback
}
