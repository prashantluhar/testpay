package engine

type FailureMode string

const (
	// Generic
	ModeSuccess FailureMode = "success"

	// Bank-side
	ModeBankDeclineHard   FailureMode = "bank_decline_hard"
	ModeBankDeclineSoft   FailureMode = "bank_decline_soft"
	ModeBankServerDown    FailureMode = "bank_server_down"
	ModeBankTimeout       FailureMode = "bank_timeout"
	ModeBankInvalidCVV    FailureMode = "bank_invalid_cvv"
	ModeBankDoNotHonour   FailureMode = "bank_do_not_honour"

	// PG server-side
	ModePGServerError   FailureMode = "pg_server_error"
	ModePGTimeout       FailureMode = "pg_timeout"
	ModePGRateLimited   FailureMode = "pg_rate_limited"
	ModePGMaintenance   FailureMode = "pg_maintenance"
	ModeNetworkError    FailureMode = "network_error"

	// Webhook anomalies
	ModeWebhookMissing     FailureMode = "webhook_missing"
	ModeWebhookDelayed     FailureMode = "webhook_delayed"
	ModeWebhookDuplicate   FailureMode = "webhook_duplicate"
	ModeWebhookOutOfOrder  FailureMode = "webhook_out_of_order"
	ModeWebhookMalformed   FailureMode = "webhook_malformed"

	// Redirect / 3DS
	ModeRedirectSuccess   FailureMode = "redirect_success"
	ModeRedirectAbandoned FailureMode = "redirect_abandoned"
	ModeRedirectTimeout   FailureMode = "redirect_timeout"
	ModeRedirectFailed    FailureMode = "redirect_failed"

	// Charge anomalies
	ModeDoubleCharge    FailureMode = "double_charge"
	ModeAmountMismatch  FailureMode = "amount_mismatch"
	ModePartialSuccess  FailureMode = "partial_success"

	// Async state transitions
	ModePendingThenFailed   FailureMode = "pending_then_failed"
	ModePendingThenSuccess  FailureMode = "pending_then_success"
	ModeFailedThenSuccess   FailureMode = "failed_then_success"
	ModeSuccessThenReversed FailureMode = "success_then_reversed"
)

var validModes = map[FailureMode]bool{
	ModeSuccess: true,
	ModeBankDeclineHard: true, ModeBankDeclineSoft: true,
	ModeBankServerDown: true, ModeBankTimeout: true,
	ModeBankInvalidCVV: true, ModeBankDoNotHonour: true,
	ModePGServerError: true, ModePGTimeout: true,
	ModePGRateLimited: true, ModePGMaintenance: true,
	ModeNetworkError: true,
	ModeWebhookMissing: true, ModeWebhookDelayed: true,
	ModeWebhookDuplicate: true, ModeWebhookOutOfOrder: true,
	ModeWebhookMalformed: true,
	ModeRedirectSuccess: true, ModeRedirectAbandoned: true,
	ModeRedirectTimeout: true, ModeRedirectFailed: true,
	ModeDoubleCharge: true, ModeAmountMismatch: true,
	ModePartialSuccess: true,
	ModePendingThenFailed: true, ModePendingThenSuccess: true,
	ModeFailedThenSuccess: true, ModeSuccessThenReversed: true,
}

func IsValidMode(mode string) bool {
	return validModes[FailureMode(mode)]
}
