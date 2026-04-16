package engine_test

import (
	"testing"

	"github.com/prashantluhar/testpay/internal/engine"
	"github.com/prashantluhar/testpay/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func scenario(steps ...store.Step) *store.Scenario {
	return &store.Scenario{
		ID:             "scn_test",
		Name:           "test",
		Gateway:        "stripe",
		Steps:          steps,
		WebhookDelayMs: 0,
	}
}

func TestExecute_success(t *testing.T) {
	e := engine.New()
	sc := scenario(store.Step{Event: "charge", Outcome: "success"})
	result, err := e.Execute(sc, 0)
	require.NoError(t, err)
	assert.Equal(t, engine.ModeSuccess, result.Mode)
	assert.Equal(t, 200, result.HTTPStatus)
	assert.False(t, result.SkipWebhook)
}

func TestExecute_bankDeclineHard(t *testing.T) {
	e := engine.New()
	sc := scenario(store.Step{Event: "charge", Outcome: "bank_decline_hard"})
	result, err := e.Execute(sc, 0)
	require.NoError(t, err)
	assert.Equal(t, engine.ModeBankDeclineHard, result.Mode)
	assert.Equal(t, 402, result.HTTPStatus)
}

func TestExecute_webhookMissing(t *testing.T) {
	e := engine.New()
	sc := scenario(store.Step{Event: "charge", Outcome: "webhook_missing"})
	result, err := e.Execute(sc, 0)
	require.NoError(t, err)
	assert.True(t, result.SkipWebhook)
}

func TestExecute_multiStep_sequencing(t *testing.T) {
	e := engine.New()
	sc := scenario(
		store.Step{Event: "charge", Outcome: "network_error"},
		store.Step{Event: "charge", Outcome: "success"},
	)
	r0, _ := e.Execute(sc, 0)
	r1, _ := e.Execute(sc, 1)
	assert.Equal(t, engine.ModeNetworkError, r0.Mode)
	assert.Equal(t, engine.ModeSuccess, r1.Mode)
}

func TestExecute_outOfBoundsStepWraps(t *testing.T) {
	e := engine.New()
	sc := scenario(store.Step{Event: "charge", Outcome: "success"})
	// step index beyond len(steps) wraps to last step
	result, err := e.Execute(sc, 99)
	require.NoError(t, err)
	assert.Equal(t, engine.ModeSuccess, result.Mode)
}
