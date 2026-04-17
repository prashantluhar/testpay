// DTO shapes for TillPayment's transaction response and webhook notifications.
// Derived from the production reference library (test-dto-/tillpayment.go).
// Scoped to the fields this mock emits; extend when a failure mode needs more
// surface area. TillPayment is a card-acquiring platform; its wire format
// centres around a `uuid` (gateway txn id), a `purchaseId`, and a
// success/result/transactionStatus triple.
package tillpay

// TillPayment's constant vocabulary, lifted verbatim from the reference DTO
// package so the mock emits the same strings merchants see in production.
const (
	statusSuccess            = "SUCCESS"
	statusFailed             = "FAILED"
	statusDebit              = "DEBIT"
	statusChargeback         = "CHARGEBACK"
	statusChargebackReversed = "CHARGEBACK-REVERSED"
	statusRefund             = "REFUND"
	returnTypeFinished       = "FINISHED"
	defaultErrorCode         = "5000"
	paymentSystemError       = "PAYMENT_SYSTEM_ERROR"
	defaultUserMessage       = "Sorry, we couldn't complete the transaction due to a technical error. Please contact support for further assistance."
	paymentMethodCreditCard  = "CREDITCARD"
)

// transactionResponse mirrors TillPaymentTransactionResponse from the
// production DTO. Emitted by the debit / tokenisation endpoints on both
// success and failure. When `Success == false`, `Errors` is populated and
// `ErrorMessage`/`ErrorCode` carry the top-level reason.
type transactionResponse struct {
	Success       bool            `json:"success"`
	ErrorMessage  string          `json:"errorMessage,omitempty"`
	ErrorCode     int             `json:"errorCode,omitempty"`
	Message       string          `json:"message,omitempty"`
	UUID          string          `json:"uuid"`
	PurchaseID    string          `json:"purchaseId"`
	ReturnType    string          `json:"returnType"`
	PaymentMethod string          `json:"paymentMethod"`
	RedirectURL   string          `json:"redirectUrl,omitempty"`
	Errors        []paymentError  `json:"errors,omitempty"`
}

// paymentError is one entry in the top-level `errors` array. TillPayment
// duplicates the primary error fields here for per-error granularity.
type paymentError struct {
	ErrorMessage   string `json:"errorMessage,omitempty"`
	ErrorCode      int    `json:"errorCode,omitempty"`
	Message        string `json:"message,omitempty"`
	Code           string `json:"code,omitempty"`
	AdapterMessage string `json:"adapterMessage,omitempty"`
	AdapterCode    string `json:"adapterCode,omitempty"`
}

// customer is the subset of TillPaymentCustomer we echo in webhooks. Full
// reference type has more billing-address fields; keeping it lean.
type customer struct {
	Identification string `json:"identification,omitempty"`
	FirstName      string `json:"firstName,omitempty"`
	LastName       string `json:"lastName,omitempty"`
	BillingCountry string `json:"billingCountry,omitempty"`
	Email          string `json:"email,omitempty"`
	EmailVerified  bool   `json:"emailVerified,omitempty"`
	IPAddress      string `json:"ipAddress,omitempty"`
}

// returnData is the card snapshot TillPayment echoes back on webhooks —
// never raw PAN; first-six/last-four + fingerprint + bin metadata.
type returnData struct {
	Type           string `json:"_TYPE,omitempty"`
	CardHolder     string `json:"cardHolder,omitempty"`
	ExpiryMonth    string `json:"expiryMonth,omitempty"`
	ExpiryYear     string `json:"expiryYear,omitempty"`
	BinDigits      string `json:"binDigits,omitempty"`
	FirstSixDigits string `json:"firstSixDigits,omitempty"`
	LastFourDigits string `json:"lastFourDigits,omitempty"`
	Fingerprint    string `json:"fingerprint,omitempty"`
	ThreeDSecure   string `json:"threeDSecure,omitempty"`
	BinBrand       string `json:"binBrand,omitempty"`
	BinBank        string `json:"binBank,omitempty"`
	BinCountry     string `json:"binCountry,omitempty"`
}

// webhookNotification mirrors TillPaymentWebhookNotification. Key invariant
// the merchant relies on:
//   - `success` (bool) and `result` ("OK"/"NOK") must agree
//   - `transactionStatus` drives the state machine (DEBIT, REFUND,
//     CHARGEBACK, CHARGEBACK-REVERSED, FAILED, SUCCESS)
//   - `amount` is a string (TillPayment quirk — not numeric)
type webhookNotification struct {
	Result                string         `json:"result"`
	Success               bool           `json:"success"`
	TransactionStatus     string         `json:"transactionStatus"`
	Message               string         `json:"message,omitempty"`
	Code                  any            `json:"code,omitempty"`
	AdapterMessage        string         `json:"adapterMessage,omitempty"`
	AdapterCode           string         `json:"adapterCode,omitempty"`
	UUID                  string         `json:"uuid"`
	MerchantTransactionID string         `json:"merchantTransactionId"`
	PurchaseID            string         `json:"purchaseId"`
	TransactionType       string         `json:"transactionType"`
	PaymentMethod         string         `json:"paymentMethod"`
	Amount                string         `json:"amount"`
	Currency              string         `json:"currency"`
	Customer              customer       `json:"customer"`
	ReturnData            returnData     `json:"returnData,omitempty"`
	Errors                []paymentError `json:"errors,omitempty"`
}
