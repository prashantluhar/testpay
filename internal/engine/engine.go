package engine

import "github.com/prashantluhar/testpay/internal/store"

// Result is the outcome of executing one scenario step.
type Result struct {
	Mode           FailureMode
	HTTPStatus     int
	ErrorCode      string  // gateway-specific error code, e.g. "insufficient_funds"
	SkipWebhook    bool    // true for webhook_missing
	DuplicateWebhook bool  // true for webhook_duplicate
	WebhookDelayMs int
	IsPending      bool    // true for async pending states
}

// Engine executes scenario steps.
type Engine struct{}

func New() *Engine { return &Engine{} }

// Execute returns the Result for step at stepIndex.
// If stepIndex >= len(steps), the last step is used.
func (e *Engine) Execute(sc *store.Scenario, stepIndex int) (*Result, error) {
	if len(sc.Steps) == 0 {
		return successResult(sc.WebhookDelayMs), nil
	}
	idx := stepIndex
	if idx >= len(sc.Steps) {
		idx = len(sc.Steps) - 1
	}
	step := sc.Steps[idx]
	return modeToResult(FailureMode(step.Outcome), step.Code, sc.WebhookDelayMs), nil
}

func modeToResult(mode FailureMode, code string, delayMs int) *Result {
	r := &Result{Mode: mode, WebhookDelayMs: delayMs}
	switch mode {
	case ModeSuccess:
		r.HTTPStatus = 200
	case ModeBankDeclineHard, ModeBankDeclineSoft, ModeBankInvalidCVV, ModeBankDoNotHonour:
		r.HTTPStatus = 402
		r.ErrorCode = string(mode)
	case ModeBankServerDown, ModeBankTimeout, ModePGTimeout, ModeNetworkError:
		r.HTTPStatus = 503
	case ModePGServerError:
		r.HTTPStatus = 500
	case ModePGRateLimited:
		r.HTTPStatus = 429
	case ModePGMaintenance:
		r.HTTPStatus = 503
		r.ErrorCode = "maintenance"
	case ModeWebhookMissing:
		r.HTTPStatus = 200
		r.SkipWebhook = true
	case ModeWebhookDelayed:
		r.HTTPStatus = 200
	case ModeWebhookDuplicate:
		r.HTTPStatus = 200
		r.DuplicateWebhook = true
	case ModeWebhookOutOfOrder, ModeWebhookMalformed:
		r.HTTPStatus = 200
	case ModeRedirectSuccess:
		r.HTTPStatus = 200
	case ModeRedirectAbandoned, ModeRedirectTimeout, ModeRedirectFailed:
		r.HTTPStatus = 402
		r.ErrorCode = string(mode)
	case ModeDoubleCharge:
		r.HTTPStatus = 200
		r.DuplicateWebhook = true
	case ModeAmountMismatch:
		r.HTTPStatus = 200
	case ModePartialSuccess:
		r.HTTPStatus = 200
	case ModePendingThenFailed, ModePendingThenSuccess, ModeFailedThenSuccess, ModeSuccessThenReversed:
		r.HTTPStatus = 200
		r.IsPending = true
	default:
		r.HTTPStatus = 200
	}
	if code != "" {
		r.ErrorCode = code
	}
	return r
}

func successResult(delayMs int) *Result {
	return &Result{Mode: ModeSuccess, HTTPStatus: 200, WebhookDelayMs: delayMs}
}
