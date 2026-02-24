// Package sqlite implements the storage.Store interface backed by SQLite.
package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"

	"github.com/philjestin/boatman-ecosystem/platform/storage"
)

// Options configures the SQLite store.
type Options struct {
	// Path is the database file path. Use ":memory:" for in-memory databases.
	Path string
}

// WithInMemory returns Options for an in-memory database (useful for tests).
func WithInMemory() Options {
	return Options{Path: ":memory:"}
}

// SQLiteStore implements storage.Store using SQLite.
type SQLiteStore struct {
	db       *sql.DB
	runs     *runStore
	memory   *memoryStore
	costs    *costStore
	policies *policyStore
	events   *eventStore
}

// New creates a new SQLiteStore with the given options.
func New(opts Options) (*SQLiteStore, error) {
	if opts.Path == "" {
		opts.Path = ":memory:"
	}

	dsn := opts.Path
	if dsn != ":memory:" {
		dsn = fmt.Sprintf("file:%s?_journal_mode=WAL&_busy_timeout=5000&_synchronous=NORMAL", opts.Path)
	}

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	// SQLite concurrency settings
	db.SetMaxOpenConns(1) // SQLite handles one writer at a time

	s := &SQLiteStore{db: db}
	s.runs = &runStore{db: db}
	s.memory = &memoryStore{db: db}
	s.costs = &costStore{db: db}
	s.policies = &policyStore{db: db}
	s.events = &eventStore{db: db}

	return s, nil
}

func (s *SQLiteStore) Runs() storage.RunStore       { return s.runs }
func (s *SQLiteStore) Memory() storage.MemoryStore   { return s.memory }
func (s *SQLiteStore) Costs() storage.CostStore      { return s.costs }
func (s *SQLiteStore) Policies() storage.PolicyStore { return s.policies }
func (s *SQLiteStore) Events() storage.EventStore    { return s.events }

// Migrate runs all DDL migrations.
func (s *SQLiteStore) Migrate(ctx context.Context) error {
	return runMigrations(ctx, s.db)
}

// Close closes the underlying database connection.
func (s *SQLiteStore) Close() error {
	return s.db.Close()
}
