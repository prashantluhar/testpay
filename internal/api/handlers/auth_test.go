package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/prashantluhar/testpay/internal/api/handlers"
	"github.com/prashantluhar/testpay/internal/store"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

// authMockStore satisfies store.Store for auth tests (only the methods we need).
type authMockStore struct {
	store.Store
	users      map[string]*store.User
	hashes     map[string]string // email -> hash
	workspaces map[string]*store.Workspace
}

func newAuthMock() *authMockStore {
	return &authMockStore{
		users:      map[string]*store.User{},
		hashes:     map[string]string{},
		workspaces: map[string]*store.Workspace{},
	}
}

func (m *authMockStore) CreateWorkspace(_ context.Context, w *store.Workspace) error {
	m.workspaces[w.ID] = w
	return nil
}
func (m *authMockStore) CreateUser(_ context.Context, u *store.User, hash string) error {
	m.users[u.ID] = u
	m.hashes[u.Email] = hash
	return nil
}
func (m *authMockStore) GetUserByEmail(_ context.Context, email string) (*store.User, string, error) {
	for _, u := range m.users {
		if u.Email == email {
			return u, m.hashes[email], nil
		}
	}
	return nil, "", assert.AnError
}
func (m *authMockStore) GetUserByID(_ context.Context, id string) (*store.User, error) {
	if u, ok := m.users[id]; ok {
		return u, nil
	}
	return nil, assert.AnError
}
func (m *authMockStore) UpdateUserLastLogin(_ context.Context, _ string, _ time.Time) error {
	return nil
}

func (m *authMockStore) GetWorkspaceBySlug(_ context.Context, slug string) (*store.Workspace, error) {
	for _, w := range m.workspaces {
		if w.Slug == slug {
			return w, nil
		}
	}
	return nil, assert.AnError
}

func (m *authMockStore) GetWorkspaceByID(_ context.Context, id string) (*store.Workspace, error) {
	if w, ok := m.workspaces[id]; ok {
		return w, nil
	}
	return nil, assert.AnError
}

func TestSignup_createsUserAndWorkspace(t *testing.T) {
	ms := newAuthMock()
	h := handlers.Signup(ms, "test-secret", "local")

	body, _ := json.Marshal(map[string]string{"email": "alice@example.com", "password": "password123"})
	req := httptest.NewRequest("POST", "/api/auth/signup", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	assert.Equal(t, 201, rec.Code)
	assert.Len(t, ms.users, 1)
	assert.Len(t, ms.workspaces, 1)
	cookies := rec.Result().Cookies()
	var found bool
	for _, c := range cookies {
		if c.Name == "testpay_session" {
			found = true
			assert.True(t, c.HttpOnly)
		}
	}
	assert.True(t, found, "session cookie not set")
}

func TestLogin_validCredentials(t *testing.T) {
	ms := newAuthMock()
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	ms.users["u1"] = &store.User{ID: "u1", WorkspaceID: "w1", Email: "alice@example.com", Role: "owner"}
	ms.hashes["alice@example.com"] = string(hash)
	ms.workspaces["w1"] = &store.Workspace{ID: "w1", Slug: "alice", APIKey: "key"}

	body, _ := json.Marshal(map[string]string{"email": "alice@example.com", "password": "password123"})
	req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	handlers.Login(ms, "test-secret", "local").ServeHTTP(rec, req)
	assert.Equal(t, 200, rec.Code)
}

func TestLogin_wrongPassword(t *testing.T) {
	ms := newAuthMock()
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	ms.users["u1"] = &store.User{ID: "u1", WorkspaceID: "w1", Email: "alice@example.com", Role: "owner"}
	ms.hashes["alice@example.com"] = string(hash)

	body, _ := json.Marshal(map[string]string{"email": "alice@example.com", "password": "wrong"})
	req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	handlers.Login(ms, "test-secret", "local").ServeHTTP(rec, req)
	assert.Equal(t, 401, rec.Code)
}

