package adapters_test

import (
	"testing"

	"github.com/prashantluhar/testpay/internal/adapters"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegistry_resolveStripe(t *testing.T) {
	r := adapters.NewRegistry()
	a, err := r.Resolve("/stripe/v1/charges")
	require.NoError(t, err)
	assert.Equal(t, "stripe", a.Name())
}

func TestRegistry_resolveUnknown(t *testing.T) {
	r := adapters.NewRegistry()
	_, err := r.Resolve("/unknown/v1/charges")
	assert.Error(t, err)
}
