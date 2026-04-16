package handlers_test

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/prashantluhar/testpay/internal/api/handlers"
	"github.com/prashantluhar/testpay/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockStore is a minimal in-memory store for handler tests.
type mockStore struct {
	store.Store
	scenarios []*store.Scenario
}

func (m *mockStore) ListScenarios(_ context.Context, _ string) ([]*store.Scenario, error) {
	return m.scenarios, nil
}
func (m *mockStore) CreateScenario(_ context.Context, s *store.Scenario) error {
	m.scenarios = append(m.scenarios, s)
	return nil
}

func TestListScenarios_returnsJSON(t *testing.T) {
	ms := &mockStore{scenarios: []*store.Scenario{
		{ID: "1", Name: "retry-storm"},
	}}
	r := chi.NewRouter()
	r.Get("/api/scenarios", handlers.ListScenarios(ms))

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest("GET", "/api/scenarios", nil))
	assert.Equal(t, 200, rec.Code)

	var resp []map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Len(t, resp, 1)
	assert.Equal(t, "retry-storm", resp[0]["name"])
}
