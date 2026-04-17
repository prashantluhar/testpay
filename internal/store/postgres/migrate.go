package postgres

import (
	"embed"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	pgmigrate "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func RunMigrations(pool *pgxpool.Pool) error {
	// Two sources of held connections need to be released before this function
	// returns, or pool.Close() will block forever on puddle's WaitGroup:
	//   1. stdlib.OpenDBFromPool wraps the pool in a *sql.DB that keeps idle
	//      connections acquired from puddle's perspective. db.Close() releases
	//      them.
	//   2. pgmigrate.WithInstance acquires a dedicated *sql.Conn to hold the
	//      PostgreSQL advisory lock for the migration session. Only m.Close()
	//      releases it — db.Close() doesn't reach it.
	db := stdlib.OpenDBFromPool(pool)
	defer db.Close()

	driver, err := pgmigrate.WithInstance(db, &pgmigrate.Config{})
	if err != nil {
		return fmt.Errorf("creating migrate driver: %w", err)
	}

	src, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("loading migration files: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", src, "postgres", driver)
	if err != nil {
		return fmt.Errorf("creating migrator: %w", err)
	}
	defer func() { _, _ = m.Close() }()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("running migrations: %w", err)
	}

	return nil
}
