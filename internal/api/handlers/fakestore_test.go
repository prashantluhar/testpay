package handlers_test

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/prashantluhar/testpay/internal/store"
)

// fakeStore is an in-memory store.Store implementation covering every method
// handler tests need. Keeps the test files themselves focused on behavior.
type fakeStore struct {
	mu sync.Mutex

	workspaces map[string]*store.Workspace // by ID
	users      map[string]*store.User
	userHash   map[string]string // email -> bcrypt hash

	scenarios       map[string]*store.Scenario
	scenarioRuns    map[string]*store.ScenarioRun
	defaultScenario map[string]*store.Scenario // by workspace ID

	sessions      map[string]*store.Session
	activeSession map[string]*store.Session // by workspace ID

	requestLogs map[string]*store.RequestLog
	webhookLogs map[string]*store.WebhookLog
	hookByReqID map[string]*store.WebhookLog // request_log_id -> webhook_log

	// error injection — set a non-nil error to force the matching store
	// method to return it on next call.
	FailCreateWorkspace error
	FailListScenarios   error
	FailCreateScenario  error
	FailGetScenario     error
	FailUpdateScenario  error
	FailDeleteScenario  error
	FailCreateSession   error
	FailDeleteSession   error
	FailListLogs        error
	FailGetLog          error
	FailListWebhooks    error
	FailGetWebhook      error
	FailGetHookByReq    error
	FailUpdateWorkspace error
}

func newFakeStore() *fakeStore {
	return &fakeStore{
		workspaces:      map[string]*store.Workspace{},
		users:           map[string]*store.User{},
		userHash:        map[string]string{},
		scenarios:       map[string]*store.Scenario{},
		scenarioRuns:    map[string]*store.ScenarioRun{},
		defaultScenario: map[string]*store.Scenario{},
		sessions:        map[string]*store.Session{},
		activeSession:   map[string]*store.Session{},
		requestLogs:     map[string]*store.RequestLog{},
		webhookLogs:     map[string]*store.WebhookLog{},
		hookByReqID:     map[string]*store.WebhookLog{},
	}
}

// seedLocal seeds a LocalWorkspace so WorkspaceFromRequest fallbacks succeed.
func (f *fakeStore) seedLocal() *store.Workspace {
	ws := &store.Workspace{ID: store.LocalWorkspaceID, Slug: "local", APIKey: "local-key"}
	f.workspaces[ws.ID] = ws
	return ws
}

// Workspace --------------------------------------------------------------

func (f *fakeStore) CreateWorkspace(_ context.Context, w *store.Workspace) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.FailCreateWorkspace != nil {
		return f.FailCreateWorkspace
	}
	f.workspaces[w.ID] = w
	return nil
}
func (f *fakeStore) GetWorkspaceByAPIKey(_ context.Context, key string) (*store.Workspace, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, w := range f.workspaces {
		if w.APIKey == key {
			return w, nil
		}
	}
	return nil, fmt.Errorf("not found")
}
func (f *fakeStore) GetWorkspaceBySlug(_ context.Context, slug string) (*store.Workspace, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, w := range f.workspaces {
		if w.Slug == slug {
			return w, nil
		}
	}
	return nil, fmt.Errorf("not found")
}
func (f *fakeStore) GetWorkspaceByID(_ context.Context, id string) (*store.Workspace, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if w, ok := f.workspaces[id]; ok {
		return w, nil
	}
	return nil, fmt.Errorf("not found")
}
func (f *fakeStore) UpdateWorkspace(_ context.Context, w *store.Workspace) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.FailUpdateWorkspace != nil {
		return f.FailUpdateWorkspace
	}
	f.workspaces[w.ID] = w
	return nil
}

// Users ------------------------------------------------------------------

func (f *fakeStore) CreateUser(_ context.Context, u *store.User, hash string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.users[u.ID] = u
	f.userHash[u.Email] = hash
	return nil
}
func (f *fakeStore) GetUserByEmail(_ context.Context, email string) (*store.User, string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, u := range f.users {
		if u.Email == email {
			return u, f.userHash[email], nil
		}
	}
	return nil, "", fmt.Errorf("not found")
}
func (f *fakeStore) GetUserByID(_ context.Context, id string) (*store.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if u, ok := f.users[id]; ok {
		return u, nil
	}
	return nil, fmt.Errorf("not found")
}
func (f *fakeStore) UpdateUserLastLogin(_ context.Context, _ string, _ time.Time) error { return nil }

// Scenarios --------------------------------------------------------------

