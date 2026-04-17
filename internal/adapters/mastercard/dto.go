// DTO shapes for Mastercard Payment Gateway Services (MPGS).
// Derived from the production reference library (test-dto-/mastercard.go).
// Kept scoped to the subset this mock actually emits on
// BuildResponse / BuildWebhookPayload — extend when a new failure mode
// needs additional surface area.
//
// Reference types used (of the 551-line source):
//   - MasterCardPaymentResponse    -> paymentResponse
//   - MasterCardOrderExtended      -> orderResponse
//   - MasterCardResponseExtended   -> responseBlock
//   - MasterCardCardSecurityCode   -> cardSecurityCode
//   - MasterCardTransactionExtended-> transactionResponse
//   - MasterCardAcquirerExtended   -> acquirerBlock
//   - MasterCardAuthorizationResponse -> authorizationResponse
//   - MasterCardError              -> errorBlock
// Status constants mirror MasterCardSuccessStatus / FailedStatus /
// PendingStatus / ErrorStatus / UnknownStatus.
package mastercard

// MPGS top-level status values. These are the `result` field values
// MPGS uses across every transaction-shaped response.
const (
	resultSuccess = "SUCCESS"
	resultFailure = "FAILURE"
	resultPending = "PENDING"
	resultError   = "ERROR"
	resultUnknown = "UNKNOWN"
)

// paymentResponse mirrors MasterCardPaymentResponse — the response MPGS
// returns from the Authorize / Pay / Capture transaction endpoints.
// Successful path sets Result=SUCCESS and Response.GatewayCode=APPROVED;
// declines set Result=FAILURE with a specific gatewayCode (e.g. DECLINED,
// INVALID_CSC, TIMED_OUT, BLOCKED, EXPIRED_CARD); pending sets
// Result=PENDING and a PENDING gatewayCode.
type paymentResponse struct {
	Result           string                `json:"result"`
	Merchant         string                `json:"merchant,omitempty"`
	Version          string                `json:"version,omitempty"`
	TimeOfRecord     string                `json:"timeOfRecord,omitempty"`
	TimeOfLastUpdate string                `json:"timeOfLastUpdate,omitempty"`
	GatewayEntryPoint string               `json:"gatewayEntryPoint,omitempty"`
	Order            orderResponse         `json:"order"`
	Response         responseBlock         `json:"response"`
	Transaction      transactionResponse   `json:"transaction"`
	AuthorizationResponse authorizationResponse `json:"authorizationResponse,omitempty"`
	Error            *errorBlock           `json:"error,omitempty"`
}

// orderResponse mirrors MasterCardOrderExtended. MPGS tracks order-level
// totals (authorized / captured / refunded) here; the mock populates
// just what's needed to round-trip amount + currency + status.
type orderResponse struct {
	ID                    string  `json:"id"`
	Amount                float64 `json:"amount"`
	Currency              string  `json:"currency"`
	Status                string  `json:"status"`
	Reference             string  `json:"reference,omitempty"`
	CreationTime          string  `json:"creationTime,omitempty"`
	LastUpdatedTime       string  `json:"lastUpdatedTime,omitempty"`
	TotalAuthorizedAmount float64 `json:"totalAuthorizedAmount"`
	TotalCapturedAmount   float64 `json:"totalCapturedAmount"`
	TotalRefundedAmount   float64 `json:"totalRefundedAmount"`
	AuthenticationStatus  string  `json:"authenticationStatus,omitempty"`
}

// responseBlock mirrors MasterCardResponseExtended. GatewayCode is the
// MPGS taxonomy merchants pattern-match on (APPROVED, DECLINED,
// INVALID_CSC, TIMED_OUT, BLOCKED, EXPIRED_CARD, SYSTEM_ERROR, ...).
// GatewayRecommendation tells the merchant what to do next
// (PROCEED / DO_NOT_PROCEED / RESUBMIT).
type responseBlock struct {
	GatewayCode           string            `json:"gatewayCode"`
	GatewayRecommendation string            `json:"gatewayRecommendation,omitempty"`
	AcquirerCode          string            `json:"acquirerCode,omitempty"`
	AcquirerMessage       string            `json:"acquirerMessage,omitempty"`
	CardSecurityCode      *cardSecurityCode `json:"cardSecurityCode,omitempty"`
}

// cardSecurityCode mirrors MasterCardCardSecurityCode — the nested
// CVV-check outcome MPGS returns on every card transaction.
type cardSecurityCode struct {
	GatewayCode  string `json:"gatewayCode"`
	AcquirerCode string `json:"acquirerCode"`
}

// transactionResponse mirrors MasterCardTransactionExtended. One order
// can have many transactions (auth, capture, refund, void); the mock
// only emits a single payment per response.
type transactionResponse struct {
	ID                   string         `json:"id"`
	Type                 string         `json:"type"`
	Amount               float64        `json:"amount"`
	Currency             string         `json:"currency"`
	AuthorizationCode    string         `json:"authorizationCode,omitempty"`
	AuthenticationStatus string         `json:"authenticationStatus,omitempty"`
	Source               string         `json:"source,omitempty"`
	Acquirer             acquirerBlock  `json:"acquirer"`
}

// acquirerBlock mirrors MasterCardAcquirerExtended.
type acquirerBlock struct {
	ID            string `json:"id,omitempty"`
	MerchantID    string `json:"merchantId"`
	TransactionID string `json:"transactionId,omitempty"`
	Date          string `json:"date,omitempty"`
	SettlementDate string `json:"settlementDate,omitempty"`
	TimeZone      string `json:"timeZone,omitempty"`
}

// authorizationResponse mirrors MasterCardAuthorizationResponse. The
// mock fills ResponseCode + Stan so the consumer can simulate ISO-8583
// parsing.
type authorizationResponse struct {
	Stan                  string `json:"stan,omitempty"`
	ResponseCode          string `json:"responseCode,omitempty"`
	ProcessingCode        string `json:"processingCode,omitempty"`
	TransactionIdentifier string `json:"transactionIdentifier,omitempty"`
	FinancialNetworkCode  string `json:"financialNetworkCode,omitempty"`
	CardSecurityCodeError string `json:"cardSecurityCodeError,omitempty"`
}

// errorBlock mirrors MasterCardError. MPGS returns this alongside (not
// instead of) the payment envelope — Result=ERROR / FAILURE, and this
// block explains what went wrong at the request level.
type errorBlock struct {
	Cause          string `json:"cause,omitempty"`
	Explanation    string `json:"explanation,omitempty"`
	Field          string `json:"field,omitempty"`
	ValidationType string `json:"validationType,omitempty"`
}

// webhookNotification is the shape MPGS POSTs to the merchant's
// notificationUrl when a transaction changes state. MPGS reuses the
// same transaction envelope for webhooks — the merchant sees the full
// order + transaction + response block, plus a notification id /
// timestamp wrapper.
type webhookNotification struct {
	NotificationID   string                `json:"notificationId"`
	NotificationType string                `json:"notificationType"`
	TimeOfNotification string              `json:"timeOfNotification"`
	Result           string                `json:"result"`
	Order            orderResponse         `json:"order"`
	Response         responseBlock         `json:"response"`
	Transaction      transactionResponse   `json:"transaction"`
	Merchant         string                `json:"merchant,omitempty"`
	Version          string                `json:"version,omitempty"`
}
