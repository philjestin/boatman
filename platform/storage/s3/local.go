package s3

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// LocalArtifactStore implements ArtifactStore using the local filesystem.
// This is the default fallback when S3 is not configured.
type LocalArtifactStore struct {
	baseDir string
}

// NewLocalArtifactStore creates a local filesystem artifact store.
func NewLocalArtifactStore(baseDir string) (*LocalArtifactStore, error) {
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("create artifact dir: %w", err)
	}
	return &LocalArtifactStore{baseDir: baseDir}, nil
}

func (s *LocalArtifactStore) Put(_ context.Context, key string, r io.Reader) error {
	path := filepath.Join(s.baseDir, key)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("create artifact parent dir: %w", err)
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create artifact file: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, r); err != nil {
		return fmt.Errorf("write artifact: %w", err)
	}
	return nil
}

func (s *LocalArtifactStore) Get(_ context.Context, key string) (io.ReadCloser, error) {
	path := filepath.Join(s.baseDir, key)
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("artifact not found: %s", key)
		}
		return nil, err
	}
	return f, nil
}

func (s *LocalArtifactStore) Delete(_ context.Context, key string) error {
	path := filepath.Join(s.baseDir, key)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (s *LocalArtifactStore) Exists(_ context.Context, key string) (bool, error) {
	path := filepath.Join(s.baseDir, key)
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