func (f *fakeStore) ListScenarios(_ context.Context, workspaceID string) ([]*store.Scenario, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.FailListScenarios != nil {
		return nil, f.FailListScenarios
	}
	out := []*store.Scenario{}
	for _, s := range f.scenarios {
		if s.WorkspaceID == workspaceID || workspaceID == "" {
			out = append(out, s)
		}
	}
	return out, nil
}
func (f *fakeStore) CreateScenario(_ context.Context, s *store.Scenario) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.FailCreateScenario != nil {
		return f.FailCreateScenario
	}
	f.scenarios[s.ID] = s
	return nil
}
func (f *fakeStore) GetScenario(_ context.Context, id string) (*store.Scenario, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.FailGetScenario != nil {
		return nil, f.FailGetScenario
	}
	if s, ok := f.scenarios[id]; ok {
		return s, nil
	}
	return nil, fmt.Errorf("not found")
}
func (f *fakeStore) UpdateScenario(_ context.Context, s *store.Scenario) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.FailUpdateScenario != nil {
		return f.FailUpdateScenario
	}
	f.scenarios[s.ID] = s
	return nil
}
func (f *fakeStore) DeleteScenario(_ context.Context, id string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.FailDeleteScenario != nil {
		return f.FailDeleteScenario
	}
	delete(f.scenarios, id)
	return nil
}
func (f *fakeStore) GetDefaultScenario(_ context.Context, workspaceID string) (*store.Scenario, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if s, ok := f.defaultScenario[workspaceID]; ok {
		return s, nil
	}
	return nil, fmt.Errorf("not found")
}
func (f *fakeStore) CreateScenarioRun(_ context.Context, r *store.ScenarioRun) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.scenarioRuns[r.ID] = r
	return nil
}
func (f *fakeStore) UpdateScenarioRun(_ context.Context, r *store.ScenarioRun) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.scenarioRuns[r.ID] = r
	return nil
}

// Sessions ---------------------------------------------------------------

func (f *fakeStore) CreateSession(_ context.Context, s *store.Session) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.FailCreateSession != nil {
		return f.FailCreateSession
	}
	f.sessions[s.ID] = s
	f.activeSession[s.WorkspaceID] = s
	return nil
}
func (f *fakeStore) GetActiveSession(_ context.Context, workspaceID string) (*store.Session, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if s, ok := f.activeSession[workspaceID]; ok {
		return s, nil
	}
	return nil, fmt.Errorf("not found")
}
func (f *fakeStore) DeleteSession(_ context.Context, id string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.FailDeleteSession != nil {
		return f.FailDeleteSession
	}
	delete(f.sessions, id)
	return nil
}
func (f *fakeStore) BumpSessionCallIndex(_ context.Context, sessionID string) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	sess, ok := f.sessions[sessionID]
	if !ok {
		return 0, fmt.Errorf("session not found")
	}
	pre := sess.CallIndex
	sess.CallIndex++
	return pre, nil
}
func (f *fakeStore) CreateFeedback(_ context.Context, _ *store.Feedback) error { return nil }

// Request logs -----------------------------------------------------------

func (f *fakeStore) CreateRequestLog(_ context.Context, l *store.RequestLog) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.requestLogs[l.ID] = l
	return nil
}
func (f *fakeStore) ListRequestLogs(_ context.Context, workspaceID string, limit, offset int) ([]*store.RequestLog, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.FailListLogs != nil {
		return nil, f.FailListLogs
	}
	all := []*store.RequestLog{}
	for _, l := range f.requestLogs {
		if l.WorkspaceID == workspaceID || workspaceID == "" {
			all = append(all, l)
		}
	}
	start := offset
	if start > len(all) {
		start = len(all)
	}
	end := start + limit
	if end > len(all) {
		end = len(all)
	}
	return all[start:end], nil
}
func (f *fakeStore) GetRequestLog(_ context.Context, id string) (*store.RequestLog, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.FailGetLog != nil {
		return nil, f.FailGetLog
	}
	if l, ok := f.requestLogs[id]; ok {
		return l, nil
	}
	return nil, fmt.Errorf("not found")
}

// Webhook logs -----------------------------------------------------------

func (f *fakeStore) CreateWebhookLog(_ context.Context, l *store.WebhookLog) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.webhookLogs[l.ID] = l
	if l.RequestLogID != "" {
		f.hookByReqID[l.RequestLogID] = l
	}
	return nil
}
func (f *fakeStore) UpdateWebhookLog(_ context.Context, l *store.WebhookLog) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.webhookLogs[l.ID] = l
	return nil
}
func (f *fakeStore) GetWebhookLogByRequestID(_ context.Context, requestLogID string) (*store.WebhookLog, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.FailGetHookByReq != nil {
		return nil, f.FailGetHookByReq
	}
	if l, ok := f.hookByReqID[requestLogID]; ok {
		return l, nil
	}
	return nil, fmt.Errorf("not found")
}
func (f *fakeStore) GetWebhookLog(_ context.Context, id string) (*store.WebhookLog, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.FailGetWebhook != nil {
		return nil, f.FailGetWebhook
	}
	if l, ok := f.webhookLogs[id]; ok {
		return l, nil
	}
	return nil, fmt.Errorf("not found")
}
func (f *fakeStore) ListWebhookLogs(_ context.Context, _ string, _, _ int) ([]*store.WebhookLog, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.FailListWebhooks != nil {
		return nil, f.FailListWebhooks
	}
	out := []*store.WebhookLog{}
	for _, l := range f.webhookLogs {
		out = append(out, l)
	}
	return out, nil
}
