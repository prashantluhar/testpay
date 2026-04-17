package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/prashantluhar/testpay/internal/adapters"
	"github.com/prashantluhar/testpay/internal/api/handlers"
	"github.com/prashantluhar/testpay/internal/api/middleware"
	"github.com/prashantluhar/testpay/internal/engine"
	"github.com/prashantluhar/testpay/internal/store"
	"github.com/prashantluhar/testpay/internal/webhook"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ── Auth edge cases ─────────────────────────────────────────────────────────

func TestSignup_invalidJSON400(t *testing.T) {
	h := handlers.Signup(newFakeStore(), "secret", "local")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("POST", "/", strings.NewReader("{not json")))
	assert.Equal(t, 400, rec.Code)
}

func TestSignup_shortPassword400(t *testing.T) {
	h := handlers.Signup(newFakeStore(), "secret", "local")
	body, _ := json.Marshal(map[string]string{"email": "a@b.com", "password": "short"})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("POST", "/", bytes.NewReader(body)))
	assert.Equal(t, 400, rec.Code)
}

func TestSignup_invalidEmail400(t *testing.T) {
	h := handlers.Signup(newFakeStore(), "secret", "local")
	body, _ := json.Marshal(map[string]string{"email": "no-at-sign", "password": "password123"})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("POST", "/", bytes.NewReader(body)))
	assert.Equal(t, 400, rec.Code)
}

func TestSignup_duplicateEmail409(t *testing.T) {
	fs := newFakeStore()
	// First signup succeeds.
	h := handlers.Signup(fs, "secret", "local")
	first, _ := json.Marshal(map[string]string{"email": "dup@x.com", "password": "password123"})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("POST", "/", bytes.NewReader(first)))
	require.Equal(t, 201, rec.Code)
	// Second signup with same email.
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("POST", "/", bytes.NewReader(first)))
	assert.Equal(t, 409, rec.Code)
}

func TestLogin_invalidJSON400(t *testing.T) {
	h := handlers.Login(newFakeStore(), "secret", "local")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("POST", "/", strings.NewReader("not json")))
	assert.Equal(t, 400, rec.Code)
}

func TestLogout_writesExpiredCookie(t *testing.T) {
	h := handlers.Logout("local")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("POST", "/", nil))
	assert.Equal(t, 204, rec.Code)
	cookies := rec.Result().Cookies()
	require.NotEmpty(t, cookies)
	assert.Equal(t, -1, cookies[0].MaxAge)
}

func TestMe_localModeReturnsLocalWorkspace(t *testing.T) {
	fs := newFakeStore()
	fs.seedLocal()
	h := handlers.Me(fs)
	req := httptest.NewRequest("GET", "/", nil)
	// Inject local-mode context: workspace id set, no user.
	ctx := withWorkspace(req.Context(), store.LocalWorkspaceID)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req.WithContext(ctx))
	assert.Equal(t, 200, rec.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Nil(t, resp["user"])
	require.NotNil(t, resp["workspace"])
}

func TestMe_401WithoutUserOrLocalWorkspace(t *testing.T) {
	fs := newFakeStore()
	h := handlers.Me(fs)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	assert.Equal(t, 401, rec.Code)
}

func TestMe_200ForAuthenticatedUser(t *testing.T) {
	fs := newFakeStore()
	ws := &store.Workspace{ID: "ws-1", Slug: "alice", APIKey: "k"}
	fs.workspaces[ws.ID] = ws
	u := &store.User{ID: "u-1", WorkspaceID: ws.ID, Email: "a@b.com", Role: "owner"}
	fs.users[u.ID] = u

	h := handlers.Me(fs)
	req := httptest.NewRequest("GET", "/", nil)
	ctx := injectSession(req.Context(), u.ID, ws.ID)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req.WithContext(ctx))
	assert.Equal(t, 200, rec.Code)
}

// ── Scenarios ───────────────────────────────────────────────────────────────

func TestListScenarios_storeErrorReturns500(t *testing.T) {
	fs := newFakeStore()
	fs.seedLocal()
	fs.FailListScenarios = errors.New("db down")
	rec := httptest.NewRecorder()
	handlers.ListScenarios(fs).ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	assert.Equal(t, 500, rec.Code)
}

