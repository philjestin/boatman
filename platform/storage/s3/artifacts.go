// Package s3 provides artifact storage for large blobs (diffs, outputs).
// It defines the ArtifactStore interface and provides a local filesystem fallback.
package s3

import (
	"context"
	"io"
)

// ArtifactStore persists large binary artifacts (diffs, logs, outputs)
// separate from the relational store.
type ArtifactStore interface {
	// Put stores data under the given key. The key follows the pattern:
	// "{org}/{team}/{repo}/{run_id}/{artifact_type}"
	Put(ctx context.Context, key string, r io.Reader) error

	// Get retrieves an artifact by key. Returns io.ErrUnexpectedEOF if not found.
	Get(ctx context.Context, key string) (io.ReadCloser, error)

	// Delete removes an artifact.
	Delete(ctx context.Context, key string) error

	// Exists checks whether an artifact exists.
	Exists(ctx context.Context, key string) (bool, error)
}
