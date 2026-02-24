package s3_test

import (
	"bytes"
	"context"
	"io"
	"os"
	"testing"

	"github.com/philjestin/boatman-ecosystem/platform/storage/s3"
)

func TestLocalArtifactStore(t *testing.T) {
	dir, err := os.MkdirTemp("", "artifacts-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	store, err := s3.NewLocalArtifactStore(dir)
	if err != nil {
		t.Fatalf("NewLocalArtifactStore: %v", err)
	}

	ctx := context.Background()
	key := "org1/team1/repo1/run-1/diff"
	data := []byte("this is a diff")

	// Put
	if err := store.Put(ctx, key, bytes.NewReader(data)); err != nil {
		t.Fatalf("Put: %v", err)
	}

	// Exists
	exists, err := store.Exists(ctx, key)
	if err != nil {
		t.Fatalf("Exists: %v", err)
	}
	if !exists {
		t.Error("expected artifact to exist")
	}

	// Get
	rc, err := store.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	got, err := io.ReadAll(rc)
	rc.Close()
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if string(got) != string(data) {
		t.Errorf("expected %q, got %q", data, got)
	}

	// Delete
	if err := store.Delete(ctx, key); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	exists, _ = store.Exists(ctx, key)
	if exists {
		t.Error("expected artifact to not exist after delete")
	}

	// Get non-existent
	_, err = store.Get(ctx, "nonexistent/key")
	if err == nil {
		t.Error("expected error for non-existent key")
	}
}