func TestCreateScenario_success(t *testing.T) {
	fs := newFakeStore()
	fs.seedLocal()
	body, _ := json.Marshal(map[string]any{"name": "retry-storm"})
	rec := httptest.NewRecorder()
	handlers.CreateScenario(fs).ServeHTTP(rec, httptest.NewRequest("POST", "/", bytes.NewReader(body)))
	assert.Equal(t, 201, rec.Code)
	assert.Len(t, fs.scenarios, 1)
}

func TestCreateScenario_invalidBody400(t *testing.T) {
	fs := newFakeStore()
	fs.seedLocal()
	rec := httptest.NewRecorder()
	handlers.CreateScenario(fs).ServeHTTP(rec, httptest.NewRequest("POST", "/", strings.NewReader("garbage")))
	assert.Equal(t, 400, rec.Code)
}

func TestCreateScenario_storeError500(t *testing.T) {
	fs := newFakeStore()
	fs.seedLocal()
	fs.FailCreateScenario = errors.New("db down")
	body, _ := json.Marshal(map[string]any{"name": "x"})
	rec := httptest.NewRecorder()
	handlers.CreateScenario(fs).ServeHTTP(rec, httptest.NewRequest("POST", "/", bytes.NewReader(body)))
	assert.Equal(t, 500, rec.Code)
}

func TestGetScenario_404WhenMissing(t *testing.T) {
	fs := newFakeStore()
	r := chi.NewRouter()
	r.Get("/{id}", handlers.GetScenario(fs))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest("GET", "/nope", nil))
	assert.Equal(t, 404, rec.Code)
}

func TestGetScenario_200WhenPresent(t *testing.T) {
	fs := newFakeStore()
	fs.scenarios["s-1"] = &store.Scenario{ID: "s-1", Name: "x"}
	r := chi.NewRouter()
	r.Get("/{id}", handlers.GetScenario(fs))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest("GET", "/s-1", nil))
	assert.Equal(t, 200, rec.Code)
}

func TestUpdateScenario_success(t *testing.T) {
	fs := newFakeStore()
	fs.scenarios["s-1"] = &store.Scenario{ID: "s-1", Name: "old"}
	r := chi.NewRouter()
	r.Put("/{id}", handlers.UpdateScenario(fs))
	body, _ := json.Marshal(map[string]any{"name": "new"})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest("PUT", "/s-1", bytes.NewReader(body)))
	assert.Equal(t, 200, rec.Code)
	assert.Equal(t, "new", fs.scenarios["s-1"].Name)
}

func TestUpdateScenario_invalidBody400(t *testing.T) {
	fs := newFakeStore()
	r := chi.NewRouter()
	r.Put("/{id}", handlers.UpdateScenario(fs))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest("PUT", "/s-1", strings.NewReader("garbage")))
	assert.Equal(t, 400, rec.Code)
}

func TestUpdateScenario_storeError500(t *testing.T) {
	fs := newFakeStore()
	fs.FailUpdateScenario = errors.New("fail")
	r := chi.NewRouter()
	r.Put("/{id}", handlers.UpdateScenario(fs))
	body, _ := json.Marshal(map[string]any{"name": "x"})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest("PUT", "/s-1", bytes.NewReader(body)))
	assert.Equal(t, 500, rec.Code)
}

func TestDeleteScenario_204(t *testing.T) {
	fs := newFakeStore()
	fs.scenarios["s-1"] = &store.Scenario{ID: "s-1"}
	r := chi.NewRouter()
	r.Delete("/{id}", handlers.DeleteScenario(fs))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest("DELETE", "/s-1", nil))
	assert.Equal(t, 204, rec.Code)
	assert.Empty(t, fs.scenarios)
}

func TestDeleteScenario_storeError500(t *testing.T) {
	fs := newFakeStore()
	fs.FailDeleteScenario = errors.New("fail")
	r := chi.NewRouter()
	r.Delete("/{id}", handlers.DeleteScenario(fs))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest("DELETE", "/s-1", nil))
	assert.Equal(t, 500, rec.Code)
}

