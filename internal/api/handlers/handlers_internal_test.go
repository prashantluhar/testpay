// Internal tests — same package, so we can reach unexported helpers.
package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prashantluhar/testpay/internal/store"
	"github.com/stretchr/testify/assert"
)

// internalFailStore returns an error from every workspace lookup so the
// WorkspaceIDFromRequest path fully exhausts its fallback and the shim
// workspaceFromCtx bubbles the same error.
type internalFailStore struct{ store.Store }

var errInternalFail = errors.New("fail")

func (internalFailStore) GetWorkspaceByAPIKey(context.Context, string) (*store.Workspace, error) {
	return nil, errInternalFail
}
func (internalFailStore) GetWorkspaceBySlug(context.Context, string) (*store.Workspace, error) {
	return nil, errInternalFail
}
func (internalFailStore) GetWorkspaceByID(context.Context, string) (*store.Workspace, error) {
	return nil, errInternalFail
}

func TestWorkspaceIDFromRequest_fallsBackToLocalID(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer any")
	got := WorkspaceIDFromRequest(req, internalFailStore{})
	assert.Equal(t, store.LocalWorkspaceID, got)
}

func TestWorkspaceFromCtx_propagatesLookupError(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	_, err := workspaceFromCtx(req, internalFailStore{})
	assert.Error(t, err)
}

func TestExtractOperation_allBranches(t *testing.T) {
	cases := map[string]string{
		"/stripe/v1/refunds":              "refund",
		"/stripe/v1/charges/ch_1/capture": "capture",
		"/stripe/v1/authorizations":       "authorize",
		"/stripe/v1/charges":              "charge",
		"/stripe/v1/payments":             "charge",
		"/stripe/v1/something-else":       "unknown",
	}
	for path, want := range cases {
		t.Run(path, func(t *testing.T) {
			assert.Equal(t, want, extractOperation(path))
		})
	}
}

func TestExtractAmountCurrency_typeVariants(t *testing.T) {
	cases := []struct {
		name        string
		body        map[string]any
		wantAmount  int64
		wantCurrency string
	}{
		{"nil body", nil, 100, "usd"},
		{"float64", map[string]any{"amount": float64(4200)}, 4200, "usd"},
		{"int", map[string]any{"amount": int(7777)}, 7777, "usd"},
		{"int64", map[string]any{"amount": int64(9999)}, 9999, "usd"},
		{"currency override", map[string]any{"amount": float64(100), "currency": "eur"}, 100, "eur"},
		{"unexpected amount type ignored", map[string]any{"amount": "string-amount"}, 100, "usd"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			a, cur := extractAmountCurrency(c.body, 100, "usd")
			assert.Equal(t, c.wantAmount, a)
			assert.Equal(t, c.wantCurrency, cur)
		})
	}
}

func TestExtractMerchantOrderID_searchPath(t *testing.T) {
	cases := []struct {
		name string
		body map[string]any
		want string
	}{
		{"nil", nil, ""},
		{"none", map[string]any{"foo": "bar"}, ""},
		{"order_id", map[string]any{"order_id": "ord-1"}, "ord-1"},
		{"merchant_order_id", map[string]any{"merchant_order_id": "ord-2"}, "ord-2"},
		{"merchantOrderId", map[string]any{"merchantOrderId": "ord-3"}, "ord-3"},
		{"reference", map[string]any{"reference": "ord-4"}, "ord-4"},
		{"nested metadata order_id", map[string]any{"metadata": map[string]any{"order_id": "ord-5"}}, "ord-5"},
		{"nested notes reference", map[string]any{"notes": map[string]any{"reference": "ord-6"}}, "ord-6"},
		{"empty string skipped", map[string]any{"order_id": ""}, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.want, extractMerchantOrderID(c.body))
		})
	}
}

func TestHeaderToMap_collapsesFirstValue(t *testing.T) {
	h := http.Header{}
	h.Add("X-Multi", "first")
	h.Add("X-Multi", "second")
	h.Set("X-Single", "only")
	m := headerToMap(h)
	assert.Equal(t, "first", m["X-Multi"])
	assert.Equal(t, "only", m["X-Single"])
}

func TestHeaderToMap_empty(t *testing.T) {
	assert.Empty(t, headerToMap(http.Header{}))
}

func TestEnsureLocalWorkspaceContext_isIdentity(t *testing.T) {
	ctx := context.WithValue(context.Background(), testKey{}, "v")
	assert.Equal(t, ctx, ensureLocalWorkspaceContext(ctx))
}

type testKey struct{}
