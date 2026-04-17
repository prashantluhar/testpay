// DTO shapes for Komoju's (Japan) payment API.
// Derived from the production reference library (test-dto-/payments.go,
// structs: KomojuInitResPayload, KomojuWebhookPayload, KomojuErrorPayload,
// KomojuCardTokenResponse, KomojuRefundRequsetPayload).
//
// Komoju's API is resource-oriented — most responses carry a `resource`
// discriminator ("payment", "event", "refund"). Status lifecycle is:
//   authorized → captured → refunded
// with failed/cancelled as terminal sad paths. Amounts are plain int (JPY
// has no minor units anyway).
package komoju

import "time"

// paymentResource mirrors the subset of KomojuInitResPayload this mock emits.
// The full struct has ~30 fields; we stick to what drives the happy + sad
// paths and omit the rich payment_details polymorphism.
type paymentResource struct {
	ID               string          `json:"id"`
	Resource         string          `json:"resource"` // always "payment" on this shape
	Status           string          `json:"status"`
	Amount           int             `json:"amount"`
	Tax              int             `json:"tax"`
	Total            int             `json:"total"`
	Currency         string          `json:"currency"`
	Description      string          `json:"description,omitempty"`
	ExternalOrderNum string          `json:"external_order_num,omitempty"`
	PaymentMethodFee int             `json:"payment_method_fee"`
	PaymentDetails   paymentDetails  `json:"payment_details"`
	CapturedAt       *time.Time      `json:"captured_at"`
	CreatedAt        time.Time       `json:"created_at"`
	AmountRefunded   int             `json:"amount_refunded"`
	Locale           string          `json:"locale,omitempty"`
	Metadata         map[string]any  `json:"metadata"`
	Refunds          []refundSummary `json:"refunds"`
}

// paymentDetails is the minimal shape of KomojuInitResPayload.PaymentDetails
// this mock emits — type + redirect_url is enough to drive a simulated
// redirect flow.
type paymentDetails struct {
	Type        string `json:"type"`
	Email       string `json:"email,omitempty"`
	RedirectURL string `json:"redirect_url,omitempty"`
}

// refundSummary mirrors the nested refund object in KomojuInitResPayload.
type refundSummary struct {
	ID          string    `json:"id"`
	Resource    string    `json:"resource"` // always "refund"
	Amount      int       `json:"amount"`
	Currency    string    `json:"currency"`
	Payment     string    `json:"payment"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	Chargeback  bool      `json:"chargeback"`
}

// errorResponse mirrors KomojuErrorPayload — Komoju wraps the error body in
// an `error` object with code + message + param, unlike the resource-at-root
// success envelope.
type errorResponse struct {
	Error errorBody `json:"error"`
}

type errorBody struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Param   string         `json:"param,omitempty"`
	Details map[string]any `json:"details,omitempty"`
}

// webhookEnvelope mirrors KomojuWebhookPayload — the shape Komoju posts to
// the merchant's callback URL. `Type` is the event code ("payment.captured",
// "payment.refunded", etc); `Data` is the full paymentResource snapshot.
type webhookEnvelope struct {
	ID        string          `json:"id"`
	Type      string          `json:"type"`
	Resource  string          `json:"resource"` // always "event"
	Data      paymentResource `json:"data"`
	CreatedAt time.Time       `json:"created_at"`
	Reason    string          `json:"reason,omitempty"`
	Details   *webhookDetails `json:"details,omitempty"`
}

type webhookDetails struct {
	ErrorCode string `json:"error_code"`
}