func TestRunScenario_success(t *testing.T) {
	fs := newFakeStore()
	fs.scenarios["s-1"] = &store.Scenario{
		ID:    "s-1",
		Steps: []store.Step{{Outcome: string(engine.ModeSuccess)}},
	}
	r := chi.NewRouter()
	r.Post("/{id}/run", handlers.RunScenario(fs, engine.New(), adapters.NewRegistry(), nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest("POST", "/s-1/run", nil))
	assert.Equal(t, 200, rec.Code)
	assert.Len(t, fs.scenarioRuns, 1)
}

func TestRunScenario_notFound404(t *testing.T) {
	fs := newFakeStore()
	r := chi.NewRouter()
	r.Post("/{id}/run", handlers.RunScenario(fs, engine.New(), adapters.NewRegistry(), nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest("POST", "/nope/run", nil))
	assert.Equal(t, 404, rec.Code)
}

// ── Logs ────────────────────────────────────────────────────────────────────

func TestListLogs_success(t *testing.T) {
	fs := newFakeStore()
	fs.seedLocal()
	fs.requestLogs["l1"] = &store.RequestLog{ID: "l1", WorkspaceID: store.LocalWorkspaceID, Path: "/stripe/v1/charges"}
	rec := httptest.NewRecorder()
	handlers.ListLogs(fs).ServeHTTP(rec, httptest.NewRequest("GET", "/?limit=10&offset=0", nil))
	assert.Equal(t, 200, rec.Code)
	var resp []map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Len(t, resp, 1)
}

func TestListLogs_storeError500(t *testing.T) {
	fs := newFakeStore()
	fs.seedLocal()
	fs.FailListLogs = errors.New("db down")
	rec := httptest.NewRecorder()
	handlers.ListLogs(fs).ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	assert.Equal(t, 500, rec.Code)
}

func TestGetLog_404(t *testing.T) {
	fs := newFakeStore()
	r := chi.NewRouter()
	r.Get("/{id}", handlers.GetLog(fs))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest("GET", "/nope", nil))
	assert.Equal(t, 404, rec.Code)
}

func TestGetLog_200WithWebhook(t *testing.T) {
	fs := newFakeStore()
	fs.requestLogs["l1"] = &store.RequestLog{ID: "l1", Path: "/stripe/v1/charges"}
	fs.hookByReqID["l1"] = &store.WebhookLog{ID: "w1", RequestLogID: "l1"}
	r := chi.NewRouter()
	r.Get("/{id}", handlers.GetLog(fs))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest("GET", "/l1", nil))
	assert.Equal(t, 200, rec.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.NotNil(t, resp["request"])
	assert.NotNil(t, resp["webhook"])
}

func TestReplayLog_success(t *testing.T) {
	fs := newFakeStore()
	fs.requestLogs["l1"] = &store.RequestLog{ID: "l1", Path: "/stripe/v1/charges"}
	r := chi.NewRouter()
	r.Post("/{id}/replay", handlers.ReplayLog(fs, engine.New(), adapters.NewRegistry(), nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest("POST", "/l1/replay", nil))
	assert.Equal(t, 200, rec.Code)
}

func TestReplayLog_logNotFound404(t *testing.T) {
	fs := newFakeStore()
	r := chi.NewRouter()
	r.Post("/{id}/replay", handlers.ReplayLog(fs, engine.New(), adapters.NewRegistry(), nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest("POST", "/nope/replay", nil))
	assert.Equal(t, 404, rec.Code)
}

func TestReplayLog_unknownGateway400(t *testing.T) {
	fs := newFakeStore()
	fs.requestLogs["l1"] = &store.RequestLog{ID: "l1", Path: "/totally-made-up/v1/charges"}
	r := chi.NewRouter()
	r.Post("/{id}/replay", handlers.ReplayLog(fs, engine.New(), adapters.NewRegistry(), nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest("POST", "/l1/replay", nil))
	assert.Equal(t, 400, rec.Code)
}

// ── Sessions ────────────────────────────────────────────────────────────────

func TestCreateSession_success(t *testing.T) {
	fs := newFakeStore()
	fs.seedLocal()
	body, _ := json.Marshal(map[string]any{"scenario_id": "s-1", "ttl_seconds": 60})
	rec := httptest.NewRecorder()
	handlers.CreateSession(fs).ServeHTTP(rec, httptest.NewRequest("POST", "/", bytes.NewReader(body)))
	assert.Equal(t, 201, rec.Code)
	assert.Len(t, fs.sessions, 1)
}

func TestCreateSession_missingScenarioID400(t *testing.T) {
	fs := newFakeStore()
	rec := httptest.NewRecorder()
	handlers.CreateSession(fs).ServeHTTP(rec, httptest.NewRequest("POST", "/", strings.NewReader(`{}`)))
	assert.Equal(t, 400, rec.Code)
}

func TestCreateSession_storeError500(t *testing.T) {
	fs := newFakeStore()
	fs.seedLocal()
	fs.FailCreateSession = errors.New("fail")
	body, _ := json.Marshal(map[string]any{"scenario_id": "s-1"})
	rec := httptest.NewRecorder()
	handlers.CreateSession(fs).ServeHTTP(rec, httptest.NewRequest("POST", "/", bytes.NewReader(body)))
	assert.Equal(t, 500, rec.Code)
}

func TestCreateSession_zeroTTLDefaultsToHour(t *testing.T) {
	fs := newFakeStore()
	fs.seedLocal()
	body, _ := json.Marshal(map[string]any{"scenario_id": "s-1", "ttl_seconds": 0})
	rec := httptest.NewRecorder()
	handlers.CreateSession(fs).ServeHTTP(rec, httptest.NewRequest("POST", "/", bytes.NewReader(body)))
	require.Equal(t, 201, rec.Code)
	for _, s := range fs.sessions {
		assert.Equal(t, 3600, s.TTLSeconds)
	}
}

func TestDeleteSession_204(t *testing.T) {
	fs := newFakeStore()
	fs.sessions["s-1"] = &store.Session{ID: "s-1"}
	r := chi.NewRouter()
	r.Delete("/{id}", handlers.DeleteSession(fs))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest("DELETE", "/s-1", nil))
	assert.Equal(t, 204, rec.Code)
}

func TestDeleteSession_204EvenOnStoreError(t *testing.T) {
	// Handler always 204s — it logs store errors but doesn't propagate.
	fs := newFakeStore()
	fs.FailDeleteSession = errors.New("fail")
	r := chi.NewRouter()
	r.Delete("/{id}", handlers.DeleteSession(fs))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest("DELETE", "/s-1", nil))
	assert.Equal(t, 204, rec.Code)
}

// ── Webhooks ────────────────────────────────────────────────────────────────

func TestTestWebhook_success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(200) }))
	defer srv.Close()
	body, _ := json.Marshal(map[string]any{"target_url": srv.URL, "payload": map[string]any{"event": "x"}})
	rec := httptest.NewRecorder()
	handlers.TestWebhook(newDispatcher()).ServeHTTP(rec, httptest.NewRequest("POST", "/", bytes.NewReader(body)))
	assert.Equal(t, 200, rec.Code)
}

