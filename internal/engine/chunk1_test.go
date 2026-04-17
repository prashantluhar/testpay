package engine_test

import (
	"testing"

	"github.com/prashantluhar/testpay/internal/engine"
	"github.com/prashantluhar/testpay/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecute_emptyStepsReturnsSuccess(t *testing.T) {
	e := engine.New()
	res, err := e.Execute(&store.Scenario{WebhookDelayMs: 200}, 0)
	require.NoError(t, err)
	assert.Equal(t, 200, res.HTTPStatus)
	assert.Equal(t, engine.ModeSuccess, res.Mode)
	assert.Equal(t, 200, res.WebhookDelayMs)
}

func TestExecute_stepIndexBeyondLastClampsToLast(t *testing.T) {
	e := engine.New()
	sc := &store.Scenario{
		Steps: []store.Step{
			{Outcome: string(engine.ModeSuccess)},
			{Outcome: string(engine.ModePGServerError)},
		},
	}
	res, err := e.Execute(sc, 99)
	require.NoError(t, err)
	assert.Equal(t, 500, res.HTTPStatus)
}

func TestExecute_customErrorCodeOverridesMode(t *testing.T) {
	e := engine.New()
	sc := &store.Scenario{
		Steps: []store.Step{{
			Outcome: string(engine.ModeBankDeclineHard),
			Code:    "custom_decline",
		}},
	}
	res, err := e.Execute(sc, 0)
	require.NoError(t, err)
	assert.Equal(t, "custom_decline", res.ErrorCode)
}

// TestModeToResult_everyKnownMode walks the full failure-mode table so any
// future switch-case regression is caught by coverage + an assertion.
func TestModeToResult_everyKnownMode(t *testing.T) {
	e := engine.New()
	cases := []struct {
		mode          engine.FailureMode
		wantStatus    int
		wantErrCode   string
		wantSkipHook  bool
		wantDupHook   bool
		wantIsPending bool
	}{
		{engine.ModeSuccess, 200, "", false, false, false},
		{engine.ModeBankDeclineHard, 402, string(engine.ModeBankDeclineHard), false, false, false},
		{engine.ModeBankDeclineSoft, 402, string(engine.ModeBankDeclineSoft), false, false, false},
		{engine.ModeBankInvalidCVV, 402, string(engine.ModeBankInvalidCVV), false, false, false},
		{engine.ModeBankDoNotHonour, 402, string(engine.ModeBankDoNotHonour), false, false, false},
		{engine.ModeBankServerDown, 503, "", false, false, false},
		{engine.ModeBankTimeout, 503, "", false, false, false},
		{engine.ModePGTimeout, 503, "", false, false, false},
		{engine.ModeNetworkError, 503, "", false, false, false},
		{engine.ModePGServerError, 500, "", false, false, false},
		{engine.ModePGRateLimited, 429, "", false, false, false},
		{engine.ModePGMaintenance, 503, "maintenance", false, false, false},
		{engine.ModeWebhookMissing, 200, "", true, false, false},
		{engine.ModeWebhookDelayed, 200, "", false, false, false},
		{engine.ModeWebhookDuplicate, 200, "", false, true, false},
		{engine.ModeWebhookOutOfOrder, 200, "", false, false, false},
		{engine.ModeWebhookMalformed, 200, "", false, false, false},
		{engine.ModeRedirectSuccess, 200, "", false, false, false},
		{engine.ModeRedirectAbandoned, 402, string(engine.ModeRedirectAbandoned), false, false, false},
		{engine.ModeRedirectTimeout, 402, string(engine.ModeRedirectTimeout), false, false, false},
		{engine.ModeRedirectFailed, 402, string(engine.ModeRedirectFailed), false, false, false},
		{engine.ModeDoubleCharge, 200, "", false, true, false},
		{engine.ModeAmountMismatch, 200, "", false, false, false},
		{engine.ModePartialSuccess, 200, "", false, false, false},
		{engine.ModePendingThenFailed, 200, "", false, false, true},
		{engine.ModePendingThenSuccess, 200, "", false, false, true},
		{engine.ModeFailedThenSuccess, 200, "", false, false, true},
		{engine.ModeSuccessThenReversed, 200, "", false, false, true},
	}
	for _, c := range cases {
		t.Run(string(c.mode), func(t *testing.T) {
			sc := &store.Scenario{Steps: []store.Step{{Outcome: string(c.mode)}}}
			res, err := e.Execute(sc, 0)
			require.NoError(t, err)
			assert.Equal(t, c.wantStatus, res.HTTPStatus)
			assert.Equal(t, c.wantErrCode, res.ErrorCode)
			assert.Equal(t, c.wantSkipHook, res.SkipWebhook)
			assert.Equal(t, c.wantDupHook, res.DuplicateWebhook)
			assert.Equal(t, c.wantIsPending, res.IsPending)
		})
	}
}

func TestModeToResult_unknownModeFallsBackTo200(t *testing.T) {
	e := engine.New()
	sc := &store.Scenario{Steps: []store.Step{{Outcome: "totally_made_up"}}}
	res, err := e.Execute(sc, 0)
	require.NoError(t, err)
	assert.Equal(t, 200, res.HTTPStatus)
}
