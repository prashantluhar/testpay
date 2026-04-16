package postgres_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prashantluhar/testpay/internal/store"
	pgstore "github.com/prashantluhar/testpay/internal/store/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestStore(t *testing.T) *pgstore.Store {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set — skipping Postgres tests")
	}
	pool, err := pgxpool.New(context.Background(), dsn)
	require.NoError(t, err)
	require.NoError(t, pgstore.RunMigrations(pool))
	s := pgstore.New(pool)
	t.Cleanup(func() {
		// Clean up all test data
		pool.Exec(context.Background(), "DELETE FROM workspaces")
		pool.Close()
	})
	return s
}

func TestCreateAndGetWorkspace(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	w := &store.Workspace{
		ID:     uuid.NewString(),
		Slug:   "test-workspace",
		APIKey: "key_test_123",
	}
	require.NoError(t, s.CreateWorkspace(ctx, w))

	got, err := s.GetWorkspaceByAPIKey(ctx, "key_test_123")
	require.NoError(t, err)
	assert.Equal(t, w.Slug, got.Slug)
}

func TestCreateAndListScenarios(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	w := &store.Workspace{ID: uuid.NewString(), Slug: "ws2", APIKey: "key_2"}
	require.NoError(t, s.CreateWorkspace(ctx, w))

	sc := &store.Scenario{
		ID:          uuid.NewString(),
		WorkspaceID: w.ID,
		Name:        "retry-storm",
		Gateway:     "stripe",
		Steps: []store.Step{
			{Event: "charge", Outcome: "network_error"},
			{Event: "charge", Outcome: "success"},
		},
	}
	require.NoError(t, s.CreateScenario(ctx, sc))

	list, err := s.ListScenarios(ctx, w.ID)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, "retry-storm", list[0].Name)
}

func TestSessionLifecycle(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	w := &store.Workspace{ID: uuid.NewString(), Slug: "ws3", APIKey: "key_3"}
	require.NoError(t, s.CreateWorkspace(ctx, w))

	sc := &store.Scenario{ID: uuid.NewString(), WorkspaceID: w.ID, Name: "test", Gateway: "stripe", Steps: []store.Step{}}
	require.NoError(t, s.CreateScenario(ctx, sc))

	sess := &store.Session{
		ID:          uuid.NewString(),
		WorkspaceID: w.ID,
		ScenarioID:  sc.ID,
		TTLSeconds:  60,
		ExpiresAt:   time.Now().Add(60 * time.Second),
	}
	require.NoError(t, s.CreateSession(ctx, sess))

	active, err := s.GetActiveSession(ctx, w.ID)
	require.NoError(t, err)
	assert.Equal(t, sess.ID, active.ID)

	require.NoError(t, s.DeleteSession(ctx, sess.ID))
	_, err = s.GetActiveSession(ctx, w.ID)
	assert.Error(t, err)
}