func TestTestWebhook_missingURL400(t *testing.T) {
	rec := httptest.NewRecorder()
	handlers.TestWebhook(newDispatcher()).ServeHTTP(rec, httptest.NewRequest("POST", "/", strings.NewReader(`{}`)))
	assert.Equal(t, 400, rec.Code)
}

func TestListWebhooks_success(t *testing.T) {
	fs := newFakeStore()
	fs.seedLocal()
	fs.webhookLogs["w1"] = &store.WebhookLog{ID: "w1"}
	rec := httptest.NewRecorder()
	handlers.ListWebhooks(fs).ServeHTTP(rec, httptest.NewRequest("GET", "/?limit=10", nil))
	assert.Equal(t, 200, rec.Code)
}

func TestListWebhooks_storeError500(t *testing.T) {
	fs := newFakeStore()
	fs.seedLocal()
	fs.FailListWebhooks = errors.New("fail")
	rec := httptest.NewRecorder()
	handlers.ListWebhooks(fs).ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	assert.Equal(t, 500, rec.Code)
}

func TestGetWebhook_200(t *testing.T) {
	fs := newFakeStore()
	fs.webhookLogs["w1"] = &store.WebhookLog{ID: "w1", TargetURL: "https://example.com"}
	r := chi.NewRouter()
	r.Get("/{id}", handlers.GetWebhook(fs))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest("GET", "/w1", nil))
	assert.Equal(t, 200, rec.Code)
}

