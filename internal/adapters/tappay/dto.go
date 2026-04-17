// DTO shapes for TapPay's payment API surface.
// Derived from the production reference library (test-dto-/payments.go,
// structs: AppotapayPaymentResponse, AppotapayWebhookOrderStatusResponse,
// AppotapayRefundResponse, AppotapayErrorResponse, ...).
//
// The reference DTOs are named "Appotapay*" — the wire format the mock
// emits matches those shapes 1:1. The Name() on the adapter remains
// "tappay" because that's what the repo registers the adapter under; only
// the payload field names / error vocabulary come from Appotapay.
//
// Wire conventions: errorCode is a numeric (0 on success, non-zero on
// failure) and signature is a fixed mock placeholder — real Appotapay uses
// HMAC-SHA256(partnerCode + apiKey + orderId + ...). Consumers verifying
// signatures should detect the mock prefix and skip.
package tappay

// paymentResponse mirrors AppotapayPaymentResponse — the init / charge
// response. errorCode == 0 means success; any non-zero value is a failure
// and callers MUST check this before trusting paymentUrl.
type paymentResponse struct {
	ErrorCode  float64 `json:"errorCode"`
	Message    string  `json:"message"`
	OrderID    string  `json:"orderId"`
	Amount     float64 `json:"amount"`
	PaymentURL string  `json:"paymentUrl,omitempty"`
	Signature  string  `json:"signature"`
}

// errorResponse mirrors AppotapayErrorResponse — minimal envelope the
// gateway returns for validation / auth errors before the transaction even
// reaches the acquirer.
type errorResponse struct {
	ErrorCode int    `json:"errorCode"`
	Message   string `json:"message"`
	Signature string `json:"signature"`
	OrderID   string `json:"orderId,omitempty"`
}

// webhookOrderStatusResponse mirrors AppotapayWebhookOrderStatusResponse —
// the POST body Appotapay delivers to the merchant's notifyUrl once a
// transaction reaches a terminal state. paymentMethod / paymentType /
// bankCode are echoed so merchants can route by channel.
type webhookOrderStatusResponse struct {
	ErrorCode        int    `json:"errorCode"`
	Message          string `json:"message"`
	PartnerCode      string `json:"partnerCode"`
	APIKey           string `json:"apiKey"`
	Amount           int    `json:"amount"`
	Currency         string `json:"currency"`
	OrderID          string `json:"orderId"`
	BankCode         string `json:"bankCode"`
	PaymentMethod    string `json:"paymentMethod"`
	PaymentType      string `json:"paymentType"`
	AppotapayTransID string `json:"appotapayTransId"`
	TransactionTs    int64  `json:"transactionTs"`
	ExtraData        string `json:"extraData,omitempty"`
	Signature        string `json:"signature"`
}

// refundResponse mirrors AppotapayRefundResponse — included for completeness
// (the mock doesn't yet expose a refund endpoint) so the DTO surface
// parallels the reference library.
type refundResponse struct {
	ErrorCode int    `json:"errorCode"`
	Message   string `json:"message"`
	Data      struct {
		AppotapayTransID string `json:"appotapayTransId"`
		RefundID         string `json:"refundId"`
		RefundOriginalID string `json:"refundOriginalId"`
		Amount           int    `json:"amount"`
		Reason           string `json:"reason"`
		Status           string `json:"status"`
		TransactionTs    int64  `json:"transactionTs"`
	} `json:"data"`
	Signature string `json:"signature,omitempty"`
}
