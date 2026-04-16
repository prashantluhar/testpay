package engine_test

import (
	"testing"
	"github.com/prashantluhar/testpay/internal/engine"
	"github.com/stretchr/testify/assert"
)

func TestFailureModeIsValid(t *testing.T) {
	assert.True(t, engine.IsValidMode("bank_decline_hard"))
	assert.True(t, engine.IsValidMode("webhook_missing"))
	assert.True(t, engine.IsValidMode("success"))
	assert.False(t, engine.IsValidMode("made_up_mode"))
}
