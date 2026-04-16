package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/prashantluhar/testpay/internal/api/handlers"
	"github.com/prashantluhar/testpay/internal/adapters"
	"github.com/prashantluhar/testpay/internal/engine"
	"github.com/stretchr/testify/assert"
)

func TestMockHandler_successScenario(t *testing.T) {
	eng := engine.New()
	reg := adapters.NewRegistry()

	h := handlers.NewMock(eng, reg, nil, nil) // nil store + dispatcher for unit test
	srv := httptest.NewServer(h)
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/stripe/v1/charges",
		"application/json", strings.NewReader(`{"amount":5000,"currency":"usd"}`))
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}
