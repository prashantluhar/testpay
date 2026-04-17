// DTO shapes for ECPay (綠界 / Green World) payment API.
//
// Sourcing note: the production reference DTO (test-dto-/ecpay.go) only
// defined the channel constants and the payment-method enum, so the
// response/webhook shapes below are modelled on ECPay's public integration
// documentation (AioCheckOut v5 + ReturnURL callback format). Key contract
// points preserved from the real gateway:
//   - `RtnCode` is the primary success flag: 1 = success, anything else
//     is a failure. ECPay returns `RtnCode` as an *int* in the JSON
//     response but as a *string* form field in the webhook (x-www-form-
//     urlencoded). Our mock emits JSON for both but keeps the numeric
//     semantics intact.
//   - `MerchantTradeNo` is the merchant's correlation ID (echoed back).
//   - `TradeNo` is ECPay's gateway-side transaction ID.
//   - `TradeAmt` is the gross amount as integer TWD (ECPay uses no minor
//     units — TWD is whole-dollar).
//   - `PaymentType` identifies the tender (Credit_CreditCard, ATM_TAISHIN,
//     CVS_CVS, BARCODE_BARCODE, etc.).
//   - `CheckMacValue` is the MD5/SHA256 signature covering the whole
//     payload, sorted alphabetically. Our mock emits a placeholder.
package ecpay

// ECPay's payment-channel vocabulary, lifted from the production reference
// DTO so merchants see identical strings in both environments.
const (
	channel         = "ECPAY"
	methodCredit    = "ECPAY_CREDIT_CARD"
	methodATM       = "ECPAY_ATM_CARD"
	methodCVS       = "ECPAY_CVS"
	methodBarcode   = "ECPAY_BARCODE"
	paymentCredit   = "Credit_CreditCard"
	paymentATM      = "ATM_TAISHIN"
	paymentCVS      = "CVS_CVS"
	paymentBarcode  = "BARCODE_BARCODE"
)

// ECPay's RtnCode vocabulary. 1 is the universal success code; everything
// else is a failure whose `RtnMsg` explains the reason. These three are the
// most common codes merchants handle in production.
const (
	rtnCodeSuccess   = 1
	rtnCodeFailed    = 10100248 // generic issuer decline
	rtnCodePending   = 10100073 // async ATM/CVS pending
)

// aioResponse mirrors ECPay's AioCheckOut v5 JSON response — the body the
// merchant receives after posting a payment order. On success, `TradeNo` +
// `TradeAmt` identify the transaction; on failure, `RtnMsg` carries the
// decline reason.
type aioResponse struct {
	MerchantID        string `json:"MerchantID"`
	MerchantTradeNo   string `json:"MerchantTradeNo"`
	TradeNo           string `json:"TradeNo"`
	TradeAmt          int64  `json:"TradeAmt"`
	RtnCode           int    `json:"RtnCode"`
	RtnMsg            string `json:"RtnMsg"`
	PaymentType       string `json:"PaymentType"`
	PaymentTypeChargeFee string `json:"PaymentTypeChargeFee,omitempty"`
	TradeDate         string `json:"TradeDate"`
	CheckMacValue     string `json:"CheckMacValue"`
}

// errorResponse is ECPay's error envelope for validation / auth failures at
// the API layer (distinct from a declined charge). Fires when the merchant
// request itself is malformed — missing CheckMacValue, unknown MerchantID,
// etc. ECPay returns HTTP 4xx with this body.
type errorResponse struct {
	MerchantID      string `json:"MerchantID"`
	MerchantTradeNo string `json:"MerchantTradeNo"`
	RtnCode         int    `json:"RtnCode"`
	RtnMsg          string `json:"RtnMsg"`
	ErrorType       string `json:"ErrorType"`
	CheckMacValue   string `json:"CheckMacValue"`
}

// webhookCallback is ECPay's ReturnURL / PaymentInfoURL notification body.
// In production this arrives as `application/x-www-form-urlencoded`; we
// serialise as JSON (the dispatcher normalises downstream). The critical
// invariants are:
//   - `RtnCode` drives the merchant's state machine (1 = success)
//   - `SimulatePaid` is "1" when the transaction is a test-mode simulation,
//     "0" for real money — merchants MUST refuse to fulfil orders where
//     SimulatePaid=1 in their production environment.
//   - `PaymentDate` uses ECPay's `YYYY/MM/DD HH:MM:SS` TW-local format.
type webhookCallback struct {
	MerchantID           string `json:"MerchantID"`
	MerchantTradeNo      string `json:"MerchantTradeNo"`
	StoreID              string `json:"StoreID,omitempty"`
	RtnCode              int    `json:"RtnCode"`
	RtnMsg               string `json:"RtnMsg"`
	TradeNo              string `json:"TradeNo"`
	TradeAmt             int64  `json:"TradeAmt"`
	PaymentDate          string `json:"PaymentDate"`
	PaymentType          string `json:"PaymentType"`
	PaymentTypeChargeFee string `json:"PaymentTypeChargeFee,omitempty"`
	TradeDate            string `json:"TradeDate"`
	SimulatePaid         string `json:"SimulatePaid"`
	CustomField1         string `json:"CustomField1,omitempty"`
	CustomField2         string `json:"CustomField2,omitempty"`
	CustomField3         string `json:"CustomField3,omitempty"`
	CustomField4         string `json:"CustomField4,omitempty"`
	CheckMacValue        string `json:"CheckMacValue"`
}