func TestGetWebhook_404(t *testing.T) {
	fs := newFakeStore()
	r := chi.NewRouter()
	r.Get("/{id}", handlers.GetWebhook(fs))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest("GET", "/nope", nil))
	assert.Equal(t, 404, rec.Code)
}

func TestGetWebhookStatus_200(t *testing.T) {
	fs := newFakeStore()
	fs.hookByReqID["req-1"] = &store.WebhookLog{ID: "w1", DeliveryStatus: "delivered"}
	r := chi.NewRouter()
	r.Get("/{id}/status", handlers.GetWebhookStatus(fs))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest("GET", "/req-1/status", nil))
	assert.Equal(t, 200, rec.Code)
}

func TestGetWebhookStatus_404(t *testing.T) {
	fs := newFakeStore()
	r := chi.NewRouter()
	r.Get("/{id}/status", handlers.GetWebhookStatus(fs))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest("GET", "/nope/status", nil))
	assert.Equal(t, 404, rec.Code)
}

// ── Workspace ──────────────────────────────────────────────────────────────

func TestGetWorkspace_localFallback(t *testing.T) {
	fs := newFakeStore()
	fs.seedLocal()
	rec := httptest.NewRecorder()
	handlers.GetWorkspace(fs).ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	assert.Equal(t, 200, rec.Code)
}

func TestGetWorkspace_404WhenNoLocal(t *testing.T) {
	fs := newFakeStore() // no seedLocal
	rec := httptest.NewRecorder()
	handlers.GetWorkspace(fs).ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	assert.Equal(t, 404, rec.Code)
}

func TestGetWorkspace_viaBearerToken(t *testing.T) {
	fs := newFakeStore()
	fs.workspaces["ws-1"] = &store.Workspace{ID: "ws-1", Slug: "alice", APIKey: "key-alice"}
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer key-alice")
	rec := httptest.NewRecorder()
	handlers.GetWorkspace(fs).ServeHTTP(rec, req)
	assert.Equal(t, 200, rec.Code)
}

func TestGetWorkspace_viaSessionContext(t *testing.T) {
	fs := newFakeStore()
	fs.workspaces["ws-1"] = &store.Workspace{ID: "ws-1", Slug: "alice"}
	req := httptest.NewRequest("GET", "/", nil)
	req = req.WithContext(withWorkspace(req.Context(), "ws-1"))
	rec := httptest.NewRecorder()
	handlers.GetWorkspace(fs).ServeHTTP(rec, req)
	assert.Equal(t, 200, rec.Code)
}

func TestUpdateWorkspace_persistsWebhookURLs(t *testing.T) {
	fs := newFakeStore()
	fs.seedLocal()
	body, _ := json.Marshal(map[string]any{
		"webhook_urls": map[string]string{"stripe": "https://example.com/hook", "razorpay": ""},
	})
	rec := httptest.NewRecorder()
	handlers.UpdateWorkspace(fs).ServeHTTP(rec, httptest.NewRequest("PUT", "/", bytes.NewReader(body)))
	assert.Equal(t, 200, rec.Code)
	ws := fs.workspaces[store.LocalWorkspaceID]
	require.NotNil(t, ws)
	assert.Equal(t, "https://example.com/hook", ws.WebhookURLs["stripe"])
	assert.NotContains(t, ws.WebhookURLs, "razorpay", "empty values should be dropped")
}

func TestUpdateWorkspace_invalidBody400(t *testing.T) {
	fs := newFakeStore()
	fs.seedLocal()
	rec := httptest.NewRecorder()
	handlers.UpdateWorkspace(fs).ServeHTTP(rec, httptest.NewRequest("PUT", "/", strings.NewReader("garbage")))
	assert.Equal(t, 400, rec.Code)
}

func TestUpdateWorkspace_404WhenWorkspaceMissing(t *testing.T) {
	fs := newFakeStore() // no seed
	body, _ := json.Marshal(map[string]any{"webhook_urls": map[string]string{}})
	rec := httptest.NewRecorder()
	handlers.UpdateWorkspace(fs).ServeHTTP(rec, httptest.NewRequest("PUT", "/", bytes.NewReader(body)))
	assert.Equal(t, 404, rec.Code)
}

