package sqlite_test

import (
	"testing"

	"github.com/philjestin/boatman-ecosystem/platform/storage"
	"github.com/philjestin/boatman-ecosystem/platform/storage/sqlite"
)

func TestSQLiteStore(t *testing.T) {
	store, err := sqlite.New(sqlite.WithInMemory())
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer store.Close()

	storage.RunStoreTests(t, store)
}
