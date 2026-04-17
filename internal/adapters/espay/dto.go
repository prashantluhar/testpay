// DTO shapes for ESPay's (Indonesia) payment API surface.
// Derived from the production reference library (test-dto-/payments.go,
// structs: ESPayInquireResponse, ESPayPaymentStatusResponse,
// ESPayWebhookResponse, ESPayRefundResponse, ...).
//
// ESPay uses RSA-signed JSON payloads with a strict error_code / error_message
// envelope — "0000" for success, arbitrary ESPxx codes otherwise. We mock the
// signature field with a fixed placeholder; callers that verify signatures
// should short-circuit on the mock prefix.
package espay

// inquireResponse mirrors ESPayInquireResponse — returned when the merchant
// creates a payment inquiry. Fields here are the subset this mock populates
// on the happy path.
type inquireResponse struct {
	RqUuid            string          `json:"rq_uuid"`
	RsDateTime        string          `json:"rs_datetime"`
	ErrorCode         string          `json:"error_code"`
	ErrorMessage      string          `json:"error_message"`
	Signature         string          `json:"signature"`
	OrderID           string          `json:"order_id"`
	Amount            string          `json:"amount"`
	CCY               string          `json:"ccy"`
	Description       string          `json:"description"`
	TrxDate           string          `json:"trx_date"`
	InstallmentPeriod string          `json:"installment_period,omitempty"`
	CustomerDetails   customerDetails `json:"customer_details"`
}

// customerDetails mirrors ESPayCustomerDetails — minimal PII envelope on the
// inquire response. Populated with placeholder values for the mock.
type customerDetails struct {
	FirstName   string `json:"firstname"`
	LastName    string `json:"lastname"`
	PhoneNumber string `json:"phone_number"`
	Email       string `json:"email"`
}

// errorResponse is the shape ESPay returns on a failed inquire/status call.
// Same envelope as success minus the rich fields — just error_code + message
// + signature + (optional) order_id.
type errorResponse struct {
	RqUuid       string `json:"rq_uuid"`
	RsDateTime   string `json:"rs_datetime"`
	ErrorCode    string `json:"error_code"`
	ErrorMessage string `json:"error_message"`
	Signature    string `json:"signature"`
	OrderID      string `json:"order_id,omitempty"`
}

// webhookResponse mirrors ESPayWebhookResponse — ESPay posts this to the
// merchant's callback URL once the customer completes the VA / wallet
// payment. reconcile_id is the bank-side settlement reference.
type webhookResponse struct {
	RqUuid            string `json:"rq_uuid"`
	RsDateTime        string `json:"rs_datetime"`
	ErrorCode         string `json:"error_code"`
	ErrorMessage      string `json:"error_message"`
	Signature         string `json:"signature"`
	OrderID           string `json:"order_id"`
	ReconcileID       string `json:"reconcile_id"`
	ReconcileDateTime string `json:"reconcile_datetime"`
}

// paymentStatusResponse mirrors ESPayPaymentStatusResponse — the rich status
// shape used when the merchant polls the gateway. We include the fields most
// likely to drive consumer state machines (tx_status, tx_reason, amount).
type paymentStatusResponse struct {
	RqUuid       string `json:"rq_uuid"`
	RsDateTime   string `json:"rs_datetime"`
	ErrorCode    string `json:"error_code"`
	ErrorMessage string `json:"error_message"`
	CommCode     string `json:"comm_code"`
	TxID         string `json:"tx_id"`
	OrderID      string `json:"order_id"`
	CCYID        string `json:"ccy_id"`
	Amount       string `json:"amount"`
	TxStatus     string `json:"tx_status"`
	TxReason     string `json:"tx_reason"`
	TxDate       string `json:"tx_date"`
	BankName     string `json:"bank_name,omitempty"`
	ProductName  string `json:"product_name,omitempty"`
}