func TestUpdateWorkspace_storeError500(t *testing.T) {
	fs := newFakeStore()
	fs.seedLocal()
	fs.FailUpdateWorkspace = errors.New("fail")
	body, _ := json.Marshal(map[string]any{"webhook_urls": map[string]string{"x": "y"}})
	rec := httptest.NewRecorder()
	handlers.UpdateWorkspace(fs).ServeHTTP(rec, httptest.NewRequest("PUT", "/", bytes.NewReader(body)))
	assert.Equal(t, 500, rec.Code)
}

// ── Gateways / misc ─────────────────────────────────────────────────────────

func TestListGateways_returnsKnownSet(t *testing.T) {
	reg := adapters.NewRegistry()
	rec := httptest.NewRecorder()
	handlers.ListGateways(reg).ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	assert.Equal(t, 200, rec.Code)
	var names []string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &names))
	assert.Contains(t, names, "stripe")
}

// ── Mock handler — exercises headerToMap, extractAmountCurrency, and the
//    rest of the ServeHTTP tail with a fully-wired fake store + dispatcher.

func TestNewMockWithMode_hostedModeRequiresBearer(t *testing.T) {
	fs := newFakeStore()
	h := handlers.NewMockWithMode(engine.New(), adapters.NewRegistry(), fs, newDispatcher(), "hosted")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("POST", "/stripe/v1/charges",
		bytes.NewReader([]byte(`{"amount":5000}`))))
	assert.Equal(t, 401, rec.Code)
}

func TestMockHandler_localFallbackWorkspacePersistsLog(t *testing.T) {
	fs := newFakeStore()
	fs.seedLocal()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(200)
	}))
	defer srv.Close()
	// Seed webhook URL on the local workspace so the goroutine path runs end-to-end.
	fs.workspaces[store.LocalWorkspaceID].WebhookURLs = map[string]string{"stripe": srv.URL}

	h := handlers.NewMock(engine.New(), adapters.NewRegistry(), fs, newDispatcher())
	body := bytes.NewReader([]byte(`{"amount":7777,"currency":"eur","metadata":{"order_id":"ord-42"}}`))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/stripe/v1/charges", body)
	h.ServeHTTP(rec, req)
	assert.Equal(t, 200, rec.Code)

	// Request log is persisted in the goroutine — wait for it.
	require.Eventually(t, func() bool {
		fs.mu.Lock()
		defer fs.mu.Unlock()
		return len(fs.requestLogs) > 0
	}, 2*time.Second, 10*time.Millisecond)
}

func TestMockHandler_unknownGatewayReturns404(t *testing.T) {
	fs := newFakeStore()
	fs.seedLocal()
	h := handlers.NewMock(engine.New(), adapters.NewRegistry(), fs, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("POST", "/madeup-gateway/v1/foo", strings.NewReader("{}")))
	assert.Equal(t, 404, rec.Code)
}

// ── helpers ─────────────────────────────────────────────────────────────────

func newDispatcher() *webhook.Dispatcher {
	// Short base delay so tests don't dawdle on the retry schedule.
	return webhook.NewDispatcher(1, 1)
}

// withWorkspace injects the middleware session-context key via the real
// Session middleware — the key type is unexported, so we ride a JWT through
// the middleware to populate it.
func withWorkspace(ctx context.Context, wsID string) context.Context {
	return injectSession(ctx, "", wsID)
}

func injectSession(parent context.Context, userID, wsID string) context.Context {
	const secret = "test-secret"
	tok, err := middleware.IssueToken(secret, userID, wsID, time.Hour)
	if err != nil {
		panic(err)
	}
	req := httptest.NewRequest("GET", "/", nil).WithContext(parent)
	req.AddCookie(&http.Cookie{Name: middleware.SessionCookieName, Value: tok})
	var enriched context.Context
	handler := middleware.Session(secret, "hosted")(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		enriched = r.Context()
	}))
	handler.ServeHTTP(httptest.NewRecorder(), req)
	return enriched
}
